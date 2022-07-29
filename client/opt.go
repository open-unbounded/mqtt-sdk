package client

type (
	PusherOption  func(pusherOptions *pusherOptions)
	pusherOptions struct {
		pushDoneHandler func(message *Message, err error)
		routeHashFunc   func([]byte) uint64
	}
)

func WithPushDoneHandler(pushDoneHandler func(message *Message, err error)) PusherOption {
	return func(pusherOptions *pusherOptions) {
		pusherOptions.pushDoneHandler = pushDoneHandler
	}
}

func WithRouteHashFunc(routeHashFunc func(b []byte) uint64) PusherOption {
	return func(pusherOptions *pusherOptions) {
		pusherOptions.routeHashFunc = routeHashFunc
	}
}
