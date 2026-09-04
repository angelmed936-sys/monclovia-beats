package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type Beat struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

var (
	beats []Beat
	mu    sync.Mutex
)

func enableCORS(w http.ResponseWriter, r *http.Request) bool {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == "OPTIONS" {
		w.WriteHeader(200)
		return true
	}
	return false
}

func main() {
	// API - obtener beats
	http.HandleFunc("/api/beats", func(w http.ResponseWriter, r *http.Request) {
		if enableCORS(w, r) {
			return
		}
		mu.Lock()
		defer mu.Unlock()
		json.NewEncoder(w).Encode(beats)
	})

	// API - login
	http.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		if enableCORS(w, r) {
			return
		}
		var creds struct {
			User, Pass string `json:"user"`
			P          string `json:"pass"`
		}
		// por si mandas user/pass o User/Pass
		json.NewDecoder(r.Body).Decode(&creds)
		if (creds.User == "monclovia" && creds.P == "monclovia123") || (creds.User == "monclovia" && creds.Pass == "monclovia123") {
			json.NewEncoder(w).Encode(map[string]any{"ok": true})
		} else {
			w.WriteHeader(401)
			json.NewEncoder(w).Encode(map[string]any{"ok": false})
		}
	})

	// API - borrar
	http.HandleFunc("/api/beats/", func(w http.ResponseWriter, r *http.Request) {
		if enableCORS(w, r) {
			return
		}
		if r.Method == "DELETE" {
			idStr := strings.TrimPrefix(r.URL.Path, "/api/beats/")
			id, _ := strconv.Atoi(idStr)
			mu.Lock()
			if id >= 0 && id < len(beats) {
				beats = append(beats[:id], beats[id+1:]...)
			}
			mu.Unlock()
			json.NewEncoder(w).Encode(map[string]any{"ok": true})
		}
	})

	// SERVIR PAGINA - ESTO ES LO QUE ARREGLA TU 404
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		// 1. Si pide / busca public/index.html o index.html
		p := r.URL.Path
		if p == "/" {
			if _, err := os.Stat("public/index.html"); err == nil {
				http.ServeFile(w, r, "public/index.html")
				return
			}
			if _, err := os.Stat("index.html"); err == nil {
				http.ServeFile(w, r, "index.html")
				return
			}
		}

		// 2. Busca el archivo en public/
		publicPath := filepath.Join("public", p)
		if _, err := os.Stat(publicPath); err == nil {
			http.ServeFile(w, r, publicPath)
			return
		}

		// 3. Busca el archivo en la raiz
		rootPath := filepath.Join(".", p)
		if _, err := os.Stat(rootPath); err == nil {
			http.ServeFile(w, r, rootPath)
			return
		}

		// Si no encuentra nada, sirve index.html (para que no de 404)
		if _, err := os.Stat("public/index.html"); err == nil {
			http.ServeFile(w, r, "public/index.html")
			return
		}
		http.ServeFile(w, r, "index.html")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "10000"
	}
	fmt.Println("Corriendo en puerto", port)
	http.ListenAndServe(":"+port, nil)
}
