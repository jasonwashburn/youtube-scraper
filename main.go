package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
)

// To get video info
// curl \
//   'https://youtube.googleapis.com/youtube/v3/videos?part=snippet&id=EjB1kz2tn5s&key=[YOUR_API_KEY]' \
//   --header 'Accept: application/json' \
//   --compressed
//

type Video struct {
	Season      int           `json:"season"`
	Episode     int           `json:"epidsode"`
	Id          string        `json:"id"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Url         string        `json:"url"`
	Thumbnail   ThumbnailInfo `json:"thumbnail"`
}

type Results struct {
	Videos []Video `json:"videos"`
}

const (
	videosApiUrl        = "https://youtube.googleapis.com/youtube/v3/videos?part=snippet&id=%s&key=%s"
	playlistItemsApiUrl = "https://youtube.googleapis.com/youtube/v3/playlistItems?part=id,snippet,contentDetails&maxResults=15&playlistId=%s&key=%s"
)

func main() {
	apiKey := os.Getenv("YOUTUBE_API_KEY")
	if apiKey == "" {
		log.Fatal("YOUTUBE_API_KEY environment variable is not set")
	}

	playlistIds := map[int]string{
		1:  "ELPPAps9oEkaQ",
		2:  "EL-tj6jbdYkbM",
		3:  "ELPrKKsYysHyY",
		4:  "ELitzlFL46ACY",
		5:  "EL-i24C31QclY",
		6:  "ELVgKmZXK5fT4",
		7:  "ELm0sr4YhRjFI",
		8:  "ELCHHzvi8gZbU",
		9:  "EL-qDGiZtJHYc",
		10: "ELiabcfl0TJ-4UbSLCPc_p0w",
	}

	var results Results

	for season, playlistId := range playlistIds {
		tmpFileName := fmt.Sprintf("/tmp/playlist_%s.json", playlistId)
		if !fileExists(tmpFileName) {
			slog.Info("Fetching playlistItems for season", "season", season, "playlistId", playlistId)
			playlistItemResponse := getPlaylistItems(playlistId, apiKey)

			if err := os.WriteFile(tmpFileName, playlistItemResponse, 0644); err != nil {
				slog.Warn("Unable to write file: ", tmpFileName)
			}
		}

		playlistItemResponse, err := os.ReadFile(tmpFileName)
		if err != nil {
			slog.Warn("Unable to read file: ", tmpFileName)
		}

		playlistItems, err := processPlaylistItems(playlistItemResponse)
		if err != nil {
			slog.Warn("Unable to unmarshal playlistItemResponse", "playListId", playlistId, "error", err)
		}

		videos := playlistItemsToVideos(playlistItems, season)

		for _, video := range videos {
			results.Videos = append(results.Videos, video)
		}

	}
	output, err := json.Marshal(results)
	if err != nil {
		slog.Warn("Unable to marshal results int output")
	}

	if err := os.WriteFile("/tmp/mickey_episodes.json", output, 0644); err != nil {
		slog.Warn("Unable to write output file")
	}
	fmt.Println(string(output))
}

func getPlaylistItems(id string, apiKey string) []byte {
	url := fmt.Sprintf(playlistItemsApiUrl, id, apiKey)
	fmt.Println(url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("Failed to get playlistItems: ", err)
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

type Snippet struct {
	Title       string                   `json:"title"`
	Description string                   `json:"description"`
	Thumbnails  map[string]ThumbnailInfo `json:"thumbnails"`
}

type ThumbnailInfo struct {
	Url    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// To get playlistItems
// curl \
//   'https://youtube.googleapis.com/youtube/v3/playlistItems?part=id%2Csnippet%2CcontentDetails&maxResults=15&playlistId=ELPPAps9oEkaQ&key=[YOUR_API_KEY]' \
//   --header 'Authorization: Bearer [YOUR_ACCESS_TOKEN]' \
//   --header 'Accept: application/json' \
//   --compressed

type PlaylistItemResponse struct {
	Items []PlaylistItem `json:"items"`
}

type PlaylistItem struct {
	Snippet        Snippet        `json:"snippet"`
	ContentDetails ContentDetails `json:"contentDetails"`
}

type ContentDetails struct {
	VideoId string `json:"videoId"`
}

func processPlaylistItems(data []byte) ([]PlaylistItem, error) {
	var playlistItemResponse PlaylistItemResponse

	if err := json.Unmarshal(data, &playlistItemResponse); err != nil {
		return playlistItemResponse.Items, err
	}

	return playlistItemResponse.Items, nil
}

func playlistItemsToVideos(items []PlaylistItem, season int) []Video {
	var videos []Video
	for index, item := range items {
		video := Video{}
		video.Season = season
		video.Episode = index + 1
		video.Title = item.Snippet.Title
		video.Thumbnail = item.Snippet.Thumbnails["default"]
		video.Description = item.Snippet.Description
		video.Id = item.ContentDetails.VideoId
		video.Url = createUrlFromId(video.Id)
		videos = append(videos, video)
	}

	return videos
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func createUrlFromId(id string) string {
	return fmt.Sprintf("https://www.youtube.com/watch?v=%s", id)
}
