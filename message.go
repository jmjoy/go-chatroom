package main

import (
	"container/list"
	"io"
	"log"

	simplejson "github.com/bitly/go-simplejson"
	"golang.org/x/net/websocket"
)

var (
	gAllConns = make(ConnUserMap)
	gMsgPool  = list.New()
)

type ConnUserMap map[*websocket.Conn]*User

func (this ConnUserMap) GetAllUserNames() []string {
	userNames := make([]string, 0, len(this))
	for _, user := range this {
		userNames = append(userNames, user.UserName)
	}
	return userNames
}

func (this ConnUserMap) SendAll(data interface{}) {
	gMsgPool.PushBack(data)
	if gMsgPool.Len() > 10 {
		gMsgPool.Remove(gMsgPool.Front())
	}

	for conn := range this {
		websocket.JSON.Send(conn, data)
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

	// send old message
	for e := gMsgPool.Front(); e != nil; e = e.Next() {
		websocket.JSON.Send(this.Conn, e.Value)
	}

	// send message
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

func handleMessage(conn *websocket.Conn) {
	msgHandler := &MessageHandler{Conn: conn}
	buf := make([]byte, 4096)

	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				msgHandler.Close()
			} else {
				log.Println(err)
			}
			return
		}
		msg := buf[:n]

		// parse body json
		json, err := simplejson.NewJson(msg)
		if err != nil {
			log.Println(err)
			return
		}
		msgHandler.Json = json

		// use type to forward handler
		t, err := json.Get("type").String()
		if err != nil {
			log.Println(err)
			return
		}

		switch t {
		case "open":
			msgHandler.Open()

		case "sendMsg":
			go msgHandler.SendMsg()

		default:
			go conn.Write(msg)
		}
	}
}
