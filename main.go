package main

import (
	"fmt"
	"net/http"

	_ "github.com/joho/godotenv/autoload"
)

/*
func startBroadcast(h *Hub) {
	cmd := exec.Command("ffmpeg", "-re", "-i", "./musique/daft-punk/discovery/01-07-superheroes.flac", "-vn", "-f", "mp3", "-ab", "192k", "pipe:1")

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
*/

func runPlaylist(hub *Hub) {
	queue := []AudioSource{
		&SpeechSource{
			Text: `
[French accent] Bonsoir, beautiful people of the night. You're listening to the smoothest frequencies in the galaxy.

[giggle] And now… something heroic is about to happen.

[dramatic tone] From the legendary robots who taught the world how to groove… here comes a track that feels like a comic book exploding on the dancefloor.

[whisper] Close your eyes… imagine neon lights, chrome helmets, and pure energy.

[sarcastic] Yes, yes… your speakers are about to work very hard.

Turn it up… this is "Superheroes" by Daft Punk.
`,
			VoiceID: "rbFGGoDXFHtVghjHuS3E",
		},
		&FileSource{Path: "./musique/daft-punk/discovery/01-07-superheroes.flac"},
	}

	for {
		for _, source := range queue {
			if err := source.Stream(hub); err != nil {
				fmt.Printf("Error streaming source: %v\n", err)
			}
		}
	}
}

func main() {
	hub := NewHub()

	go runPlaylist(hub)

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
