package main

import (
	"container/list"
	"encoding/json"
	"errors"

	"golang.org/x/net/websocket"
)

var (
	gContexts = make(ContextPool)
	gMsgPool  = list.New()
)

type ContextPool map[*Context]struct{}

func (this ContextPool) SendAll(resp *Response) {
	gMsgPool.PushBack(resp)
	if gMsgPool.Len() > 10 {
		gMsgPool.Remove(gMsgPool.Front())
	}

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
	ChanResponse chan *Response

	UserName string
}

func NewContext(conn *websocket.Conn) *Context {
	this := &Context{
		Conn:         conn,
		ChanResponse: make(chan *Response),
	}

	go func() {
		for {
			resp := <-this.ChanResponse
			// close this connect
			if resp == nil {
				close(this.ChanResponse)
				this.Conn.Close()
				this = nil
				return
			}
			buf, err := resp.EncodeBytes()
			if err != nil {
				logError(err)
				continue
			}
			this.Conn.Write(buf)
		}
	}()

	return this
}

func (this *Context) Send(resp *Response) {
	this.ChanResponse <- resp
}

func (this *Context) HasAuth() bool {
	return this != nil && this.UserName != ""
}

func (this *Context) Auth(req *Request) error {
	userName := req.Values.Get("userName")
	if userName == "" {
		return errors.New("用户名不能为空")
	}
	this.UserName = userName

	// send message
	allUsers, err := json.Marshal(gContexts.GetAllUserNames())
	if err != nil {
		return err
	}

	gContexts.SendAll(NewResponse("join", allUsers, "message", userName+"进入聊天室"))
	return nil
}

func (this *Context) Message(req *Request) error {
	gContexts.SendAll(NewResponse("message", req.Body, "userName", this.UserName))
	return nil
}

func (this *Context) Leave() error {
	allUsers, err := json.Marshal(gContexts.GetAllUserNames())
	if err != nil {
		return err
	}

	gContexts.SendAll(NewResponse("leave", allUsers, "message", this.UserName+"离开聊天室"))
	return nil
}
