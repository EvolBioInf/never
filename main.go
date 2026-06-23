package main

import (
	"runtime/debug"

	"github.com/evolbioinf/never/util"

	"github.com/evolbioinf/clio"

	"flag"
	"os"

	neverv1 "github.com/evolbioinf/never/api/v1/never"
	apiv2 "github.com/evolbioinf/never/api/v2"
	docsv2 "github.com/evolbioinf/never/docs/v2"

	"net/http"

	"time"

	"golang.org/x/time/rate"

	"sync"

	"net"
	"strings"

	"fmt"
	"log"
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
	debug.PrintStack()
	w.ResponseWriter.WriteHeader(statusCode)
}

type SyncMap struct {
	ma map[string]*rate.Limiter
	mu sync.Mutex
}

func main() {
	certificate, dbPath, dateFilePath, privateKey, host, port := ioHandling()

	docsPref := "/docs"
	apiPref := "/api"
	docsV1Pref := docsPref + apiPref + "/v1"
	docsV2Pref := docsPref + apiPref + "/v2"

	neverv1.RegisterRoutes(apiPref+"/v1", docsV1Pref, dbPath, dateFilePath)
	apiv2.RegisterRoutes(apiPref+"/v2", dbPath)
	docsv2.RegisterRoutes(docsV2Pref, host == "", port)

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
			allowed := !strings.HasPrefix(r.URL.Path, apiPref)
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

	var err error
	if host == "" {
		fmt.Printf("Starting server at http://localhost:%d ...\n", port)
		err = http.ListenAndServe(fmt.Sprintf(":%d", port), handlerChain)
	} else {
		fmt.Printf("Starting server at %s:%d ...\n", host, port)
		err = http.ListenAndServeTLS(fmt.Sprintf("%s:%d", host, port), certificate, privateKey, handlerChain)
	}

	if err != http.ErrServerClosed {
		log.Fatal(err)
	}

	fmt.Println("...Stopping server")
}

func ioHandling() (string, string, string, string, string, int) {
	util.PrepLog("never")

	clio.Usage(
		"-o 10.254.1.21 -p 443",
		"This is the webserver never. It hosts the neighbors' REST API versions 1 and 2, "+
			"as well as their respective documentation. "+
			"New packages may be added in a similar fashion as seen "+
			"in the main function.",
		"Starts the webserver at specified address with given port.")

	cFlag := flag.String("c", "certificates/cert.pem", "certificate")
	dbFlag := flag.String("db", "neidb", "path to database from execution position")
	dFlag := flag.String("d", "updated.txt", "path to dateFile from execution position")
	kFlag := flag.String("k", "certificates/private_key.pem", "private key")
	oFlag := flag.String("o", "", "host address")
	pFlag := flag.Int("p", 8080, "port")

	vFlag := flag.Bool("v", false, "print progam info")

	flag.Parse()

	if *vFlag {
		util.PrintInfo()
		os.Exit(0)
	}

	return *cFlag, *dbFlag, *dFlag, *kFlag, *oFlag, *pFlag

}
