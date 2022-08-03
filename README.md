# mqtt-sdk

> a mqtt sdk.

## example

```go
package main

import (
	"strconv"

	"github.com/open-unbounded/mqtt-sdk/client"
	"github.com/open-unbounded/mqtt-sdk/config"
	"github.com/zeromicro/go-zero/core/logx"
)

func main() {
	err := logx.SetUp(logx.LogConf{
		ServiceName:         "mqtt",
		Mode:                "file",
		Encoding:            "plain",
		TimeFormat:          "",
		Path:                "logs",
		Level:               "info",
		Compress:            false,
		KeepDays:            0,
		StackCooldownMillis: 100,
	})

	if err != nil {
		panic(err)
	}

	pusher := client.NewPusher(config.PushConfig{
		Brokers:            []string{"tcp://localhost:1883"},
		Conns:              30,
		Pushers:            30,
		FileStoreDirPrefix: "data/mqtt-data-",
	})
	go func() {
		i := 0
		for {

			pusher.Add(client.Message{
				Topic:    strconv.Itoa(i % 10),
				Qos:      0,
				Retained: false,
				Payload:  []byte("111"),
			})
			i++
		}
	}()
	pusher.Start()
	defer pusher.Stop()
}

```