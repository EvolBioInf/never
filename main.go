package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	apiv2 "neverv2/api/v2"
	"neverv2/clui"
	docsv2 "neverv2/docs/v2"

	cors "github.com/rs/cors"
)

func main() {
	local, port := ioHandling()
	docsv2.RegisterRoutes("/docs/api/v2", local, port)
	apiv2.RegisterRoutes("/api/v2")

	http.HandleFunc("/{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/api/v2", http.StatusSeeOther)
	})

	if local {
		fmt.Println("Initializing cors middleware")
		c := cors.New(cors.Options{
			AllowedOrigins: []string{"http://localhost:8080"},
			AllowedMethods: []string{http.MethodGet},
			Debug:          true,

			AllowCredentials: true,
		})

		fmt.Printf("Starting server at Port %d...\n", port)
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), c.Handler(http.DefaultServeMux)))
	} else {
		fmt.Printf("Starting server at Port %d...\n", port)
		log.Fatal(http.ListenAndServeTLS(fmt.Sprintf(":%d", port), "certificates/cert", "certificates/key", nil))
	}

	fmt.Println("...Stopping server")
}

func ioHandling() (bool, int) {
	flag.Usage = usage

	lFlag := flag.Bool("l", false, "local mode")
	pFlag := flag.Int("p", 8080, "port")
	vFlag := flag.Bool("v", false, "print progam info")

	flag.Parse()

	if *vFlag {
		clui.Info("2.0.0", "2026-05-18")
	}

	return *lFlag, *pFlag
}

func usage() {
	clui.Usage(
		"This is the webserver neverV2. It hosts the neighbors' REST API versions 1 and 2, "+
			"as well as their documentations. New packages may be added in a simmilar fashion as seen "+
			"in the main function. ",
		"-l",
		"Starts the webserver in local mode, without https.")
}
