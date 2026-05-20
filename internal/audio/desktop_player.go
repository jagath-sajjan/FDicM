//go:build !cross

package audio

import (
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
)

func PlayAudioFromURL(url string, fallbackText string) {
	if url != "" {
		go func() {
			res, err := http.Get(url)
			if err != nil {
				executeTTSFallback(fallbackText)
				return
			}
			defer res.Body.Close()

			streamer, format, err := mp3.Decode(res.Body)
			if err != nil {
				executeTTSFallback(fallbackText)
				return
			}
			defer streamer.Close()

			_ = speaker.Init(format.SampleRate, format.SampleRate.N(time.Millisecond*200))
			done := make(chan bool)
			speaker.Play(beep.Seq(streamer, beep.Callback(func() {
				done <- true
			})))
			<-done
		}()
		return
	}
	executeTTSFallback(fallbackText)
}

func executeTTSFallback(text string) {
	go func() {
		switch runtime.GOOS {
		case "darwin":
			_ = exec.Command("say", text).Run()
		case "linux":
			_ = exec.Command("spd-say", text).Run()
		}
	}()
}
