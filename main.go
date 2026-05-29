package main

import (
	"github.com/evolbioinf/never/util"

	"github.com/evolbioinf/clio"

	"flag"
	"os"

	"golang.org/x/time/rate"

	neverv1 "github.com/evolbioinf/never/api/v1/never"
	apiv2 "github.com/evolbioinf/never/api/v2"
	docsv2 "github.com/evolbioinf/never/docs/v2"

	"net/http"

	"fmt"
	"log"
)

func main() {
	certificate, dbPath, dateFilePath, privateKey, host, port := ioHandling()

	limiter := rate.NewLimiter(rate.Limit(1), 1)

	docsv2.RegisterRoutes("/docs/api/v2", host == "", port)
	apiv2.RegisterRoutes("/api/v2", dbPath, func(w http.ResponseWriter) bool { return handleLimit(limiter, w) })
	neverv1.RegisterRoutes("/api/v1", "/docs/api/v1", dbPath, dateFilePath)

	http.HandleFunc("/{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/api/v2", http.StatusSeeOther)
	})

	if host == "" {
		fmt.Printf("Starting server at http://localhost:%d ...\n", port)
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
	} else {
		fmt.Printf("Starting server at %s:%d ...\n", host, port)
		log.Fatal(http.ListenAndServeTLS(fmt.Sprintf("%s:%d", host, port), certificate, privateKey, nil))
	}

	fmt.Println("...Stopping server")
}

func ioHandling() (string, string, string, string, string, int) {
	util.PrepLog("never")

	clio.Usage(
		"-o 10.254.1.21 -p 443",
		"This is the webserver never. It hosts the neighbors' REST API versions 1 and 2, "+
			"as well as their respective documentation. "+
			"New packages may be added in a simmilar fashion as seen "+
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

func handleLimit(limiter *rate.Limiter, w http.ResponseWriter) bool {
	allowed := limiter.Allow()
	if !allowed {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte("Exceeded rate limit"))
	}

	return allowed

}
