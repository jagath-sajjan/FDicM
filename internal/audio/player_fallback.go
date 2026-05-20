//go:build cross

package audio

import (
	"os/exec"
	"runtime"
)

func PlayAudioFromURL(url string, fallbackText string) {
	executeTTSFallback(fallbackText)
}

func executeTTSFallback(text string) {
	go func() {
		switch runtime.GOOS {
		case "darwin":
			_ = exec.Command("say", text).Run()
		case "linux":
			_ = exec.Command("spd-say", text).Run()
		case "windows":
			_ = exec.Command("powershell", "-Command", "Add-Type -AssemblyName System.Speech; (New-Object System.Speech.Synthesis.SpeechSynthesizer).Speak('"+text+"')").Run()
		}
	}()
}
