package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
)

func startBroadcast(h *Hub) {
	cmd := exec.Command("ffmpeg", "-re", "-i", "./musique/daft-punk/discovery/Daft Punk - Discovery - 01-07 Superheroes.flac", "-vn", "-f", "mp3", "-ab", "192k", "pipe:1")

	// This is only for debugging
	// cmd.Stderr = os.Stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	buf := make([]byte, 4096)
	for {
		n, err := stdout.Read(buf)
		if n > 0 {
			chunk := make([]byte, n)
			copy(chunk, buf[:n])
			h.Broadcast(chunk)
		}
		if err != nil {
			log.Fatal(err)
			break
		}
	}
}

func main() {
	hub := NewHub()

	go startBroadcast(hub)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "audio/mpeg")
		w.Header().Set("Cache-control", "no-cache")
		w.Header().Set("Transfer-Encoding", "chunked")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming is not supported", http.StatusInternalServerError)
			return
		}

		ch := make(chan []byte, 10)
		hub.AddClient(ch)
		defer hub.RemoveClient(ch)

		for {
			select {
			case chunk, ok := <-ch:
				if !ok {
					return
				}
				if _, err := w.Write(chunk); err != nil {
					// Client disconnected
					return
				}
				flusher.Flush()
			case <-r.Context().Done():
				// Client disconnected
				return
			}
		}
	})

	fmt.Println("🎶 on port 8080")
	http.ListenAndServe(":8080", nil)
}
