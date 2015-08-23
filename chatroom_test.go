package main_test

import (
	"testing"

	"golang.org/x/net/websocket"
)

var chReceiveExit = make(chan struct{})

func TestHandleMessage(t *testing.T) {
	origin := "http://localhost/"
	url := "ws://localhost:9000/message"

	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		websocket.JSON.Send(ws, map[string]interface{}{
			"type": "auth",
			"data": map[string]interface{}{
				"userName": "Tester",
			},
		})

		websocket.JSON.Send(ws, map[string]interface{}{
			"type": "message",
			"data": map[string]interface{}{
				"content": "fuck you",
			},
		})

		ws.Close()
	}()

	for {
		var v interface{}
		err := websocket.JSON.Receive(ws, &v)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(v)
	}

}
