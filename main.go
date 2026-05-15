package main

import (
	"fmt"
	"log"
	"net/http"
	apiv2 "neverv2/api/v2"
	docsv2 "neverv2/docs/v2"
)

func main() {
	docsv2.RegisterRoutes("/docs/api/v2")
	apiv2.RegisterRoutes("/api/v2")

	http.HandleFunc("/{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/api/v2", http.StatusSeeOther)
	})

	fmt.Println("Starting server at Port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
	log.Fatal(http.ListenAndServe(":8080", nil))

	fmt.Println("...Stopping server")
}
