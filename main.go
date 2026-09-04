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

type Beat struct {
	ID     int    `json:"ID"`
	Nombre string `json:"Nombre"`
	BPM    int    `json:"BPM"`
	Precio int    `json:"Precio"`
	Audio  string `json:"Audio"`
}

var (
	beats  []Beat
	mu     sync.Mutex
	nextID = 1
)

func cargarBeats() {
	data, err := os.ReadFile("beats.json")
	if err == nil {
		json.Unmarshal(data, &beats)
		for _, b := range beats {
			if b.ID >= nextID {
				nextID = b.ID + 1
			}
		}
	} else {
		beats = []Beat{{ID: 1, Nombre: "Trap Sad - Lluvia", BPM: 140, Precio: 500, Audio: "/audio/1.mp3"}}
		nextID = 2
		guardarBeats()
	}
}
func guardarBeats() {
	data, _ := json.MarshalIndent(beats, "", " ")
	os.WriteFile("beats.json", data, 0644)
}

func main() {
	cargarBeats()

	http.HandleFunc("/api/beats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "GET" {
			json.NewEncoder(w).Encode(beats)
			return
		}
		if r.Method == "POST" {
			var b Beat
			json.NewDecoder(r.Body).Decode(&b)
			mu.Lock()
			b.ID = nextID
			nextID++
			beats = append(beats, b)
			guardarBeats()
			mu.Unlock()
			json.NewEncoder(w).Encode(b)
			return
		}
		if r.Method == "DELETE" {
			// /api/beats/1 o /api/beats?id=1
			idStr := r.URL.Query().Get("id")
			if idStr == "" {
				parts := strings.Split(r.URL.Path, "/")
				idStr = parts[len(parts)-1]
			}
			id, _ := strconv.Atoi(idStr)
			mu.Lock()
			nuevos := []Beat{}
			for _, b := range beats {
				if b.ID != id {
					nuevos = append(nuevos, b)
				}
			}
			beats = nuevos
			guardarBeats()
			mu.Unlock()
			json.NewEncoder(w).Encode(map[string]string{"status": "borrado"})
			return
		}
	})
	http.HandleFunc("/api/beats/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.Method == "DELETE" {
			parts := strings.Split(r.URL.Path, "/")
			idStr := parts[len(parts)-1]
			id, _ := strconv.Atoi(idStr)
			mu.Lock()
			nuevos := []Beat{}
			for _, b := range beats {
				if b.ID != id {
					nuevos = append(nuevos, b)
				}
			}
			beats = nuevos
			guardarBeats()
			mu.Unlock()
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "borrado"})
		}
	})

	http.HandleFunc("/api/upload", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		r.ParseMultipartForm(32 << 20)
		file, handler, _ := r.FormFile("audio")
		defer file.Close()
		os.MkdirAll("./audio", 0755)
		dst, _ := os.Create(filepath.Join("./audio", handler.Filename))
		io.Copy(dst, file)
		dst.Close()
		json.NewEncoder(w).Encode(map[string]string{"url": "/audio/" + handler.Filename})
	})

	http.Handle("/audio/", http.StripPrefix("/audio/", http.FileServer(http.Dir("./audio"))))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "./index.html")
			return
		}
		if r.URL.Path == "/admin.html" || r.URL.Path == "/admin" {
			http.ServeFile(w, r, "./admin.html")
			return
		}
		http.ServeFile(w, r, "."+r.URL.Path)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("TIENDA EN: http://localhost:" + port)
	http.ListenAndServe(":"+port, nil)
}
