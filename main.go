// Command mine-pflow-xyz serves mine.pflow.xyz. Today that is the landing page,
// a health check and the push-to-deploy webhook; ROADMAP.md has what comes next.
package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
)

//go:embed web
var webFS embed.FS

// App holds what the handlers need. deploy is swapped out in tests.
type App struct {
	deploy func() (string, error)
}

func (a *App) routes() http.Handler {
	site, err := fs.Sub(webFS, "web")
	if err != nil {
		panic(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "ok")
	})
	mux.HandleFunc("POST /webhook/github", a.handleWebhook)
	// http.FileServer answers / with index.html and 404s anything else.
	mux.Handle("GET /", http.FileServerFS(site))
	return mux
}

func main() {
	port := flag.Int("port", 8099, "listen port")
	flag.Parse()

	app := &App{deploy: runDeploy}
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("mine-pflow-xyz listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, app.routes()))
}
