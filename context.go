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

func (this ContextPool) SendAll(resp *Response) {
	//gMsgPool.PushBack(resp *Response)
	//if gMsgPool.Len() > 10 {
	//gMsgPool.Remove(gMsgPool.Front())
	//}

	for c := range this {
		c.Send(resp)
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
	ChanMsg chan *Response

	UserName string
}

func NewContext(conn *websocket.Conn) *Context {
	this := &Context{
		Conn:    conn,
		ChanMsg: make(chan *Response),
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

func (this *Context) Send(resp *Response) {
	this.ChanMsg <- resp
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
	//gContexts.SendAll(map[string]interface{}{
	//"type":      "auth",
	//"message":   userName + "进入聊天室",
	//"userNames": userNames,
	//})
}

func (this *Context) Message(req *Request) {
	gContexts.SendAll(map[string]interface{}{
		"type":     "message",
		"userName": this.UserName,
		"content":  content,
	})
}
