package client

import (
	"fmt"
)

type Message struct {
	Topic    string
	Qos      byte
	Retained bool
	Payload  interface{}
}

func (msg *Message) String() string {
	return fmt.Sprintf("Message{Topic:%s, Qos:%d, Retained:%v, Payload:%v}", msg.Topic, msg.Qos, msg.Retained, msg.Payload)
}
