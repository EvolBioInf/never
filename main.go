package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/evolbioinf/clio"
	neverv1 "github.com/evolbioinf/never/api/v1/never"
	apiv2 "github.com/evolbioinf/never/api/v2"
	docsv2 "github.com/evolbioinf/never/docs/v2"
	"github.com/evolbioinf/never/util"
)

func main() {
	host, certificate, privateKey, port := ioHandling()
	docsv2.RegisterRoutes("/docs/api/v3", host == "", port)
	apiv2.RegisterRoutes("/api/v2")
	neverv1.RegisterRoutes("/docs/api/v1")

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

func ioHandling() (string, string, string, int) {
	util.PrepLog("never")
	clio.Usage(
		"-o 10.254.1.21 -p 443",
		"This is the webserver never. It hosts the neighbors' REST API versions 1 and 2, "+
			"as well as their respective documentation. "+
			"New packages may be added in a simmilar fashion as seen "+
			"in the main function.",
		"Starts the webserver in local mode, without https.")

	lFlag := flag.String("o", "", "local mode")
	cFlag := flag.String("c", "certificates/cert.pem", "certificate")
	kFlag := flag.String("k", "certificates/private_key.pem", "private key")
	pFlag := flag.Int("p", 8080, "port")
	vFlag := flag.Bool("v", false, "print progam info")

	flag.Parse()

	if *vFlag {
		util.PrintInfo()
	}

	return *lFlag, *cFlag, *kFlag, *pFlag
}
