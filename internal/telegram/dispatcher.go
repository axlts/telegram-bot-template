package telegram

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"sync"

	"github.com/axlts/telegram-bot-template/internal/config"
	"github.com/mymmrac/telego"
	"github.com/valyala/fasthttp"
)

const defaultBufferSize = 100

// dispatcher abstracts the source of Telegram updates.
type dispatcher interface {
	Updates(context.Context) (<-chan telego.Update, error)
	Shutdown(context.Context) error
}

type pollingDispatcher struct {
	bot    *telego.Bot
	cfg    config.Polling
	cancel context.CancelFunc
}

var _ dispatcher = (*pollingDispatcher)(nil)

func newPollingDispatcher(bot *telego.Bot, cfg config.Polling) dispatcher {
	return &pollingDispatcher{bot: bot, cfg: cfg}
}

func (d *pollingDispatcher) Updates(ctx context.Context) (<-chan telego.Update, error) {
	if exists, err := hasWebhook(ctx, d.bot); err != nil {
		return nil, err
	} else if exists {
		if err = d.bot.DeleteWebhook(ctx, nil); err != nil {
			return nil, err
		}
	}

	params := &telego.GetUpdatesParams{
		Offset:  d.cfg.Offset,
		Limit:   d.cfg.Limit,
		Timeout: d.cfg.Timeout,
	}
	opts := []telego.LongPollingOption{
		telego.WithLongPollingBuffer(defaultBufferSize),
	}

	ctx, d.cancel = context.WithCancel(ctx)
	return d.bot.UpdatesViaLongPolling(ctx, params, opts...)
}

func hasWebhook(ctx context.Context, bot *telego.Bot) (bool, error) {
	info, err := bot.GetWebhookInfo(ctx)
	if err != nil {
		return false, err
	}
	return info.URL != "", err
}

func (d *pollingDispatcher) Shutdown(ctx context.Context) error {
	d.cancel() // closes the updates channel
	return nil
}

type webhookDispatcher struct {
	bot    *telego.Bot
	cfg    config.Webhook
	srv    *fasthttp.Server
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

var _ dispatcher = (*webhookDispatcher)(nil)

func newWebhookDispatcher(bot *telego.Bot, cfg config.Webhook) dispatcher {
	return &webhookDispatcher{bot: bot, cfg: cfg, srv: &fasthttp.Server{}}
}

func (d *webhookDispatcher) Updates(ctx context.Context) (<-chan telego.Update, error) {
	u, err := url.Parse(d.cfg.URL)
	if err != nil {
		return nil, err
	}

	// set webhook every time because you can
	// change the configuration between runs.
	params := &telego.SetWebhookParams{
		URL:                u.String(),
		MaxConnections:     d.cfg.MaxConn,
		DropPendingUpdates: d.cfg.DropPendingUpdates,
		SecretToken:        d.bot.SecretToken(),
	}
	if d.cfg.SSL.Enabled {
		cert, err := os.OpenFile(d.cfg.SSL.Cert, os.O_RDONLY, 0444)
		if err != nil {
			return nil, err
		}
		defer cert.Close() //nolint:errcheck

		params = params.WithCertificate(&telego.InputFile{File: cert})
	}
	if err = d.bot.SetWebhook(ctx, params); err != nil {
		return nil, err
	}

	regHandler := telego.WebhookFastHTTP(d.srv, u.Path)

	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		d.runServer()
	}()

	opts := []telego.WebhookOption{
		telego.WithWebhookBuffer(defaultBufferSize),
	}

	ctx, d.cancel = context.WithCancel(ctx)
	return d.bot.UpdatesViaWebhook(ctx, regHandler, opts...)
}

func (d *webhookDispatcher) runServer() {
	addr := fmt.Sprintf(":%d", d.cfg.Port)
	log.Printf("starting server on %s", addr)

	var err error
	if d.cfg.SSL.Enabled {
		err = d.srv.ListenAndServeTLS(addr, d.cfg.SSL.Cert, d.cfg.SSL.Key)
	} else {
		err = d.srv.ListenAndServe(addr)
	}
	if err != nil {
		log.Printf("server stopped with error: %v", err)
		return
	}
	log.Print("server stopped")
}

func (d *webhookDispatcher) Shutdown(ctx context.Context) error {
	d.cancel() // closes the updates channel
	err := d.srv.ShutdownWithContext(ctx)
	d.wg.Wait()
	return err
}
