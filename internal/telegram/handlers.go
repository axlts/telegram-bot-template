package telegram

import (
	"context"
	"errors"

	"github.com/mymmrac/telego"
)

var (
	ErrUnknownCmd = errors.New("unknown command")
)

const (
	cmdStart = "start"
)

func (b *Bot) handleCmd(ctx context.Context, msg *telego.Message) {
	switch command(msg) {
	case cmdStart:
		b.send(ctx, msg.Chat.ChatID(), "Welcome! Send me a message and I’ll echo it back.")
	default:
		b.send(ctx, msg.Chat.ChatID(), ErrUnknownCmd.Error())
	}
}

func command(msg *telego.Message) string {
	if len(msg.Entities) == 0 {
		return ""
	}
	entity := msg.Entities[0]
	if len(msg.Text) < entity.Length {
		return ""
	}
	return msg.Text[1:entity.Length]
}

func (b *Bot) handleMsg(ctx context.Context, msg *telego.Message) {
	b.send(ctx, msg.Chat.ChatID(), msg.Text) // echo
}
