package main_test

import (
	"testing"

	"golang.org/x/net/websocket"
)

var chReceiveExit = make(chan struct{})

func _TestHandleMessage(t *testing.T) {
	origin := "http://localhost/"
	url := "ws://localhost:10001/ws"

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

func TestProtocol(t *testing.T) {
	origin := "http://localhost/"
	url := "ws://localhost:10001/ws"

	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		t.Fatal(err)
	}

	websocket.Message.Send(ws, "00000031\nuserName=__Tester__&type=auth\n")
	readConn(t, ws)

	//websocket.Message.Send(ws, "00000047\ntype=message\na<script>alert('hello');</script>")
	//readConn(t, ws)

	websocket.Message.Send(ws, "00000021\ntype=image\n123456789")
	readConn(t, ws)

	ws.Close()
}

func readConn(t *testing.T, ws *websocket.Conn) {
	buf := make([]byte, 4096)
	n, err := ws.Read(buf)
	if err != nil {
		t.Error(err)
	}
	t.Log(string(buf[:n]))
}
