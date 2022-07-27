package client

type (
	PusherOption  func(pusherOptions *pusherOptions)
	pusherOptions struct {
		pushDoneHandler func(message *Message, err error)
	}
)

func WithPushDoneHandler(pushDoneHandler func(message *Message, err error)) PusherOption {
	return func(pusherOptions *pusherOptions) {
		pusherOptions.pushDoneHandler = pushDoneHandler
	}
}
