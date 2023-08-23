package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/levindecaro/streamdeck"
)

var (
	PauseInfoRetrival bool // Use to pause song retrieval between encoder action
	StopInfoRetrival  bool // Use to exit song retrieval loop
)

var (
	ytmdHost  string
	ytmdToken string
	ytmdPort  string
)

type Settings struct {
	Address string `json:"address"`
	Token   string `json:"token"`
	Port    string `json:"port"`
}

func main() {
	ctx := context.Background()

	if err := run(ctx); err != nil {
		log.Fatalf("%v\n", err)
	}
}

func run(ctx context.Context) error {
	params, err := streamdeck.ParseRegistrationParams(os.Args)
	if err != nil {
		return err
	}

	client := streamdeck.NewClient(ctx, params)
	setup(client)

	return client.Run()
}

func setup(client *streamdeck.Client) {
	action := client.Action("com.ytmd.remote.encoder")

	settings := make(map[string]*Settings)

	action.RegisterHandler(streamdeck.DidReceiveSettings, func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
		p := streamdeck.DidReceiveSettingsPayload{}

		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return err
		}
		s, ok := settings[event.Context]
		if !ok {
			s = &Settings{}
			settings[event.Context] = s
		}

		if err := json.Unmarshal(p.Settings, s); err != nil {
			return err
		}

		ytmdHost = s.Address + ":" + s.Port
		ytmdToken = s.Token
		StopInfoRetrival = false
		client.LogMessage("YTMD Address: " + ytmdHost + ", YTMD Token: " + ytmdToken)

		return nil
	})

	action.RegisterHandler(streamdeck.TouchTap, func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {

		p := streamdeck.TouchTapPayload{}

		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return err
		}

		y, err := ytmdGetPlayer(ytmdHost)
		if err != nil {
			client.SetFeedback(ctx, ErrorPayload(), streamdeck.HardwareAndSoftware)
		}

		PauseInfoRetrival = true

		showTitle := ""

		if p.Hold {

			if y.LikeStatus == "DISLIKE" {
				showTitle = "Unset Dislike"
			} else {
				showTitle = "Dislike"
			}

			err := ytmdCmd(ytmdHost, ytmdToken, "track-thumbs-down")
			if err != nil {
				client.LogMessage(err.Error())
			}

		} else {

			if y.LikeStatus == "LIKE" {
				showTitle = "Unset Like"
			} else {
				showTitle = "Like"
			}

			err := ytmdCmd(ytmdHost, ytmdToken, "track-thumbs-up")
			if err != nil {
				client.LogMessage(err.Error())
			}
		}

		return client.SetFeedbackValue(ctx, showTitle, streamdeck.HardwareAndSoftware)

	})

	action.RegisterHandler(streamdeck.DialDown, func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {

		p := streamdeck.DialDownPayload{}

		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return err
		}

		y, err := ytmdGetPlayer(ytmdHost)
		if err != nil {
			return err
		}

		SetFeedbackPayload := map[string]interface{}{
			"play-icon": "images/controller-paus",
		}

		if y.IsPaused {
			ytmdCmd(ytmdHost, ytmdToken, "track-play")
		} else {
			ytmdCmd(ytmdHost, ytmdToken, "track-pause")
		}

		return client.SetFeedback(ctx, SetFeedbackPayload, streamdeck.HardwareAndSoftware)

	})

	action.RegisterHandler(streamdeck.DialRotate, func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {

		p := streamdeck.DialRotateEventPayload{}

		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return err
		}

		s, ok := settings[event.Context]
		if !ok {
			s = &Settings{}
			settings[event.Context] = s
		}

		if err := json.Unmarshal(p.Settings, s); err != nil {
			return err
		}

		showImg := ""
		showValue := ""
		PauseInfoRetrival = true
		if p.Ticks > 0 { // Rotated CW
			err := ytmdCmd(ytmdHost, ytmdToken, "track-next")
			if err != nil {
				client.SetFeedback(ctx, ErrorPayload(), streamdeck.HardwareAndSoftware)
				client.LogMessage(err.Error())
			}

			showImg = "images/controller-next"
			showValue = "Next Track"

		} else { // Rotated CCW
			err := ytmdCmd(ytmdHost, ytmdToken, "track-previous")
			if err != nil {
				client.SetFeedback(ctx, ErrorPayload(), streamdeck.HardwareAndSoftware)
				client.LogMessage(err.Error())
			}
			showImg = "images/controller-jump-to-start"
			showValue = "Previous Track"
		}

		SetFeedbackPayload := map[string]interface{}{
			"title":     "Loading..",
			"value":     showValue,
			"play-icon": showImg,
			"like-icon": "images/heart-outlined",
			"indicator": map[string]interface{}{
				"value":   0,
				"enabled": false,
			},
		}

		return client.SetFeedback(ctx, SetFeedbackPayload, streamdeck.HardwareAndSoftware)

	})

	action.RegisterHandler(streamdeck.WillDisappear, func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
		StopInfoRetrival = true
		return nil
	})

	action.RegisterHandler(streamdeck.WillAppear, func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {

		p := streamdeck.WillAppearPayload{}

		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return err
		}

		s, ok := settings[event.Context]
		if !ok {
			s = &Settings{}
			settings[event.Context] = s
		}

		if err := json.Unmarshal(p.Settings, s); err != nil {
			return err
		}

		client.LogMessage("YTMD Remote Started")

		ytmdHost = s.Address + ":" + s.Port
		ytmdToken = s.Token

		ticker := time.NewTicker((1000 * time.Millisecond))
		stop := make(chan bool)

		var PreviousSong string

		StopInfoRetrival = false

		go func() {
			for {
				select {
				case <-stop:
					return
				case <-ticker.C:

					recall := map[string]any{
						"PreviousSong": PreviousSong, // Pass Previous Song Name to RetrieveSongInfo, skip cover image download if unchanged.
					}
					if !PauseInfoRetrival { // Skip calling RetrieveSongInfo if pause flag set

						SetFeedbackPayload, err := RetrieveSongInfo(recall)
						if err != nil {

							client.SetFeedback(ctx, SetFeedbackPayload, streamdeck.HardwareAndSoftware)
							client.LogMessage(err.Error())
							time.Sleep(3 * time.Second) // Delay retry
							break
						}
						PreviousSong = SetFeedbackPayload["value"].(string) // Store Previous Song Name

						client.SetFeedback(ctx, SetFeedbackPayload, streamdeck.HardwareAndSoftware)
					}
					PauseInfoRetrival = false // reset for next iterration

				}

				if StopInfoRetrival { // Exit the loop when plugin unload
					stop <- true
				}
			}
		}()
		return (client.SetFeedbackTitle(ctx, "YTMD Remote", streamdeck.HardwareAndSoftware))
	})
}
