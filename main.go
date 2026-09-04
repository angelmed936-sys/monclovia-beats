package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

const ADMIN_USER = "monclovia"
const ADMIN_PASS = "monclovia123"

type Beat struct {
	ID    int    `json:"ID"`
	Title string `json:"Title"`
}

var (
	beats []Beat
	mu    sync.Mutex
)

func adminAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != ADMIN_USER || pass != ADMIN_PASS {
			w.Header().Set("WWW-Authenticate", `Basic realm="Admin"`)
			http.Error(w, "No autorizado", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func main() {
	// Cargar beats existentes si hay
	loadBeats()

	http.HandleFunc("/api/beats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		mu.Lock()
		defer mu.Unlock()
		json.NewEncoder(w).Encode(beats)
	})

	http.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		if r.Method == "OPTIONS" {
			return
		}
		var creds struct {
			User string `json:"user"`
			Pass string `json:"pass"`
		}
		json.NewDecoder(r.Body).Decode(&creds)
		if creds.User == ADMIN_USER && creds.Pass == ADMIN_PASS {
			w.Write([]byte(`{"ok": true}`))
		} else {
			http.Error(w, "credenciales incorrectas", 401)
		}
	})

	http.HandleFunc("/api/upload", adminAuth(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		// ... tu logica de subida ...
		fmt.Println("Upload autorizado")
	}))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("public", "index.html"))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	fmt.Println("Servidor en puerto", port)
	http.ListenAndServe(":"+port, nil)
}

func loadBeats() {
	// si no hay archivo, inicia vacio
	beats = []Beat{}
}

func init() {
	// para que no marque error de imports no usados
	_ = io.ReadAll
	_ = strconv.Itoa
	_ = strings.Contains
}
