// pending-pods-lab is a tiny HTTP service used by the Kubernetes scheduling lab.
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	name, err := os.Hostname()
	if err != nil {
		log.Fatalf("read hostname: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintf(w, "pending-pods-lab is running\npod: %s\ntime: %s\n", name, time.Now().UTC().Format(time.RFC3339))
	})
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})

	server := &http.Server{Addr: ":8080", Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	log.Printf("listening on %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}
