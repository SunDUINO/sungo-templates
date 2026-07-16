/*
 * ╔════════════════════════════════════════════════════════════════╗
 * ║ Go_REST_webui                                                  ║
 * ║ Plik / File: main.go                                           ║
 * ╠════════════════════════════════════════════════════════════════╣
 * ║ Autor / Author:                                                ║
 * ║   SunRiver                                                     ║
 * ║   Lothar TeaM                                                  ║
 * ╠════════════════════════════════════════════════════════════════╣
 * ║ GitHub  : Go_REST_webui                                        ║
 * ║ WWW     : https://lothar-team.pl                               ║
 * ║ Forum   : https://forum.lothar-team.pl                         ║
 * ║                                                                ║
 * ║ Licencja / License: MIT                                        ║
 * ║ Rok / Year: 2026                                               ║
 * ╚════════════════════════════════════════════════════════════════╝
 */
package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"
	
    "Go_REST_webui/src/Internal/api"
	"Go_REST_webui/src/Internal/store"

)

//go:embed web/static
var webFS embed.FS

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	st := store.New()
	startedAt := time.Now()
	srv := api.NewServer(st, startedAt)

	staticFS, err := fs.Sub(webFS, "web/static")
	if err != nil {
		log.Fatalf("nie udało się przygotować plików statycznych: %v", err)
	}

	mux := http.NewServeMux()

	// --- API ---
	mux.HandleFunc("GET /api/health", srv.Health)
	mux.HandleFunc("GET /api/status", srv.Status)
	mux.HandleFunc("GET /api/items", srv.ListItems)
	mux.HandleFunc("POST /api/items", srv.CreateItem)
	mux.HandleFunc("GET /api/items/{id}", srv.GetItem)
	mux.HandleFunc("PUT /api/items/{id}", srv.UpdateItem)
	mux.HandleFunc("DELETE /api/items/{id}", srv.DeleteItem)

	// --- Web UI (embedded) ---
	mux.Handle("/", http.FileServer(http.FS(staticFS)))

	handler := api.WithLogging(api.WithRequestCounter(srv, mux))

	addr := ":" + port
	log.Printf("mikrousługa uruchomiona -> http://localhost%s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("serwer zakończył działanie z błędem: %v", err)
	}
}