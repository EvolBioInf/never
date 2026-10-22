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

type UserLimiter struct {
	l Limiter
	t time.Time
}

type SyncMap struct {
	ma map[string]*UserLimiter
	mu sync.Mutex
}

type RequestEntry struct {
	r     *http.Request
	w     http.ResponseWriter
	t     time.Time
	c     float64
	state chan string
}

type RequestQueue struct {
	entries []RequestEntry
	maxSize int
	mu      sync.Mutex
}

func (l *Limiter) Reserve(cpuCost float64) bool {
	resp := false
	l.mu.Lock()
	if cpuCost+l.usage < l.limit {
		l.usage += cpuCost
		resp = true
	}
	l.mu.Unlock()
	return resp

}

func (l *Limiter) Free(cpuCost float64) {
	l.mu.Lock()
	l.usage -= cpuCost
	if l.usage < 0.0 {
		l.usage = 0.0
	}
	l.mu.Unlock()

}

func (ul *SyncMap) getOrCreate(remoteAddr string) *UserLimiter {
	ip, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		fmt.Printf("Couldn't parse remote address %s\n", remoteAddr)
	}

	ul.mu.Lock()
	userLimiter, ok := ul.ma[ip]
	if !ok {
		userLimiter = new(UserLimiter{
			l: Limiter{usage: 0.0, limit: 50.0},
			t: time.Now(),
		})
		ul.ma[ip] = userLimiter
	}
	ul.mu.Unlock()

	return userLimiter
}

func (q *RequestQueue) IsFull() bool {
	return len(q.entries) >= q.maxSize
}

func (q *RequestQueue) Push(re RequestEntry) bool {
	q.mu.Lock()
	var str string
	var resp = false
	if !q.IsFull() {
		q.entries = append(q.entries, re)
		str = "Pushed request to queue"
		resp = true
	} else {
		str = "Denied push to queue"
	}

	util.LogInfoDef(
		util.InfoEntry{
			RequestUrl:  re.r.URL.String(),
			Description: fmt.Sprintf("%s New size: %d", str, len(q.entries)),
		})

	q.mu.Unlock()
	return resp
}

func (q *RequestQueue) Pop() *RequestEntry {
	q.mu.Lock()
	if len(q.entries) > 0 {
		res := q.entries[0]
		q.entries = q.entries[1:]
		util.LogInfoDef(
			util.InfoEntry{
				RequestUrl: res.r.URL.String(),
				Description: fmt.Sprintf(
					"Popped from queue. New size: %d",
					len(q.entries),
				)})
		q.mu.Unlock()
		return &res
	} else {
		q.mu.Unlock()
		return nil
	}
}

func (l *Limiter) Capacity() float64 {
	l.mu.Lock()
	c := l.limit - l.usage
	l.mu.Unlock()
	return c
}

func (q *RequestQueue) activateEntries(
	generalLimiter *Limiter,
	userLimiters *SyncMap,
) {
	indexed := make([]int, 0)
	q.mu.Lock()
	for i := 0; i < len(q.entries) && generalLimiter.Capacity() > 0.1; i++ {
		curr := q.entries[i]
		ul := userLimiters.getOrCreate(curr.r.RemoteAddr)
		glAllowed, ulAllowed := limit(
			generalLimiter,
			&ul.l,
			curr.c,
		)

		if glAllowed && ulAllowed {
			printLimiterUsage(curr.r.URL.Path, true, generalLimiter, userLimiters)
			curr.state <- "ready"
			indexed = append(indexed, i)
		}
	}

	for i := len(indexed) - 1; i >= 0; i-- {
		q.entries = append(q.entries[:i], q.entries[i+1:]...)
	}

	if len(indexed) > 0 {
		util.LogInfoDef(
			util.InfoEntry{
				Description: fmt.Sprintf(
					"Pulled %d entries from queue. New size: %d",
					len(indexed),
					len(q.entries),
				)})
	}

	q.mu.Unlock()
}

func (q *RequestQueue) clearQueue() {
	const maxTime = 45
	keep := make([]RequestEntry, 0)
	removed := 0
	q.mu.Lock()
	now := time.Now()

	for _, e := range q.entries {
		if now.After(e.t.Add(maxTime * time.Second)) {
			removed++
			e.state <- "abort"
		} else {
			keep = append(keep, e)
		}
	}

	q.entries = keep
	if removed > 0 {
		util.LogInfoDef(
			util.InfoEntry{
				Description: fmt.Sprintf(
					"Cleared %d entries from queue. New size: %d",
					removed,
					len(q.entries),
				)})
	}
	q.mu.Unlock()
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
		logDir,
		vitaxDir,
		progDir,
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

	neverv1.RegisterRoutes(
		apiPref+"/v1",
		docsV1Pref,
		dbs["latest"].Path,
		dateFilePath,
	)
	apiv2.RegisterRoutes(
		apiPref+"/v2",
		host+":"+strconv.Itoa(port),
		dbs,
		progDir,
	)
	docsv2.RegisterRoutes(
		docsV2Pref,
		isLocal,
		port,
		dbDirPath,
		dateFilePath,
		dbs["latest"].Db,
	)

	vitaxFiles := http.FileServer(http.Dir(vitaxDir))
	http.Handle("/vitax/", http.StripPrefix("/vitax/", vitaxFiles))

	http.HandleFunc("/{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, docsV2Pref, http.StatusSeeOther)
	})

	util.SetupLog(logDir)
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
		for i := range len(urlParts) {
			part := urlParts[i]
			idx, found := slices.BinarySearchFunc(
				curr.children,
				part,
				func(a UrlNode, b string) int {
					return strings.Compare(a.prefix, b)
				})

			if found {
				curr = &curr.children[idx]
			} else {
				cost := 0.0
				if i == len(urlParts)-1 {
					cost = rawUrl.CpuCost
				}
				node := UrlNode{prefix: part, cpuCost: cost}
				curr.children = append(curr.children, node)
				curr = &curr.children[len(curr.children)-1]
			}
		}
	}

	generalLimiter := Limiter{usage: 0, limit: 85}

	userLimiters := SyncMap{ma: make(map[string]*UserLimiter)}

	q := RequestQueue{entries: make([]RequestEntry, 0), maxSize: 100}
	middlewareLimiter := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rateLimiting := !noRateLimit
			if rateLimiting {
				c := getUrlCost(r, urlLimits)
				ul := userLimiters.getOrCreate(r.RemoteAddr)
				glAllowed, ulAllowed := limit(
					&generalLimiter,
					&ul.l,
					c,
				)
				if glAllowed && ulAllowed {
					printLimiterUsage(r.URL.Path, true, &generalLimiter, &userLimiters)
					next.ServeHTTP(w, r)
					freeLimiters(&generalLimiter, &ul.l, c)
					printLimiterUsage(r.URL.Path, false, &generalLimiter, &userLimiters)
					go q.activateEntries(&generalLimiter, &userLimiters)

				} else {
					// In case queue is full
					reqEntry := RequestEntry{
						r:     r,
						w:     w,
						t:     time.Now(),
						c:     c,
						state: make(chan string),
					}

					if q.Push(reqEntry) {
						s := <-reqEntry.state
						if s == "ready" {
							next.ServeHTTP(w, r)
							freeLimiters(&generalLimiter, &ul.l, c)
							printLimiterUsage(r.URL.Path, false, &generalLimiter, &userLimiters)
							go q.activateEntries(&generalLimiter, &userLimiters)

						} else if s == "abort" {
							w.WriteHeader(http.StatusServiceUnavailable)
							w.Write([]byte("Server is busy. This request cannot be accepted at " +
								"this moment, please try again later.\n"))

						}

					} else if !glAllowed {
						w.WriteHeader(http.StatusServiceUnavailable)
						w.Write([]byte("Server is busy. This request cannot be accepted at " +
							"this moment, please try again later.\n"))

					} else if !ulAllowed {
						w.WriteHeader(http.StatusTooManyRequests)
						w.Write([]byte("Rate limit exceeded. This request cannot be " +
							"accepted at this moment, please try again later.\n"))
					}

				}
			} else {
				next.ServeHTTP(w, r)
			}

		})
	}

	go func() {
		for {
			time.Sleep(15 * time.Second)
			q.clearQueue()
		}
	}()

	go func(ul *SyncMap) {
		for {
			time.Sleep(5 * time.Minute)
			ul.mu.Lock()
			threshold := time.Now().Add(time.Minute * -5)
			for k := range ul.ma {
				if ul.ma[k].t.Before(threshold) {
					delete(ul.ma, k)
				}
			}
			ul.mu.Unlock()
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
	lFlag := flag.String("l", "logs/", "directory for log storage")
	vvFlag := flag.String("V", "vitax/", "directory containing vitax files")
	ppFlag := flag.String("P", "prog/", "directory containing neighbors programs")
	pFlag := flag.Int("p", 8080, "port")
	rFlag := flag.Bool("no-rate-limit", false, "Turn of rate limiting")
	vFlag := flag.Bool("v", false, "print progam info")

	flag.Parse()

	if *vFlag {
		util.PrintInfo()
		os.Exit(0)
	}

	return *cFlag, *dFlag, *ddFlag, *uFlag, *kFlag,
		*oFlag, *lFlag, *vvFlag, *ppFlag, *pFlag, *rFlag

}

func printTree(node UrlNode, depth int) {
	for range depth {
		fmt.Print(" ")
	}
	fmt.Printf(node.prefix+" | Cost: %.2f\n", node.cpuCost)
	for _, child := range node.children {
		printTree(child, depth+2)
	}
}

func getUrlCost(r *http.Request, limits UrlNode) float64 {
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

	return urlCost
}

func limit(
	generalLimiter *Limiter,
	userLimiter *Limiter,
	c float64,
) (glAllowed, ulAllowed bool) {
	glAllowed = generalLimiter.Reserve(c)
	if glAllowed {
		ulAllowed = userLimiter.Reserve(c)
		if !ulAllowed {
			// Need to free general limiter if user limiter is denying request
			generalLimiter.Free(c)
		}
	}
	return
}

func freeLimiters(
	generalLimiter *Limiter,
	userLimiter *Limiter,
	c float64,
) {
	generalLimiter.Free(c)
	userLimiter.Free(c)
}

var activeUrls = make(map[string]int)
var activeUrlsMu sync.Mutex

func printLimiterUsage(path string, add bool, generalLimiter *Limiter, userLimiters *SyncMap) {
	const liveDebug = false
	activeUrlsMu.Lock()
	if add {
		activeUrls[path]++
	} else {
		activeUrls[path]--
		if activeUrls[path] < 1 {
			delete(activeUrls, path)
		}
	}
	var sb strings.Builder

	for key := range activeUrls {
		fmt.Fprintf(&sb, "%s: %d | ", key, activeUrls[key])
		if liveDebug {
			sb.WriteString("\n")
		}
	}
	activeUrlsMu.Unlock()

	if liveDebug {
		fmt.Fprintln(&sb, "==== Limiters ====")
	}
	generalLimiter.mu.Lock()
	fmt.Fprintf(
		&sb,
		"General - usage: %.2f %%, limit: %.2f %% ",
		generalLimiter.usage,
		generalLimiter.limit,
	)
	generalLimiter.mu.Unlock()
	if liveDebug {
		sb.WriteString("\n")
	}

	userLimiters.mu.Lock()
	for key := range userLimiters.ma {
		fmt.Fprintf(
			&sb,
			"user lim %s - usage: %.2f %%, limit: %.2f %% ",
			key,
			userLimiters.ma[key].l.usage,
			userLimiters.ma[key].l.limit,
		)
		if liveDebug {
			sb.WriteString("\n")
		}
	}
	userLimiters.mu.Unlock()

	if liveDebug {
		fmt.Print("\033[H\033[2J" + sb.String())
	} else {
		util.LogInfoDef(
			util.InfoEntry{
				RequestUrl:  path,
				Description: sb.String(),
			})
	}
}
