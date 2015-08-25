package main

import (
	"container/list"
	"log"

	"github.com/bitly/go-simplejson"
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
