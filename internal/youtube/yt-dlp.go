package youtube

import (
	"fmt"
	"os/exec"
	"strings"
)

func GetMp3FromUrl(url string) (string, error) {
	var cmd []byte
	var err error
	if cmd, err = exec.Command("yt-dlp", "--extract-audio", "--audio-format", "mp3", url).Output(); err != nil {
		return "", fmt.Errorf("yt-dlp failed: %w", err)
	}
	audio := strings.TrimSpace(string(cmd))

	return audio, nil
}
