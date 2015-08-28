package main

import (
	"bytes"
	"fmt"
	"net/url"
)

type Response struct {
	Type string
	url.Values
	Body []byte
}

func NewResponse(typ string, body []byte, valuePairs ...string) *Response {
	if typ == "" {
		panic("NewRequest: empty type")
	}
	if len(valuePairs)%2 != 0 {
		panic("NewRequest: num of valuePairs isnt't even")
	}

	values := make(url.Values)
	for i := 0; i < len(valuePairs); i += 2 {
		values.Set(valuePairs[i], valuePairs[i+1])
	}

	return &Response{
		Type:   typ,
		Values: values,
		Body:   body,
	}
}

func (this *Response) EncodeBytes() []byte {
	fmt.Println("resp:", this)

	var buffer bytes.Buffer

	this.Values.Add("type", this.Type)
	_, err := buffer.WriteString(this.Values.Encode() + "\n")
	if err != nil {
		return []byte(err.Error())
	}

	_, err = buffer.Write(this.Body)
	if err != nil {
		return []byte(err.Error())
	}

	return buffer.Bytes()
}
