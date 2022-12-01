package internal

import (
	"encoding/json"
	"fmt"
	"jrest/internal/models"
	"jrest/internal/security"
	"log"
	"net/http"
	"os"

	"github.com/fsnotify/fsnotify"
)

type App struct {
	filename string
	watcher  *fsnotify.Watcher
	source   *models.Source
}

func NewApp(filename string) *App {
	// Create new watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	app := App{
		filename: filename,
		watcher:  watcher,
	}
	app.loadSource()

	// Start listening for events.
	go func(a *App) {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) {
					log.Println("modified file:", event.Name)
					a.loadSource()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}(&app)

	// Add a path.
	err = watcher.Add(filename)
	if err != nil {
		log.Fatal(err)
	}

	return &app
}

func (a *App) loadSource() {
	bs, err := os.ReadFile(a.filename)
	if err != nil {
		log.Fatalf("unable to read %s: %v\n", a.filename, err)
	}

	source := models.Source{}

	err = json.Unmarshal(bs, &source)
	if err != nil {
		log.Fatalf("unable to process %s: %v", a.filename, err)
	}

	source.ApplyDefaults()
	a.source = &source
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[len(a.source.Base):]
	m, ok := a.source.Api[path]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		a.auditLog(r.Method, path, "Not found")
		return
	}
	payload, ok := m[r.Method]
	if !ok {
		w.WriteHeader(http.StatusMethodNotAllowed)
		a.auditLog(r.Method, path, "Method not allowed")
		return
	}

	authorized := true
	if payload.Credentials != nil {
		authorized = security.CredentialsAuthorized(r, payload.Credentials)
	}
	if !authorized && payload.Bearer != nil {
		authorized = security.BearerAuthorized(r, payload.Bearer)
	}
	if !authorized {
		w.WriteHeader(http.StatusUnauthorized)
		a.auditLog(r.Method, path, "Not authorized")
		return
	}

	for key, value := range payload.Headers {
		w.Header().Set(key, value)
	}

	if payload.Status != 0 {
		w.WriteHeader(payload.Status)
		a.auditLog(r.Method, path, fmt.Sprintf("%d", payload.Status))
	}
	_, _ = w.Write(payload.Data)
	if payload.Status == 0 {
		a.auditLog(r.Method, path, "200")
	}
}

func (a *App) auditLog(method, path, status string) {
	log.Printf("Serving: %s:%s -> %s\n", method, path, status)
}

func (a *App) Serve() {
	listenAddress := fmt.Sprintf("%s:%d", a.source.Host, a.source.Port)
	mux := http.NewServeMux()
	mux.Handle(a.source.Base+"/", a)
	log.Printf("Starting server: %s%s\n", listenAddress, a.source.Base)
	err := http.ListenAndServe(listenAddress, mux)
	if err != nil {
		log.Fatalf("unable to start server: %v", err)
	}
}
