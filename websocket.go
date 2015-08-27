package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
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
		// if the user has auth, send leave message for all users
		if context.HasAuth() {
			err := context.Leave()
			if err != nil {
				logError(err)
			}
		}
		context = nil

		if err := recover(); err != nil {
			// if err is io.EOF, maye be beacause of client closing
			if err != io.EOF {
				logError(err)
			}
		}
	}()

	// send old message
	//for e := gMsgPool.Front(); e != nil; e = e.Next() {
	//context.Send(e.Value.(*Response))
	//}

	for {
		req, err := protocol(conn)
		if err != nil {
			// handle error
			context.Send(NewResponse("error", nil, "message", err.Error()))
			return
		}

		// service logic
		err = service(context, req)
		if err != nil {
			context.Send(NewResponse("error", nil, "message", err.Error()))
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

	fmt.Println("---" + string(buf) + "---")

	length, err := strconv.ParseInt(string(buf), 10, 32)
	if err != nil {
		return nil, ErrPackageHeaderLength
	}

	// limit to 50M
	if length < 1 || length > 50*1024*1024 {
		return nil, ErrContentLength
	}

	// read all content by length
	buffer := new(bytes.Buffer)
	i, err := io.CopyN(buffer, conn, length)

	if i != length {
		return nil, ErrPackageHeaderLength
	}

	fmt.Println("---" + string(buffer.Bytes()) + "---")

	// parse request
	return ParseRequest(buffer.Bytes())
}

func service(context *Context, req *Request) error {
	// auth operation
	if req.Type == "auth" {
		err := context.Auth(req)
		if err != nil {
			return err
		}
	}

	// below operation need check auth
	if !context.HasAuth() {
		return errors.New("no auth")
	}

	var err error
	switch req.Type {
	case "message":
		err = context.Message(req)

	default:
		return errors.New("unknow type")
	}

	// handle above error
	if err != nil {
		return errors.New("未知异常")
	}

	return nil
}
