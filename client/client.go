package client

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/cespare/xxhash/v2"
	"github.com/chenquan/orderhash"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/open-unbounded/mqtt-sdk/config"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/core/stat"
	"github.com/zeromicro/go-zero/core/threading"
	"github.com/zeromicro/go-zero/core/timex"
)

var metrics = stat.NewMetrics("pusher")

type (
	Pusher struct {
		clients       []*pusher
		group         *service.ServiceGroup
		size          int64
		hash64        func(b []byte) uint64
		routeHashFunc func([]byte) uint64
	}

	pusher struct {
		cli             mqtt.Client
		pushers         *threading.RoutineGroup
		channels        []chan *Message
		size            int64
		stopChan        chan struct{}
		doStopOnce      sync.Once
		metrics         *stat.Metrics
		pushDoneHandler func(message *Message, err error)
		hash64          func(b []byte) uint64
	}
)

func NewPusher(c config.PushConfig, opts ...PusherOption) *Pusher {
	if c.Conns < 1 {
		c.Conns = 1
	}
	if c.Pushers < 1 {
		c.Pushers = 8
	}

	op := new(pusherOptions)
	for _, opt := range opts {
		opt(op)
	}

	cli := &Pusher{group: service.NewServiceGroup(), hash64: orderhash.Hash64(xxhash.Sum64)}
	for i := 0; i < c.Conns; i++ {
		cli.clients = append(cli.clients, newPusher(c, c.ClientIdPrefix+strconv.Itoa(i), op))
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
	routeHashFunc := p.hash64
	if p.routeHashFunc != nil {
		routeHashFunc = p.routeHashFunc
	}

	hashCode := routeHashFunc([]byte(message.Topic))
	i := hashCode % uint64(len(p.clients))
	p.clients[i].Add(message)
}

func newPusher(conf config.PushConfig, clientID string, op *pusherOptions) *pusher {
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
		panic(connect.Error())
	}

	channels := make([]chan *Message, 0, conf.Pushers)
	for i := 0; i < conf.Pushers; i++ {
		channels = append(channels, make(chan *Message, 8))
	}

	c := &pusher{
		cli:             cli,
		channels:        channels,
		stopChan:        make(chan struct{}),
		pushers:         threading.NewRoutineGroup(),
		pushDoneHandler: op.pushDoneHandler,
		hash64:          orderhash.Hash64(xxhash.Sum64),
		metrics:         stat.NewMetrics(fmt.Sprintf("pusher(%s)", clientID)),
	}

	return c
}

func (p *pusher) startPusher() {
	for _, channel := range p.channels {
		channel := channel
		p.pushers.Run(
			func() {
				for msg := range channel {
					startTime := timex.Now()

					token := p.cli.Publish(msg.Topic, msg.Qos, msg.Retained, msg.Payload)
					task := stat.Task{}
					var err error
					if token.Wait() && token.Error() != nil {
						err = token.Error()
						logx.Errorw("数据发送失败", logx.Field("data", msg.String()), logx.Field("err", err))
						task.Drop = true
					}

					if p.pushDoneHandler != nil {
						p.pushDoneHandler(msg, err)
					}

					task.Duration = timex.Since(startTime)
					metrics.Add(task)
					p.metrics.Add(task)
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
	hashCode := p.hash64([]byte(message.Topic))
	i := hashCode % uint64(len(p.channels))
	p.channels[i] <- &message
}
