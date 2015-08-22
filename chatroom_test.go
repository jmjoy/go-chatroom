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

	for i := 0; i < 10; i++ {
		if _, err := ws.Write([]byte(`{"type":"null","name":"jmjoy"}`)); err != nil {
			t.Fatal(err)
		}

		var msg = make([]byte, 512)
		var n int
		if n, err = ws.Read(msg); err != nil {
			t.Fatal(err)
		}

		t.Logf("Received: %s.\n", msg[:n])
	}
}
