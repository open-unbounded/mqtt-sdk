package client

import (
	"fmt"
)

type (
	Header  map[string][]string
	Message struct {
		Header   Header
		Topic    string
		Qos      byte
		Retained bool
		Payload  []byte
	}
)

func (msg *Message) String() string {
	return fmt.Sprintf("Message{Topic:%s, Qos:%d, Retained:%v, Payload:%s}", msg.Topic, msg.Qos, msg.Retained, string(msg.Payload))
}

func (h Header) Values(key string) []string {
	return h[key]
}

func (h Header) Get(key string) string {
	values, ok := h[key]
	if !ok {
		return ""
	}

	if len(values) == 0 {
		return ""
	}

	return values[0]
}

func (h Header) Set(key string, values ...string) {
	h[key] = values
}

func (h Header) Append(key string, values ...string) {
	h[key] = append(h[key], values...)
}

func (h Header) Clone() Header {
	header := Header{}
	for k, v := range h {
		header[k] = v
	}

	return header
}
