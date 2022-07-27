package client

import (
	"fmt"
)

type Message struct {
	Topic    string
	Qos      byte
	Retained bool
	Payload  []byte
}

func (msg *Message) String() string {
	return fmt.Sprintf("Message{Topic:%s, Qos:%d, Retained:%v, Payload:%s}", msg.Topic, msg.Qos, msg.Retained, string(msg.Payload))
}
