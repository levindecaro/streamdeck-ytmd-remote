package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type TrackData struct {
	Author string `json:"author"`
	Album  string `json:"album"`
	Title  string `json:"title"`
	Cover  string `json:"cover"`
}

type PlayerData struct {
	HasSong      bool    `json:"hasSong"`
	IsPaused     bool    `json:"isPaused"`
	StatePercent float64 `json:"StatePercent"`
	LikeStatus   string  `json:"likeStatus"`
}

func ytmdGetTrack(address string) (*TrackData, error) {

	path := "/query/track"
	url := "http://" + address + path

	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Request failed with status:", resp.Status)
		return nil, err
	}

	var responseData TrackData
	err = json.NewDecoder(resp.Body).Decode(&responseData)
	if err != nil {
		fmt.Println("Error decoding JSON response:", err)
		return nil, err
	}

	return &responseData, err
}

func ytmdGetPlayer(address string) (*PlayerData, error) {

	path := "/query/player"
	url := "http://" + address + path

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Request failed with status:", resp.Status)
		return nil, err
	}

	var responseData PlayerData
	err = json.NewDecoder(resp.Body).Decode(&responseData)
	if err != nil {
		fmt.Println("Error decoding JSON response:", err)
		return nil, err
	}

	return &responseData, err
}

func ytmdCmd(address string, token string, playerCmd string) error {
	// Define the JSON data you want to send
	data := map[string]interface{}{
		"command": playerCmd,
	}

	// Convert the JSON data to bytes
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	path := "/query"
	url := "http://" + address + path

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "access_token "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return err
	}
	defer resp.Body.Close()
	return err
}
