package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"

	simplejson "github.com/bitly/go-simplejson"
	"golang.org/x/net/websocket"
)

var gAllConns = make(ConnUserMap)

type ConnUserMap map[*websocket.Conn]*User

func (this ConnUserMap) GetAllUserNames() []string {
	userNames := make([]string, 0, 4)
	for _, user := range gAllConns {
		userNames = append(userNames, user.UserName)
	}
	return userNames
}

func (this ConnUserMap) SendAll(data interface{}) {
	for conn := range this {
		renderJson(conn, data)
	}
}

type User struct {
	UserName string
}

type MessageHandler struct {
	Conn *websocket.Conn
	Json *simplejson.Json
}

func (this *MessageHandler) Open() {
	_, ok := gAllConns[this.Conn]
	if ok {
		return
	}

	userName, err := this.Json.Get("data").Get("userName").String()
	if err != nil {
		log.Println(err)
		return
	}

	gAllConns[this.Conn] = &User{
		UserName: userName,
	}

	userNames := gAllConns.GetAllUserNames()
	gAllConns.SendAll(map[string]interface{}{
		"type":      "open",
		"message":   userName + "进入聊天室",
		"userNames": userNames,
	})
}

func (this *MessageHandler) Close() {
	userName := gAllConns[this.Conn].UserName
	delete(gAllConns, this.Conn)
	this.Conn.Close()

	userNames := gAllConns.GetAllUserNames()
	gAllConns.SendAll(map[string]interface{}{
		"type":      "close",
		"message":   userName + "离开聊天室",
		"userNames": userNames,
	})
}

func (this *MessageHandler) GetName() {
	user, ok := gAllConns[this.Conn]
	if !ok {
		this.RenderJson(map[string]interface{}{
			"userName": nil,
		})
		return
	}

	this.RenderJson(map[string]interface{}{
		"type":     "getName",
		"userName": user.UserName,
	})
}

func (this *MessageHandler) SendMsg() {
	content, err := this.Json.Get("data").Get("content").String()
	if err != nil {
		log.Println(err)
		return
	}

	gAllConns.SendAll(map[string]interface{}{
		"type":     "sendMsg",
		"userName": gAllConns[this.Conn].UserName,
		"content":  content,
	})
}

func (this *MessageHandler) RenderJson(data interface{}) {
	renderJson(this.Conn, data)
}

func handleMessage(conn *websocket.Conn) {
	msgHandler := &MessageHandler{Conn: conn}
	buffer := new(bytes.Buffer)
	char := make([]byte, 1)

LOOP:
	for {
		buffer.Reset()
		for {
			_, err := conn.Read(char)
			if err != nil {
				if err != io.EOF {
					log.Println(err)
				}
				// client has close
				msgHandler.Close()
				break LOOP
			}
			if bytes.Equal(char, []byte("\n")) {
				break
			}
			buffer.Write(char)
		}

		buf := buffer.Bytes()

		// parse body json
		json, err := simplejson.NewJson(buf)
		if err != nil {
			log.Println(err)
			continue
		}
		msgHandler.Json = json

		// use type to forward handler
		t, err := json.Get("type").String()
		if err != nil {
			log.Println(err)
			continue
		}
		switch t {
		case "open":
			msgHandler.Open()

		case "getName":
			go msgHandler.GetName()

		case "sendMsg":
			go msgHandler.SendMsg()
		}
	}
}

func renderJson(conn *websocket.Conn, data interface{}) {
	buf, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	conn.Write(buf)
}
