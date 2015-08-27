package main

import (
	"errors"
	"io"
	"log"
	"strconv"

	"golang.org/x/net/websocket"
)

var (
	ErrPackageHeaderLength = errors.New("package header length error")
	ErrContentLength       = errors.New("content too long or empty")
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

	for {
		req, err := protocol(conn)
		if err != nil {
			// TODO handle error
			panic(err)
		}

		// auth operation
		if req.Type == "auth" {
			//context.Auth(req)
			continue
		}

		// below operation need check auth
		if !context.HasAuth() {
			//context.Send(map[string]interface{}{
			//"type":    "error",
			//"message": "no ahth",
			//})
			continue
		}

		switch req.Type {
		case "message":
			context.Message(req)

		default:
			context.Send(map[string]interface{}{
				"type":    "error",
				"message": "unknow type",
			})
		}
	}
}

func protocol(conn *websocket.Conn) (*Request, error) {
	// read content length
	buf := make([]byte, 8)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}
	if n != 8 {
		return nil, ErrPackageHeaderLength
	}

	length, err := strconv.Atoi(string(buf))
	if err != nil {
		return nil, ErrPackageHeaderLength
	}

	// limit to 50M
	if length < 1 || length > 50*1024*1024 {
		return nil, ErrContentLength
	}

	// read all content by length
	buf = make([]byte, length)
	n, err = conn.Read(buf)
	if err != nil {
		return nil, err
	}
	if n != length {
		return nil, ErrPackageHeaderLength
	}

	// parse request
	return ParseRequest(buf)
}
