package consumer

import (
	"context"

	"github.com/nats-io/nats.go/jetstream"
)

type MsgHandler struct {
}

func NewMsgHandler() *MsgHandler {
	return &MsgHandler{}
}

func (h *MsgHandler) Handle(ctx context.Context, msg jetstream.Msg) error {
	return nil
}
