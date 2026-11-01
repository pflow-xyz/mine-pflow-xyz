package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	port := flag.Int("port", 8099, "listen port")
	flag.Parse()
	if v := os.Getenv("PORT"); v != "" {
		fmt.Sscanf(v, "%d", port)
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "mine-pflow-xyz")
	})

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("mine-pflow-xyz listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
