package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	seconds := 45
	if s := os.Getenv("CRASH_AFTER_SECONDS"); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil || n < 1 {
			log.Fatalf("CRASH_AFTER_SECONDS must be a positive integer: %q", s)
		}
		seconds = n
	}
	start := time.Now()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Go crash demo running; uptime=%s\n", time.Since(start).Truncate(time.Second))
	})
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK); fmt.Fprintln(w, "ok") })
	go func() {
		log.Printf("HTTP listening on :8080; intentionally exiting after %d seconds", seconds)
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatal(err)
		}
	}()
	time.Sleep(time.Duration(seconds) * time.Second)
	log.Printf("Intentional crash after %d seconds (exit code 1)", seconds)
	os.Exit(1)
}
