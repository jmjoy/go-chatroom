package main

import (
	"io"
	"log"

	simplejson "github.com/bitly/go-simplejson"
	"golang.org/x/net/websocket"
)

func handleWebsocket(conn *websocket.Conn) {
	context := NewContext(conn)
	gContexts.Add(context)

	// handle panic and quit chatroom
	defer func() {
		context.Send(nil)
		gContexts.Remove(context)
		if context.HasAuth() {
			gContexts.SendAll(map[string]interface{}{
				"type":      "close",
				"message":   context.UserName + "离开聊天室",
				"userNames": gContexts.GetAllUserNames(),
			})
		}
		context = nil

		if err := recover(); err != nil {
			// if err is io.EOF, maye be beacause of client closing
			if err != io.EOF {
				log.Println(err)
			}
		}
	}()

	// send old message
	for e := gMsgPool.Front(); e != nil; e = e.Next() {
		context.Send(e.Value)
	}

	buf := make([]byte, 4096*1024*10)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			panic(err)
		}
		msg := buf[:n]

		// parse body json
		json, err := simplejson.NewJson(msg)
		if err != nil {
			panic(err)
		}

		// use type to forward handler
		typ, err := json.Get("type").String()
		if err != nil {
			panic(err)
		}
		data := json.Get("data")

		// auth operation
		if typ == "auth" {
			context.Auth(data)
			continue
		}

		// below operation need check auth
		if !context.HasAuth() {
			context.Send(map[string]interface{}{
				"type":    "error",
				"message": "no ahth",
			})
			continue
		}

		switch typ {
		case "message":
			context.Message(data)

		default:
			context.Send(map[string]interface{}{
				"type":    "error",
				"message": "unknow type",
			})
		}
	}
}
