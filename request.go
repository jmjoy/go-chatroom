package main

import (
	"bytes"
	"errors"
	"net/url"
)

var (
	ErrRequestFormat = errors.New("request content format error")
)

type Request struct {
	ContentLength int
	Type          string
	url.Values
	Body []byte
}

func ParseRequest(buf []byte) (*Request, error) {
	// the first letter must be '\n'
	if buf[0] != '\n' {
		return nil, ErrRequestFormat
	}

	// parse Values
	index := bytes.IndexByte(buf[1:], '\n')
	index++
	if index < 5 {
		return nil, ErrRequestFormat
	}

	values, err := url.ParseQuery(string(buf[1:index]))
	if err != nil {
		return nil, err
	}

	// get type
	typ := values.Get("type")
	if typ == "" {
		return nil, ErrRequestFormat
	}

	// get body reader
	var body []byte
	if index < len(buf)-1 {
		body = buf[index+1:]
	}

	req := &Request{
		ContentLength: len(buf),
		Type:          typ,
		Values:        values,
		Body:          body,
	}

	return req, nil
}
