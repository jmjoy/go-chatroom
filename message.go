package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"

	simplejson "github.com/bitly/go-simplejson"
	"github.com/pborman/uuid"
	"golang.org/x/net/websocket"
)

var users = make(map[string]*User)

type User struct {
	Conn     *websocket.Conn
	UserName string
}

var messageHandler = MessageHandler{}

type MessageHandler struct{}

func (MessageHandler) Open(ws *websocket.Conn, json *simplejson.Json) {
	userName, err := json.Get("user_name").String()
	if err != nil {
		log.Println(err)
		return
	}

	token := uuid.New()
	users[token] = &User{
		Conn:     ws,
		UserName: userName,
	}

	renderJson(ws, map[string]string{
		"token": token,
	})
}

func handleMessage(ws *websocket.Conn) {
	scanner := bufio.NewScanner(ws)
	for scanner.Scan() {
		buf := scanner.Bytes()
		fmt.Println(string(buf))

		json, err := simplejson.NewJson(buf)
		if err != nil {
			log.Println(err)
			continue
		}
		t, err := json.Get("type").String()
		if err != nil {
			log.Println(err)
			continue
		}

		io.Copy(ws, ws)

		switch t {
		case "open":
			messageHandler.Open(ws, json)

		default:
			renderJson(ws, map[string]string{
				"msg": "unknow action",
			})
		}
	}
}

func renderJson(w io.Writer, data interface{}) {
	buf, err := json.Marshal(data)
	if err != nil {
		log.Println(err)
		return
	}
	w.Write(buf)
}
