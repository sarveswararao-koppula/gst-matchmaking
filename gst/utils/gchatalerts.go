package utils

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type Message struct {
	Text string `json:"text"`
}

func Gchatalert(message string,gst string) {
	webhookURL := "https://chat.googleapis.com/v1/spaces/AAAA1Csmi8U/messages?key=AIzaSyDdI0hCZtE6vySjMm-WEfRq3CPzqKqqsHI&token=mcWLBHZnwkwbm9wuptfKjcJx3rRqAv5AXrPQ3kkOEhU"
	alertMessage := "Gst:"+ gst + "Error:" + message
	// if message != nil {
	// 	alertMessage = message.Error()
	// }

	messageBody := Message{
		Text: alertMessage,
	}

	payloadBytes, err := json.Marshal(messageBody)
	if err != nil {
			panic(err)
	}
	payloadBody := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("POST", webhookURL, payloadBody)
	if err != nil {
			panic(err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
			panic(err)
	}
	defer resp.Body.Close()
}