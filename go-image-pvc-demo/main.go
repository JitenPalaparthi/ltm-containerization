package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const dir = "/data/images"
const maxUpload = 10 << 20

var page = template.Must(template.New("index").Parse(`<!doctype html><html><head><meta charset="utf-8"><title>Kubernetes PVC Image Demo</title><style>body{font:16px system-ui;max-width:900px;margin:50px auto;padding:0 16px}img{width:180px;height:150px;object-fit:cover;border-radius:8px}.grid{display:flex;gap:20px;flex-wrap:wrap}.item{padding:10px;border:1px solid #ddd;border-radius:10px}</style></head><body><h1>Image uploads on a Kubernetes PVC</h1><p>Upload an image, delete the Deployment, redeploy it, and the image remains on the PVC.</p><form action="/upload" method="post" enctype="multipart/form-data"><input type="file" name="image" accept="image/jpeg,image/png,image/gif,image/webp" required><button>Upload</button></form><h2>Stored images ({{len .}})</h2><div class="grid">{{range .}}<div class="item"><a href="/images/{{.}}" target="_blank"><img src="/images/{{.}}" alt="image"></a><p>{{.}}</p></div>{{else}}<p>No images uploaded yet.</p>{{end}}</div></body></html>`))

func main() {
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/", index)
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK); fmt.Fprintln(w, "ok") })
	http.HandleFunc("/upload", upload)
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir(dir))))
	log.Println("listening on :8080, images stored at", dir)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
func index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		http.Error(w, "storage unavailable", 500)
		return
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() {
			names = append(names, e.Name())
		}
	}
	if err := page.Execute(w, names); err != nil {
		log.Println(err)
	}
}
func upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", 405)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxUpload+1024*1024)
	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "image missing or too large", 400)
		return
	}
	defer file.Close()
	var sniff [512]byte
	n, err := io.ReadFull(file, sniff[:])
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		http.Error(w, "read failed", 400)
		return
	}
	mime := http.DetectContentType(sniff[:n])
	ext := map[string]string{"image/jpeg": ".jpg", "image/png": ".png", "image/gif": ".gif", "image/webp": ".webp"}[mime]
	if ext == "" {
		http.Error(w, "only JPG, PNG, GIF, WEBP allowed", 400)
		return
	}
	_ = header
	name := fmt.Sprintf("%d-%d%s", time.Now().UnixNano(), os.Getpid(), ext)
	target, err := os.OpenFile(filepath.Join(dir, name), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		http.Error(w, "storage write failed", 500)
		return
	}
	ok := false
	defer func() {
		target.Close()
		if !ok {
			os.Remove(filepath.Join(dir, name))
		}
	}()
	if _, err = target.Write(sniff[:n]); err != nil {
		http.Error(w, "write failed", 500)
		return
	}
	if _, err = io.Copy(target, io.LimitReader(file, maxUpload+1)); err != nil {
		http.Error(w, "write failed", 500)
		return
	}
	if err = target.Sync(); err != nil {
		http.Error(w, "sync failed", 500)
		return
	}
	ok = true
	log.Printf("stored %s (%s)", strings.TrimSpace(name), mime)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
