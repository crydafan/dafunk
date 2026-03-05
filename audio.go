package main

import (
	"context"
	"io"
	"os"
	"os/exec"

	"github.com/taigrr/elevenlabs/client"
	"github.com/taigrr/elevenlabs/client/types"
)

type AudioSource interface {
	Stream(hub *Hub) error
}

type FileSource struct {
	Path string
}

func (f *FileSource) Stream(hub *Hub) error {
	cmd := exec.Command("ffmpeg", "-re", "-i", f.Path, "-vn", "-f", "mp3", "-ab", "192k", "pipe:1")

	// For debugging
	// cmd.Stderr = os.Stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	return pipeToHub(stdout, hub)
}

type SpeechSource struct {
	Text    string
	VoiceID string
}

func (s *SpeechSource) Stream(hub *Hub) error {
	ctx := context.Background()

	client := client.New(os.Getenv("XI_API_KEY"))

	pipeReader, pipeWriter := io.Pipe()

	go func() {
		err := client.TTSStream(ctx, pipeWriter, s.Text, s.VoiceID, types.SynthesisOptions{Stability: 0.5, SimilarityBoost: 0.75}, func(t *types.TTS) {
			t.ModelID = "eleven_v3"
		})
		if err != nil {
			pipeWriter.CloseWithError(err)
		}
		pipeWriter.Close()
	}()

	return pipeToHub(pipeReader, hub)

	/*body, _ := json.Marshal(map[string]any{
		"text":     s.Text,
		"model_id": "eleven_v3",
		"voice_settings": map[string]any{
			"stability":        0.5,
			"similarity_boost": 0.75,
			//"apply_text_normalization": "off",
		},
	})

	req, _ := http.NewRequest("POST",
		fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s/stream", s.VoiceID),
		bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("xi-api-key", os.Getenv("XI_API_KEY"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ElevenLabs error %d: %s", resp.StatusCode, string(body))
	}*/

	//return pipeToHub(resp.Body, hub)
}

// Reads from any io.Reader and broadcasts chunks
func pipeToHub(r io.Reader, hub *Hub) error {
	buf := make([]byte, 4096)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			chunk := make([]byte, n)
			copy(chunk, buf[:n])
			hub.Broadcast(chunk)
		}
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}
