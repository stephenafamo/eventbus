package eventbus

type Handler[Payload any] interface {
	Handle(payload Payload)
}

type HandlerFunc[Payload any] func(payload Payload)

func (e HandlerFunc[Payload]) Handle(payload Payload) {
	e(payload)
}
