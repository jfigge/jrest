package internal

import (
	"encoding/json"
	"fmt"
	"jrest/internal/handlers/routing"
	"jrest/internal/models"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"
)

var (
	extensions = []string{"", ".yaml", ".yml", ".json"}
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
		source:   &models.Source{},
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
				if event.Has(fsnotify.Rename) || event.Has(fsnotify.Remove) {
					time.Sleep(1 * time.Second)
					err = watcher.Add(a.filename)
					if err != nil {
						log.Fatalf("source file `%s` cannot be found", a.filename)
					}
					log.Println("modified file:", event.Name)
					a.loadSource()
					a.source.LogPaths()
				} else if event.Has(fsnotify.Write) {
					log.Println("modified file:", event.Name)
					a.loadSource()
					a.source.LogPaths()
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
	err = watcher.Add(app.filename)
	if err != nil {
		log.Fatal(err)
	}

	return &app
}

func (a *App) loadSource() {
	var bs []byte
	var err error
	for _, extension := range extensions {
		extendedFilename := fmt.Sprintf("%s%s", a.filename, extension)
		bs, err = os.ReadFile(extendedFilename)
		if err == nil {
			a.filename = extendedFilename
			break
		}
	}
	if err != nil {
		log.Fatalf("unable to read %s: %v\n", a.filename, err)
	}

	switch filepath.Ext(a.filename)[1:] {
	case "json":
		err = json.Unmarshal(bs, a.source)
	case "yml", "yaml":
		err = yaml.Unmarshal(bs, a.source)
	}
	if err != nil {
		log.Fatalf("unable to process %s: %v", a.filename, err)
	}

	a.source.ApplyDefaults()
	a.source.Cleanse()
}

func (a *App) Serve() {
	listenAddress := fmt.Sprintf("%s:%d", a.source.Host, a.source.Port)
	mux := http.NewServeMux()
	mux.Handle("/", routing.BaseHandler(a.source))

	log.Printf("Starting server: %s%s\n", listenAddress, a.source.Base)
	a.source.LogPaths()
	err := http.ListenAndServe(listenAddress, mux)
	if err != nil {
		log.Fatalf("unable to start server: %v", err)
	}
}
