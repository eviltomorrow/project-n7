package telegrambot

import (
	"fmt"
	"net/http"

	"github.com/eviltomorrow/project-n7/lib/zlog"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

type Bot struct {
	DomainName  string
	Pattern     string
	Port        int
	AccessToken string

	bot *tgbotapi.BotAPI
}

func (b *Bot) Run() error {
	bot, err := tgbotapi.NewBotAPI(b.AccessToken)
	if err != nil {
		return err
	}
	b.bot = bot

	hook, err := tgbotapi.NewWebhook(fmt.Sprintf("%s%s%s", b.DomainName, b.Pattern, b.AccessToken))
	if err != nil {
		return err
	}
	if _, err := bot.Request(hook); err != nil {
		return err
	}

	update := bot.ListenForWebhook(fmt.Sprintf("%s%s", b.Pattern, b.AccessToken))
	server := http.Server{
		Addr: fmt.Sprintf("0.0.0.0:%d", b.Port),
	}

	if err := lib.load(); err != nil {
		return err
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			zlog.Fatal("ListenAndServe on http failure", zap.Error(err))
		}
	}()
	go func() {
		for u := range update {
			go func(u tgbotapi.Update) {
				if err := reply(bot, u); err != nil {
					zlog.Error("Reply to failure", zap.Error(err), zap.String("username", u.Message.From.UserName))
				}
			}(u)
		}
	}()
	return nil
}

func (b *Bot) Stop() error {
	if b.bot != nil {
		b.bot.StopReceivingUpdates()
	}
	return nil
}
