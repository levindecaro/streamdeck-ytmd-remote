package main

import (
	"encoding/base64"
	"io"
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

func DefaultPayload() map[string]any {

	ErrorFeedbackPayload := map[string]interface{}{

		"title":     "YTMD Remote",
		"value":     "Waiting for data...",
		"play-icon": nil,
		"like-icon": nil,
		"indicator": map[string]interface{}{
			"value":   0,
			"enabled": false,
		},
	}
	return ErrorFeedbackPayload
}
