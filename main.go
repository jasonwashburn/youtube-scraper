package main

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
)

// curl \
//   'https://youtube.googleapis.com/youtube/v3/videos?part=snippet&id=EjB1kz2tn5s&key=[YOUR_API_KEY]' \
//   --header 'Accept: application/json' \
//   --compressed
//

const apiUrl = "https://youtube.googleapis.com/youtube/v3/videos?part=snippet&id=%s&key=%s"

func main() {
	apiKey := os.Getenv("YOUTUBE_API_KEY")
	if apiKey == "" {
		log.Fatal("YOUTUBE_API_KEY environment variable is not set")
	}

	videoIds := []string{"EjB1kz2tn5s"}
	for _, id := range videoIds {
		tmpFileName := fmt.Sprintf("/tmp/%s.json", id)
		if !fileExists(tmpFileName) {
			videoInfo := getVideoInfo("EjB1kz2tn5s", apiKey)
			slog.Info("Fetching info for video: ", id)
			if err := os.WriteFile(tmpFileName, videoInfo, 0644); err != nil {
				slog.Warn("Unable to write file: ", tmpFileName)
			}
		}
		videoInfo, err := os.ReadFile(tmpFileName)
		if err != nil {
			slog.Warn("Unable to read file: ", tmpFileName)
		}

		fmt.Println(string(videoInfo))
	}
}

func getVideoInfo(id string, apiKey string) []byte {
	url := fmt.Sprintf(apiUrl, id, apiKey)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("Failed to get video info: ", err)
	}

	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Failed to read response body: ", err)
	}

	return body
}

func processVideoInfo(s string) {
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return true
}
