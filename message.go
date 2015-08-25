package main

import (
	"container/list"
	"io"
	"log"

	simplejson "github.com/bitly/go-simplejson"
	"golang.org/x/net/websocket"
)

var (
	gContexts = make(ContextPool)
	gMsgPool  = list.New()
)

type ContextPool map[*Context]struct{}

func (this ContextPool) SendAll(msg interface{}) {
	gMsgPool.PushBack(msg)
	if gMsgPool.Len() > 10 {
		gMsgPool.Remove(gMsgPool.Front())
	}

	for c := range this {
		c.Send(msg)
	}
}

func (this ContextPool) GetAllUserNames() []string {
	userNames := make([]string, 0, len(this)/2)
	for c := range this {
		if c.HasAuth() {
			userNames = append(userNames, c.UserName)
		}
	}
	return userNames
}

func (this ContextPool) Add(c *Context) {
	this[c] = struct{}{}
}

func (this ContextPool) Remove(c *Context) {
	delete(this, c)
}

type Context struct {
	*websocket.Conn
	ChanMsg chan interface{}

	UserName string
}

func NewContext(conn *websocket.Conn) *Context {
	this := &Context{
		Conn:    conn,
		ChanMsg: make(chan interface{}),
	}

	go func() {
		for {
			msg := <-this.ChanMsg
			// close this connect
			if msg == nil {
				close(this.ChanMsg)
				this.Conn.Close()
				this = nil
				return
			}
			websocket.JSON.Send(this.Conn, msg)
		}
	}()

	return this
}

func (this *Context) Send(msg interface{}) {
	this.ChanMsg <- msg
}

func (this *Context) HasAuth() bool {
	return this != nil && this.UserName != ""
}

func (this *Context) Auth(data *simplejson.Json) {
	userName, err := data.Get("userName").String()
	if err != nil {
		log.Println(err)
		return
	}
	this.UserName = userName

	// send message
	userNames := gContexts.GetAllUserNames()
	gContexts.SendAll(map[string]interface{}{
		"type":      "auth",
		"message":   userName + "进入聊天室",
		"userNames": userNames,
	})
}

func (this *Context) Message(data *simplejson.Json) {
	content, err := data.Get("content").String()
	if err != nil {
		log.Println(err)
		return
	}

	gContexts.SendAll(map[string]interface{}{
		"type":     "message",
		"userName": this.UserName,
		"content":  content,
	})
}

func handleMessage(conn *websocket.Conn) {
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
