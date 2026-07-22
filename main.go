package main

import (
	"github.com/evolbioinf/never/util"

	"github.com/evolbioinf/clio"

	"flag"
	"os"

	"errors"
	"io/fs"
	"slices"

	"github.com/evolbioinf/neighbors/tdb"

	"bytes"

	neverv1 "github.com/evolbioinf/never/api/v1/never"
	apiv2 "github.com/evolbioinf/never/api/v2"
	docsv2 "github.com/evolbioinf/never/docs/v2"
	"net/http"
	"strconv"

	"time"

	"golang.org/x/time/rate"

	"sync"

	"net"
	"strings"

	"fmt"
	"log"

	"context"
	"os/signal"
	"syscall"
)

type CustomResponseWriter struct {
	http.ResponseWriter
	code,
	size int
}

func (w *CustomResponseWriter) Write(b []byte) (int, error) {
	a, err := w.ResponseWriter.Write(b)
	w.size += a
	return a, err
}

func (w *CustomResponseWriter) WriteHeader(statusCode int) {
	w.code = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

type SyncMap struct {
	ma map[string]*rate.Limiter
	mu sync.Mutex
}

func main() {
	certificate,
		dbDirPath,
		dateFilePath,
		privateKey,
		host,
		port,
		noRateLimit := ioHandling()

	dbs := make(map[string]apiv2.Database)

	entries, err := os.ReadDir(dbDirPath)
	if err != nil {
		log.Fatal("error while trying to read database directory", err.Error())
	}
	var fileInfos []fs.FileInfo
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			log.Fatal("error while trying to read info", err.Error())
		}
		fileInfos = append(fileInfos, info)
	}
	if len(fileInfos) == 0 {
		log.Fatal(errors.New("No databases in provided directory."))
	}
	slices.SortFunc(fileInfos, func(a, b fs.FileInfo) int {
		return b.ModTime().Compare(a.ModTime())
	})

	for _, inf := range fileInfos {
		n := inf.Name()
		path := dbDirPath + "/" + n
		f, err := os.Open(path)
		if err != nil {
			continue
		}
		fileHeader := []byte{0x53, 0x51, 0x4c, 0x69, 0x74, 0x65, 0x20, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x20, 0x33, 0x00}
		h := make([]byte, 16)
		f.Read(h)
		if !bytes.Equal(h, fileHeader) {
			continue
		}

		db, err := tdb.OpenTaxonomyDBcheck(path)
		if err != nil {
			log.Fatal(fmt.Sprintf("error while opening database %s: ", path), err.Error())
		}
		defer db.Close()
		dbs[n] = apiv2.Database{Path: path, Db: db}
		if len(dbs) == 1 {
			dbs["latest"] = apiv2.Database{Path: path, Db: db}
		}

	}

	docsPref := "/docs"
	apiPref := "/api"
	docsV1Pref := docsPref + apiPref + "/v1"
	docsV2Pref := docsPref + apiPref + "/v2"
	isLocal := host == "http://localhost"

	neverv1.RegisterRoutes(apiPref+"/v1", docsV1Pref, dbs["latest"].Path, dateFilePath)
	apiv2.RegisterRoutes(apiPref+"/v2", host+":"+strconv.Itoa(port), dbs)
	docsv2.RegisterRoutes(docsV2Pref, isLocal, port)

	vitaxFiles := http.FileServer(http.Dir("vitax"))
	http.Handle("/vitax/", http.StripPrefix("/vitax/", vitaxFiles))

	http.HandleFunc("/{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, docsV2Pref, http.StatusSeeOther)
	})

	util.SetupLog()
	defer util.StopLogging()
	middlewareLogger := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bef := time.Now().UnixMilli()
			cw := &CustomResponseWriter{ResponseWriter: w, code: 200, size: 0}
			next.ServeHTTP(cw, r)

			af := time.Now().UnixMilli()
			if r.URL.Path == docsV2Pref ||
				r.URL.Path == docsV1Pref ||
				!strings.Contains(r.URL.Path, docsPref) {
				ip, _, err := net.SplitHostPort(r.RemoteAddr)
				if err != nil {
					util.LogWarningDef(
						util.WarningEntry{
							Warning: fmt.Sprintf("Couldn't parse remote address %s during %s request", r.RemoteAddr, r.URL.Path),
						})
				}

				util.LogInfoDef(util.InfoEntry{
					RequestIp:     ip,
					RequestUrl:    r.URL.String(),
					RequestMethod: r.Method,
					ResponseCode:  cw.code,
					ResponseSize:  fmt.Sprintf("%db", cw.size),
					ResponseTime:  fmt.Sprintf("%dms", int(af-bef)),
					Description:   "middleware request log entry",
				})

			}

		})
	}

	generalLimiter := rate.NewLimiter(rate.Limit(10), 100)
	userLimiters := SyncMap{ma: make(map[string]*rate.Limiter)}

	middlewareLimiter := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			allowed := noRateLimit || !strings.HasPrefix(r.URL.Path, apiPref)
			if !allowed {
				if generalLimiter.Allow() {
					ip, _, err := net.SplitHostPort(r.RemoteAddr)
					if err != nil {
						fmt.Printf("Couldn't parse remote address %s\n", r.RemoteAddr)
					}

					userLimiters.mu.Lock()
					lim, ok := userLimiters.ma[ip]
					if !ok {
						lim = rate.NewLimiter(rate.Limit(10), 25)
						userLimiters.ma[ip] = lim
					}

					allowed = lim.Allow()
					userLimiters.mu.Unlock()

				}
			}

			if !allowed {
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte("Exceeded rate limit\n"))
			} else {
				next.ServeHTTP(w, r)
			}

		})
	}

	handlerChain := middlewareLimiter(middlewareLogger(http.DefaultServeMux))
	var addr string
	if isLocal {
		addr = fmt.Sprintf(":%d", port)
	} else {
		addr = fmt.Sprintf("%s:%d", host, port)
	}

	s := http.Server{
		Addr:    addr,
		Handler: handlerChain,
	}

	go func() {
		fmt.Printf("Starting server at %s:%d ...\n", host, port)
		if isLocal {
			err = s.ListenAndServe()
		} else {
			err = s.ListenAndServeTLS(certificate, privateKey)
		}
	}()

	term, c := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer c()
	<-term.Done()

	fmt.Println("...Stopping server")
	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()
	err = s.Shutdown(ctx)

	if err != nil {
		util.LogInfoDef(util.InfoEntry{Description: "Server shutdown with error: " + err.Error()})
	} else {
		util.LogInfoDef(util.InfoEntry{Description: "Server shutdown gracefully"})
	}
	fmt.Println("Shutdown complete")

}

func ioHandling() (string, string, string, string, string, int, bool) {
	util.PrepLog("never")

	clio.Usage(
		"-o 10.254.1.21 -p 443",
		"This is the webserver never. It hosts the neighbors' REST API versions 1 and 2, "+
			"as well as their respective documentation. "+
			"New packages may be added in a similar fashion as seen "+
			"in the main function.",
		"Starts the webserver at specified address with given port.")

	cFlag := flag.String("c", "certificates/cert.pem", "certificate")
	dFlag := flag.String("d", "databases", "path to database dir from execution position")
	uFlag := flag.String("u", "updated.txt", "path to dateFile from execution position")
	kFlag := flag.String("k", "certificates/private_key.pem", "private key")
	oFlag := flag.String("o", "http://localhost", "host address")
	pFlag := flag.Int("p", 8080, "port")
	rFlag := flag.Bool("no-rate-limit", false, "Turn of rate limiting")
	vFlag := flag.Bool("v", false, "print progam info")

	flag.Parse()

	if *vFlag {
		util.PrintInfo()
		os.Exit(0)
	}

	return *cFlag, *dFlag, *uFlag, *kFlag, *oFlag, *pFlag, *rFlag

}
