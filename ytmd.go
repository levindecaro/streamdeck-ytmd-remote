package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func ytmdCmd(address string, token string, playerCmd string) error {
	// Define the JSON data you want to send
	data := map[string]interface{}{
		"command": playerCmd,
		"data":    "",
	}

	// Convert the JSON data to bytes
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	path := "/api/v1/command"
	url := "http://" + address + path

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return err
	}
	defer resp.Body.Close()
	return err
}
