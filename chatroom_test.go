package main_test

import (
	"testing"

	"golang.org/x/net/websocket"
)

func TestHandleMessage(t *testing.T) {
	origin := "http://localhost/"
	url := "ws://localhost:9000/message"

	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		t.Fatal(err)
	}
	defer ws.Close()

	go func() {
		for {
			var v interface{}
			websocket.JSON.Receive(ws, &v)
			t.Log(v)
		}
	}()

	websocket.JSON.Send(ws, map[string]interface{}{
		"type": "open",
		"data": map[string]interface{}{
			"userName": "Tester",
		},
	})

	websocket.JSON.Send(ws, map[string]interface{}{
		"type": "sendMsg",
		"data": map[string]interface{}{
			"content": "fuck you",
		},
	})
}
