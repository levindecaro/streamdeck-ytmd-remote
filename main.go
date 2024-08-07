package main

import (
	"context"
	"encoding/json"
	"log"
	"math"
	"os"
	"strings"
	"time"

	"github.com/LiterMC/socket.io"
	"github.com/LiterMC/socket.io/engine.io"
	"github.com/levindecaro/streamdeck"
)

var isActivated bool

var (
	LikeStatusInfo    int
	coverImg          string
	PreviousThumbnail string
)

type Payload struct {
	Player Player `json:"player"`
	Video  Video  `json:"video"`
}

type Player struct {
	Queue         Queue   `json:"queue"`
	VideoProgress float64 `json:"videoProgress"`
	Volume        int     `json:"volume"`
	TrackState    int     `json:"trackState"`
}

type PlayerState struct {
	Player Player `json:"player"`
}

type Queue struct {
	Items     []Item `json:"items"`
	ItemIndex int    `json:"selectedItemIndex"`
	Infinite  bool   `json:"isInfinite"`
	AutoPlay  bool   `json:"autoplay"`
}

type Item struct {
	Selected   bool        `json:"selected"`
	Title      string      `json:"title"`
	Author     string      `json:"author"`
	Duration   string      `json:"duration"`
	Thumbnails []Thumbnail `json:"thumbnails"`
}

type Thumbnail struct {
	URL string `json:"url"`
}

type Video struct {
	Title           string      `json:"title"`
	Author          string      `json:"author"`
	Album           string      `json:"album"`
	DurationSeconds int         `json:"durationSeconds"`
	Thumbnails      []Thumbnail `json:"thumbnails"`
	LikeStatus      int         `json:"likeStatus"`
}

var (
	ytmdHost  string
	ytmdToken string
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
		client.LogMessage("YTMD Address: " + ytmdHost + ", YTMD Token: " + ytmdToken)

		return nil
	})

	action.RegisterHandler(streamdeck.TouchTap, func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {

		p := streamdeck.TouchTapPayload{}

		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return err
		}

		showTitle := ""

		if p.Hold {

			if LikeStatusInfo == -1 {
				showTitle = "Unset Dislike"
			} else {
				showTitle = "Dislike"
			}

			err := ytmdCmd(ytmdHost, ytmdToken, "toggleDisLike")
			if err != nil {
				client.LogMessage(err.Error())
			}

		} else {

			if LikeStatusInfo == 1 {
				showTitle = "Unset Like"
			} else {
				showTitle = "Like"
			}

			err := ytmdCmd(ytmdHost, ytmdToken, "toggleLike")
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

		SetFeedbackPayload := map[string]interface{}{
			"play-icon": "images/controller-paus",
		}

		ytmdCmd(ytmdHost, ytmdToken, "playPause")

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
		if p.Ticks > 0 { // Rotated CW
			err := ytmdCmd(ytmdHost, ytmdToken, "next")
			if err != nil {
				client.SetFeedback(ctx, DefaultPayload(), streamdeck.HardwareAndSoftware)
				client.LogMessage(err.Error())
			}

			showImg = "images/controller-next"
			showValue = "Next Track"

		} else { // Rotated CCW
			err := ytmdCmd(ytmdHost, ytmdToken, "previous")
			if err != nil {
				client.SetFeedback(ctx, DefaultPayload(), streamdeck.HardwareAndSoftware)
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

		client.LogMessage("YTMD Remote Start")

		ytmdHost = s.Address + ":" + s.Port
		ytmdToken = s.Token

		opts := engine.Options{
			Secure: false,
			Host:   "ws://" + ytmdHost,
			Path:   "/socket.io/",
		}

		messageChan := make(chan []byte, 1)

		engio, err := engine.NewSocket(opts)
		if err != nil {
			client.LogMessage("Failed to create a new socket: " + err.Error())
		}

		if !isActivated {
			isActivated = true

			engio.OnMessage(func(s *engine.Socket, data []byte) {
				select {
				case messageChan <- data:
				default:
					<-messageChan
					messageChan <- data
				}
			})
			ticker := time.NewTicker(1000 * time.Millisecond)
			go func() {
				for range ticker.C {
					select {
					case data := <-messageChan:
						messageProcesser(data, ctx, client, event)
					default:
						// No message to process
					}
				}
			}()
			err = engio.Dial(context.Background())
			if err != nil {
				client.LogMessage("Failed to connect to the server: " + err.Error())
			}

			sio := socket.NewSocket(engio, socket.WithAuthToken(ytmdToken))

			for {
				client.LogMessage("trying to connect")
				if sio.Status() == 0 {
					sio.Connect("/api/v1/realtime")
					break
				} else if sio.Status() == 1 {
					sio.Emit("state-update")
					break
				}
				time.Sleep(time.Second)
			}
		}
		return (client.SetFeedback(ctx, DefaultPayload(), streamdeck.HardwareAndSoftware))
	})

}

func messageProcesser(data []byte, ctx context.Context, client *streamdeck.Client, event streamdeck.Event) {

	t := string(data)
	startIdx := strings.Index(t, `{`)
	endIdx := strings.LastIndex(t, `}`)
	jsonData := t[startIdx : endIdx+1]

	var result map[string]interface{}
	err := json.Unmarshal([]byte(jsonData), &result)
	if err != nil {
		client.LogMessage("Error unmarshaling JSON: " + err.Error())
	}
	if _, ok := result["sid"]; ok {
		return

	}
	var message Payload
	if err := json.Unmarshal([]byte(jsonData), &message); err != nil {
		client.LogMessage("Error parsing JSON: " + err.Error())
		return
	}
	floatSecond := float64(message.Video.DurationSeconds)
	progressInt := int(math.Round(message.Player.VideoProgress / floatSecond * 100))
	songIndex := message.Player.Queue.ItemIndex
	indicator := map[string]interface{}{
		"value":   progressInt,
		"enabled": true,
	}

	likeImg := "images/heart-outlined"
	LikeStatusInfo = message.Video.LikeStatus
	if LikeStatusInfo == 2 {
		likeImg = "images/thumbs-heart"
	}
	if LikeStatusInfo == 0 {
		likeImg = "images/thumbs-down"
	}

	playImg := "images/controller-stop"
	if message.Player.TrackState == 1 {
		playImg = "images/controller-play"
	}
	currentThumbnail := message.Video.Thumbnails[0].URL
	if PreviousThumbnail != currentThumbnail {
		coverImgBase64, err := getImageAsBase64(currentThumbnail)
		coverImg = "data:image/png;base64," + coverImgBase64
		if err != nil {
			client.LogMessage("Failed to download thumbnail from " + currentThumbnail + "Error: " + err.Error())
		}
	}

	SetFeedbackPayload := map[string]interface{}{
		"title":      message.Player.Queue.Items[songIndex].Author,
		"value":      message.Player.Queue.Items[songIndex].Title,
		"indicator":  indicator,
		"cover-icon": coverImg,
		"like-icon":  likeImg,
		"play-icon":  playImg,
	}
	// client.LogMessage(message.Player.Queue.Items[songIndex].Title)
	client.SetFeedback(ctx, SetFeedbackPayload, streamdeck.HardwareAndSoftware)
	PreviousThumbnail = currentThumbnail
}
