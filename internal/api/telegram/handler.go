package telegram

import (
	"attune/internal/api"
	"attune/internal/models"
	"attune/internal/service"
	"attune/pkg/cache"
	"attune/pkg/logger"
	"context"
	"log"
	"time"

	tb "gopkg.in/telebot.v4"
)

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
}

func NewTelegramAPI(
	token, baseURL string,
	pollTimeout time.Duration,
	services service.Services,
	logger logger.Logger,
	cache cache.Cache,
) api.ExternalAPI {
	if token == "" {
		log.Fatal("API: Token is required")
	}
	//if cfg.BaseURL == "" {
	//	log.Fatal("API: BaseURL is required")
	//}
	if pollTimeout == 0 {
		pollTimeout = 10 * time.Second
	}

	pref := tb.Settings{
		Token:  token,
		Poller: &tb.LongPoller{Timeout: pollTimeout},
	}
	bot, err := tb.NewBot(pref)
	if err != nil {
		log.Fatalf("API: Failed to create bot: %v", err)
	}

	return &API{
		token:       token,
		baseURL:     baseURL,
		bot:         bot,
		pollTimeout: pollTimeout,
		services:    services,
		logger:      logger,
		cache:       cache,
	}
}

func (a *API) Start(ctx context.Context) error {
	RegisterStartCommand(a)

	go a.bot.Start()

	<-ctx.Done()
	a.bot.Stop()

	a.logger.Info(ctx, "Telegram bot stopped")

	return ctx.Err()
}

func (a *API) SendMessage(ctx context.Context, message models.Message) error {
	const op = "telegramAPI.SendMessage"

	log := a.logger.With("operation", op)

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	recipient := &tb.User{Username: message.VendorID}
	_, err := a.bot.Send(recipient, message.Text)
	if err != nil {
		log.Error(ctx, "failed to send message", err)
		return err
	}

	return nil
}
