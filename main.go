package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/evolbioinf/clio"
	apiv2 "github.com/evolbioinf/never/api/v2"
	docsv2 "github.com/evolbioinf/never/docs/v2"
	"github.com/evolbioinf/never/util"
)

func main() {
	local, port := ioHandling()
	docsv2.RegisterRoutes("/docs/api/v2", local, port)
	apiv2.RegisterRoutes("/api/v2")

	http.HandleFunc("/{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/api/v2", http.StatusSeeOther)
	})

	fmt.Printf("Starting server at Port %d...\n", port)
	if local {
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
	} else {
		log.Fatal(http.ListenAndServeTLS(fmt.Sprintf(":%d", port), "certificates/cert.pem", "certificates/key.pem", nil))
	}

	fmt.Println("...Stopping server")
}

func ioHandling() (bool, int) {
	util.PrepLog("never")
	clio.Usage(
		"-l",
		"This is the webserver neverV2. It hosts the neighbors' REST API versions 1 and 2, "+
			"as well as their documentations. New packages may be added in a simmilar fashion as seen "+
			"in the main function. ",
		"Starts the webserver in local mode, without https.")

	lFlag := flag.Bool("l", false, "local mode")
	pFlag := flag.Int("p", 8080, "port")
	vFlag := flag.Bool("v", false, "print progam info")

	flag.Parse()

	if *vFlag {
		util.PrintInfo()
	}

	return *lFlag, *pFlag
}
