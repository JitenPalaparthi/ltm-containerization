package main

import (
 "fmt"
 "log"
 "net/http"
 "os"
 "strconv"
 "sync/atomic"
 "time"
)
func envInt(key string, fallback int) int { n,e:=strconv.Atoi(os.Getenv(key)); if e!=nil{return fallback}; return n }
func main(){
 start:=time.Now(); startup:=time.Duration(envInt("STARTUP_SECONDS",20))*time.Second
 readyAfter:=time.Duration(envInt("READY_AFTER_SECONDS",30))*time.Second
 failAfter:=time.Duration(envInt("FAIL_AFTER_SECONDS",0))*time.Second
 crashAfter:=time.Duration(envInt("CRASH_AFTER_SECONDS",0))*time.Second
 var requests atomic.Int64
 http.HandleFunc("/",func(w http.ResponseWriter,r *http.Request){fmt.Fprintf(w,"Go probe lab uptime=%s requests=%d\n",time.Since(start).Truncate(time.Second),requests.Add(1))})
 http.HandleFunc("/startup",func(w http.ResponseWriter,r *http.Request){if time.Since(start)<startup{http.Error(w,"still starting",503);return};fmt.Fprintln(w,"started")})
 http.HandleFunc("/ready",func(w http.ResponseWriter,r *http.Request){elapsed:=time.Since(start);if elapsed<readyAfter || (failAfter>0 && elapsed>=failAfter){http.Error(w,"not ready",503);return};fmt.Fprintln(w,"ready")})
 http.HandleFunc("/live",func(w http.ResponseWriter,r *http.Request){if failAfter>0 && time.Since(start)>=failAfter{http.Error(w,"unhealthy",503);return};fmt.Fprintln(w,"alive")})
 go func(){if crashAfter>0{time.Sleep(crashAfter);log.Printf("intentional application error after %s: exiting 1",crashAfter);os.Exit(1)}}()
 log.Printf("starting server :8080 startup=%s readyAfter=%s failAfter=%s crashAfter=%s",startup,readyAfter,failAfter,crashAfter)
 log.Fatal(http.ListenAndServe(":8080",nil))
}
