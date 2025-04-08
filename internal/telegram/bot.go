package telegram

import (
	"context"
	"log"
	"time"

	"github.com/axlts/telegram-bot-template/internal/config"
	"github.com/mymmrac/telego"
)

type Bot struct {
	bot *telego.Bot
	dsp dispatcher
}

func NewBot(cfg config.Bot) (*Bot, error) {
	var opts []telego.BotOption
	if cfg.Debug {
		opts = append(opts, telego.WithDefaultDebugLogger())
	}
	tb, err := telego.NewBot(cfg.Token, opts...)
	if err != nil {
		return nil, err
	}

	var dsp dispatcher
	if cfg.Mode == config.PollingMode {
		dsp = newPollingDispatcher(tb, cfg.Polling)
	} else {
		dsp = newWebhookDispatcher(tb, cfg.Webhook)
	}

	return &Bot{bot: tb, dsp: dsp}, nil
}

func (b *Bot) Run(ctx context.Context) error {
	updates, err := b.dsp.Updates(ctx)
	if err != nil {
		return err
	}

	for update := range updates {
		// ignore all non-message updates
		if update.Message == nil {
			continue
		}
		// ignore all non-text messages
		if update.Message.Text == "" {
			continue
		}

		if isCommand(update.Message) {
			b.handleCmd(ctx, update.Message)
		} else {
			b.handleMsg(ctx, update.Message)
		}
	}
	return nil
}

func isCommand(msg *telego.Message) bool {
	if len(msg.Entities) == 0 {
		return false
	}
	entity := msg.Entities[0]
	return entity.Offset == 0 && entity.Type == "bot_command"
}

func (b *Bot) Shutdown(ctx context.Context) error {
	return b.dsp.Shutdown(ctx)
}

func (b *Bot) send(ctx context.Context, chatID telego.ChatID, text string) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	params := &telego.SendMessageParams{ChatID: chatID, Text: text}
	if _, err := b.bot.SendMessage(ctx, params); err != nil {
		log.Printf("error sending message: %v", err)
	}
}
