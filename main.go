package main

import (
	"fmt"
	"log"
	"net/http"
	"neverv2/docs"
)

func main() {
	docs.RegisterRoutes("/docs/api/v2")

	http.HandleFunc("/{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/api/v2", http.StatusSeeOther)
	})

	fmt.Println("Starting server...")
	log.Fatal(http.ListenAndServe(":8080", nil))

	fmt.Println("...Stopping server")
}
