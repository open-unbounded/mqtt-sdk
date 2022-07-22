package client

import (
	"strconv"
	"sync"
	"sync/atomic"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/open-unbounded/mqtt-sdk/config"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/core/threading"
)

type (
	Message struct {
		Topic    string
		Qos      byte
		Retained bool
		Payload  interface{}
	}

	Pusher struct {
		clients []*pusher
		group   *service.ServiceGroup
		size    int64
	}

	pusher struct {
		cli        mqtt.Client
		pushers    *threading.RoutineGroup
		channels   []chan *Message
		size       int64
		stopChan   chan struct{}
		doStopOnce sync.Once
	}
)

func NewPusher(c config.PushConfig) *Pusher {
	if c.Conns < 1 {
		c.Conns = 1
	}
	if c.Pushers < 1 {
		c.Pushers = 8
	}

	cli := &Pusher{group: service.NewServiceGroup()}
	for i := 0; i < c.Conns; i++ {
		cli.clients = append(cli.clients, newClient(c, c.ClientIdPrefix+strconv.Itoa(i)))
	}

	return cli
}

func (p *Pusher) Start() {
	for _, cli := range p.clients {
		p.group.Add(cli)
	}
	p.group.Start()
}

func (p *Pusher) Stop() {
	p.group.Stop()
	_ = logx.Close()
}

func (p *Pusher) Add(message Message) {
	i := atomic.AddInt64(&p.size, 1) % int64(len(p.clients))
	p.clients[i].Add(message)
}

func newClient(conf config.PushConfig, clientID string) *pusher {
	options := mqtt.NewClientOptions()
	for _, broker := range conf.Brokers {
		options.AddBroker(broker)
	}

	options.SetClientID(clientID)
	options.SetUsername(conf.Username)
	options.SetPassword(conf.Password)
	options.SetAutoReconnect(true)      //启用自动重连功能
	options.SetMaxReconnectInterval(30) //每30秒尝试重连
	options.Store = mqtt.NewFileStore(conf.FileStoreDirPrefix + clientID)
	cli := mqtt.NewClient(options)
	if connect := cli.Connect(); connect.Wait() && connect.Error() != nil {
		panic(connect.Wait())
	}

	channels := make([]chan *Message, 0, conf.Pushers)
	for i := 0; i < conf.Pushers; i++ {
		channels = append(channels, make(chan *Message, 8))
	}

	c := &pusher{
		cli:      cli,
		channels: channels,
		stopChan: make(chan struct{}),
		pushers:  threading.NewRoutineGroup(),
	}

	return c
}

func (p *pusher) startPusher() {
	for _, channel := range p.channels {
		channel := channel
		p.pushers.Run(
			func() {
				for msg := range channel {
					token := p.cli.Publish(msg.Topic, msg.Qos, msg.Retained, msg.Payload)
					if token.Wait() && token.Error() != nil {
						logx.Error(token.Error())
					}
				}
			},
		)
	}
}

func (p *pusher) Start() {
	p.startPusher()
	select {
	case <-p.stopChan:
		return
	}
}

func (p *pusher) Stop() {
	p.doStopOnce.Do(func() {
		for _, channel := range p.channels {
			close(channel)
		}
		p.pushers.Wait()
		close(p.stopChan)
	})
}

func (p *pusher) Add(message Message) {
	i := atomic.AddInt64(&p.size, 1) % int64(len(p.channels))
	p.channels[i] <- &message
}
