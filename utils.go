package main

import (
	"encoding/base64"
	"io"
	"math"
	"net/http"
)

func getImageAsBase64(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	imageBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	base64Image := base64.StdEncoding.EncodeToString(imageBytes)
	return base64Image, nil
}

func ErrorPayload() map[string]any {

	ErrorFeedbackPayload := map[string]interface{}{

		"title":      "YTMD Remote",
		"value":      "Connecting...",
		"play-icon":  nil,
		"like-icon":  nil,
		"cover-icon": nil,
		"indicator": map[string]interface{}{
			"value":   0,
			"enabled": false,
		},
	}
	return ErrorFeedbackPayload
}

func RetrieveSongInfo(recall map[string]any) (map[string]any, error) {

	s, err := ytmdGetTrack(ytmdHost)
	if err != nil {
		return ErrorPayload(), err

	}

	y, err := ytmdGetPlayer(ytmdHost)
	if err != nil {
		return ErrorPayload(), err
	}

	artistName := s.Author
	titleName := s.Title

	playImg := "images/controller-stop"

	if titleName != "" {
		switch y.IsPaused {
		case true:
			playImg = "images/controller-paus"
		case false:
			playImg = "images/controller-play"
		default:
			playImg = "images/controller-stop"
		}
	}

	likeImg := ""
	switch likeImg = y.LikeStatus; likeImg {

	case "LIKE":
		likeImg = "images/heart"
	case "DISLIKE":
		likeImg = "images/thumbs-down"
	default:
		likeImg = "images/heart-outlined"
	}

	progressInt := int(math.Round(y.StatePercent * 100))

	indicator := map[string]interface{}{
		"value":   progressInt,
		"enabled": true,
	}

	coverImg := ""

	if s.Title != recall["PreviousSong"] && s.Cover != "" {
		coverImgBase64, err := getImageAsBase64(s.Cover)
		if err != nil {
			return ErrorPayload(), err
		}
		coverImg = "data:image/png;base64," + coverImgBase64
	}

	SetFeedbackPayload := map[string]interface{}{

		"title":      artistName,
		"value":      titleName,
		"play-icon":  playImg,
		"like-icon":  likeImg,
		"indicator":  indicator,
		"cover-icon": coverImg,
	}

	return SetFeedbackPayload, err

}
