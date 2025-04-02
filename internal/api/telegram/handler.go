package telegram

import (
	"attune/internal/api"
	"attune/internal/models"
	"attune/internal/service"
	"attune/pkg/cache"
	"attune/pkg/logger"
	"context"
	"fmt"
	"time"

	tb "gopkg.in/telebot.v4"
)

var (
	ErrMissingToken               = "API: Token is required"
	ErrBotCreation                = "API: Failed to create bot"
	ErrSendMessage                = "failed to send message"
	ErrSendErrorMessage           = "failed to send error message"
	ErrSendErrorMessageWithMarkup = "failed to send error message with markup"
)

type sendErrorMsgParams struct {
	C      tb.Context
	ErrMsg string
	Markup *tb.ReplyMarkup
}

type APITelegramMessagesConfig struct {
	WelcomeMsg string
	FocusMsg   string
}

type API struct {
	token       string
	baseURL     string
	pollTimeout time.Duration
	bot         *tb.Bot
	services    service.Services
	logger      logger.Logger
	cache       cache.Cache
	apiCh       <-chan api.Trigger
}

func NewTelegramAPI(
	token, baseURL string,
	pollTimeout time.Duration,
	services service.Services,
	logger logger.Logger,
	cache cache.Cache,
	apiCh <-chan api.Trigger,
) api.ExternalAPI {
	if token == "" {
		panic(ErrMissingToken)
	}
	if pollTimeout == 0 {
		pollTimeout = 10 * time.Second
	}

	pref := tb.Settings{
		Token:  token,
		Poller: &tb.LongPoller{Timeout: pollTimeout},
	}
	bot, err := tb.NewBot(pref)
	if err != nil {
		panic(fmt.Errorf("%s: %w", ErrBotCreation, err))
	}

	return &API{
		token:       token,
		baseURL:     baseURL,
		bot:         bot,
		pollTimeout: pollTimeout,
		services:    services,
		logger:      logger,
		cache:       cache,
		apiCh:       apiCh,
	}
}

func (a *API) Start(ctx context.Context) error {
	registerStartCommand(a)
	a.registerFocusSessionCallbacks()

	go a.bot.Start()
	go a.ListenTriggers(ctx)

	<-ctx.Done()
	a.bot.Stop()

	return ctx.Err()
}

func (a *API) Trigger(ctx context.Context, vendorID string, trigger api.Trigger) error {
	switch trigger.Type {
	case api.TriggerTypeFinishSession:
		return a.finishFocusSession(ctx, vendorID, trigger.FocusSessionStatus)
	}
	return nil
}

func (a *API) ListenTriggers(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case trig, ok := <-a.apiCh:
			if !ok {
				return
			}
			vendorID := trig.VendorID
			if err := a.Trigger(ctx, vendorID, trig); err != nil {
				a.logger.Error(ctx, "failed to trigger", err)
			}
		}
	}
}

func (a *API) SendMessage(_ context.Context, message models.Message) error {
	recipient := &tb.User{Username: message.VendorID}
	_, err := a.bot.Send(recipient, message.Text)
	if err != nil {
		return fmt.Errorf("%s: %w", ErrSendMessage, err)
	}
	return nil
}

func (a *API) sendErrorMsg(_ context.Context, input sendErrorMsgParams) {
	opts := &tb.SendOptions{ParseMode: tb.ModeMarkdown}
	if _, err := a.bot.Send(input.C.Sender(), input.ErrMsg, opts); err != nil {
		return
	}
	if input.Markup != nil {
		if _, err := a.bot.Send(input.C.Sender(), "Please try again.", input.Markup); err != nil {
			return
		}
	}
}
