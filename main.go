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

	"path/filepath"

	neverv1 "github.com/evolbioinf/never/api/v1/never"
	apiv2 "github.com/evolbioinf/never/api/v2"
	docsv2 "github.com/evolbioinf/never/docs/v2"
	"net/http"
	"strconv"

	"time"

	_ "embed"

	"encoding/json"

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

type RawUrl struct {
	Url     string  `json:"url"`
	CpuCost float64 `json:"CpuCost"`
}

type UrlNode struct {
	prefix   string
	cpuCost  float64
	children []UrlNode
}

type Limiter struct {
	usage float64
	limit float64
	mu    sync.Mutex
}

func (l *Limiter) Reserve(r *http.Request, limits UrlNode) bool {
	urlCost := 0.0
	path := r.URL.Path
	parts := strings.Split(path, "/")
	node := limits
	inTree := true
	for i := 0; i < len(parts) && inTree; i++ {
		part := parts[i]
		var idx int
		idx, inTree = slices.BinarySearchFunc(
			node.children,
			part,
			func(a UrlNode, b string) int {
				if a.prefix == "*" {
					return 0
				}
				return strings.Compare(a.prefix, b)
			})
		if inTree {
			node = node.children[idx]
		}
	}

	if inTree {
		urlCost = node.cpuCost
	}

	resp := false
	l.mu.Lock()
	if urlCost+l.usage < l.limit {
		l.usage += urlCost
		fmt.Printf("Reserving: %f, Usage: %f\n", urlCost, l.usage)
		resp = true
	}
	l.mu.Unlock()
	return resp

}

func (l *Limiter) Free(r *http.Request, limits UrlNode) {
	urlCost := 0.0
	path := r.URL.Path
	parts := strings.Split(path, "/")
	node := limits
	inTree := true
	for i := 0; i < len(parts) && inTree; i++ {
		part := parts[i]
		var idx int
		idx, inTree = slices.BinarySearchFunc(
			node.children,
			part,
			func(a UrlNode, b string) int {
				if a.prefix == "*" {
					return 0
				}
				return strings.Compare(a.prefix, b)
			})
		if inTree {
			node = node.children[idx]
		}
	}

	if inTree {
		urlCost = node.cpuCost
	}

	fmt.Printf("in Free\n")
	l.mu.Lock()
	l.usage -= urlCost
	if l.usage < 0.0 {
		l.usage = 0.0
	}
	fmt.Printf("Freeing: %f, Usage: %f\n", urlCost, l.usage)
	l.mu.Unlock()

}

type UserLimiter struct {
	l Limiter
	t time.Time
}

type SyncMap struct {
	ma map[string]*UserLimiter
	mu sync.Mutex
}

//go:embed rate_config.json
var rateConfig []byte

func main() {
	certificate,
		dbDirPath,
		defaultDb,
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
		fileHeader := []byte{0x53, 0x51, 0x4c, 0x69, 0x74, 0x65, 0x20, 0x66, 0x6f,
			0x72, 0x6d, 0x61, 0x74, 0x20, 0x33, 0x00}
		h := make([]byte, 16)
		f.Read(h)
		if !bytes.Equal(h, fileHeader) {
			continue
		}

		db, err := tdb.OpenTaxonomyDBcheck(path)
		if err != nil {
			log.Fatal(fmt.Sprintf("error while opening database %s: ", path),
				err.Error())
		}
		defer db.Close()
		dbs[n] = apiv2.Database{Path: path, Db: db}
		if len(dbs) == 1 {
			dbs["latest"] = apiv2.Database{Path: path, Db: db}
		}

	}
	d, ok := dbs[defaultDb]
	if ok {
		dbs["latest"] = d
	} else {
		fmt.Printf("Could not find provided default db: %s, "+
			"falling back to %s (%s) instead.\n", defaultDb,
			filepath.Base(dbs["latest"].Path), dbs["latest"].Path)
	}

	docsPref := "/docs"
	apiPref := "/api"
	docsV1Pref := docsPref + apiPref + "/v1"
	docsV2Pref := docsPref + apiPref + "/v2"
	isLocal := host == "http://localhost"

	neverv1.RegisterRoutes(apiPref+"/v1", docsV1Pref, dbs["latest"].Path,
		dateFilePath)
	apiv2.RegisterRoutes(apiPref+"/v2", host+":"+strconv.Itoa(port), dbs)
	docsv2.RegisterRoutes(docsV2Pref, isLocal, port, dbDirPath)

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
							Warning: fmt.Sprintf(
								"Couldn't parse remote address %s during %s request",
								r.RemoteAddr,
								r.URL.Path,
							),
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

	var rawUrls []RawUrl
	err = json.Unmarshal(rateConfig, &rawUrls)

	slices.SortFunc(rawUrls, func(a, b RawUrl) int {
		return strings.Compare(a.Url, b.Url)
	})
	urlLimits := UrlNode{}
	for _, rawUrl := range rawUrls {
		urlParts := strings.Split(rawUrl.Url, "/")
		curr := &urlLimits
		for _, part := range urlParts {
			idx, found := slices.BinarySearchFunc(
				curr.children,
				part,
				func(a UrlNode, b string) int {
					return strings.Compare(a.prefix, b)
				})

			if found {
				curr = &curr.children[idx]
			} else {
				node := UrlNode{prefix: part, cpuCost: rawUrl.CpuCost}
				curr.children = append(curr.children, node)
				curr = &curr.children[len(curr.children)-1]
			}
		}
	}

	PrintTree(urlLimits, 0)

	generalLimiter := Limiter{usage: 0, limit: 85}

	userLimiters := SyncMap{ma: make(map[string]*UserLimiter)}

	middlewareLimiter := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rateLimiting := !noRateLimit
			if rateLimiting {
				if generalLimiter.Reserve(r, urlLimits) {
					ip, _, err := net.SplitHostPort(r.RemoteAddr)
					if err != nil {
						fmt.Printf("Couldn't parse remote address %s\n", r.RemoteAddr)
					}

					userLimiters.mu.Lock()
					userLimiter, ok := userLimiters.ma[ip]
					if !ok {
						userLimiter = new(UserLimiter{
							l: Limiter{usage: 0.0, limit: 30.0},
							t: time.Now(),
						})
						userLimiters.ma[ip] = userLimiter
					}
					userLimiters.mu.Unlock()

					allowed := userLimiter.l.Reserve(r, urlLimits)

					if allowed {
						next.ServeHTTP(w, r)
						generalLimiter.Free(r, urlLimits)
						userLimiter.l.Free(r, urlLimits)
					} else {
						w.WriteHeader(http.StatusTooManyRequests)
						w.Write([]byte("You've exceed your rate limit\n"))
						generalLimiter.Free(r, urlLimits)
					}
				} else {
					w.WriteHeader(http.StatusServiceUnavailable)
					w.Write([]byte("Server is busy\n"))
				}
			} else {
				next.ServeHTTP(w, r)
			}

		})
	}

	go func(ul *SyncMap) {
		for {
			time.Sleep(5 * time.Minute)
			userLimiters.mu.Lock()
			threshold := time.Now().Add(time.Minute * -5)
			for k := range userLimiters.ma {
				if userLimiters.ma[k].t.Before(threshold) {
					delete(userLimiters.ma, k)
				}
			}
			userLimiters.mu.Unlock()
			fmt.Print("cleared map ")
			fmt.Println(userLimiters)
		}
	}(&userLimiters)

	middlewareTimeout := func(next http.Handler, d time.Duration) http.Handler {
		return http.TimeoutHandler(next, d, "Request timed out. Please compute "+
			"large requests locally.\n")
	}

	handlerChain :=
		middlewareLogger(
			middlewareLimiter(
				middlewareTimeout(
					http.DefaultServeMux,
					30*time.Second,
				)))
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

	term, c := signal.NotifyContext(context.Background(),
		syscall.SIGTERM, syscall.SIGINT)
	defer c()
	<-term.Done()

	fmt.Println("...Stopping server")
	ctx, cancel := context.WithTimeout(
		context.Background(),
		1500*time.Millisecond,
	)
	defer cancel()
	err = s.Shutdown(ctx)

	if err != nil {
		util.LogInfoDef(
			util.InfoEntry{
				Description: "Server shutdown with error: " + err.Error(),
			})
	} else {
		util.LogInfoDef(
			util.InfoEntry{
				Description: "Server shutdown gracefully",
			})
	}
	fmt.Println("Shutdown complete")

}

func ioHandling() (
	string,
	string,
	string,
	string,
	string,
	string,
	int,
	bool,
) {
	util.PrepLog("never")

	clio.Usage(
		"-o 10.254.1.21 -p 443",
		"This is the webserver never. It hosts the neighbors' REST API"+
			"versions 1 and 2, "+
			"as well as their respective documentation. "+
			"New packages may be added in a similar fashion as seen "+
			"in the main function.",
		"Starts the webserver at specified address with given port.")

	cFlag := flag.String("c", "certificates/cert.pem", "certificate")
	dFlag := flag.String("d", "databases", "path to database directory"+
		"relative to current working directory")
	ddFlag := flag.String("D", "neidb", "base name of the database that"+
		"should be used by default.")
	uFlag := flag.String("u", "updated.txt", "path to dateFile"+
		"relative to current working directory")
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

	return *cFlag, *dFlag, *ddFlag, *uFlag, *kFlag, *oFlag, *pFlag, *rFlag

}

func PrintTree(node UrlNode, depth int) {
	for range depth {
		fmt.Print(" ")
	}
	fmt.Printf(node.prefix+" | Cost: %.2f\n", node.cpuCost)
	for _, child := range node.children {
		PrintTree(child, depth+2)
	}
}
