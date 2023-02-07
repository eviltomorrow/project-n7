package telegrambot

import (
	"fmt"
	"strings"

	"github.com/eviltomorrow/project-n7/app/n7-telegram-bot/handler/db"
	"github.com/eviltomorrow/project-n7/lib/zlog"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

func reply(bot *tgbotapi.BotAPI, u tgbotapi.Update) error {
	if u.Message == nil || u.Message.From == nil {
		return fmt.Errorf("u.Message is nil")
	}
	var (
		username = u.Message.From.UserName
		chatID   = u.Message.Chat.ID
		text     = u.Message.Text
		// textID   = u.Message.MessageID
	)
	zlog.Info("User text", zap.String("text", text), zap.String("username", username))

	if username != "eviltomorrow" {
		return fmt.Errorf("not public service")
	}
	text = strings.TrimSpace(text)

	switch text {
	case "/start":
		msg := tgbotapi.NewMessage(chatID, `这是一个私人消息提醒机器人助理(自用), [提醒]请输入"/订阅", [取消提醒]请输入"/取消"`)
		if _, err := bot.Send(msg); err != nil {
			return err
		}
		lib.Set(&db.Session{Username: username, ChatID: chatID, Status: Subscribe})

	case "/订阅":
		msg := tgbotapi.NewMessage(chatID, `订阅成功`)
		if _, err := bot.Send(msg); err != nil {
			return err
		}
		lib.Set(&db.Session{Username: username, ChatID: chatID, Status: Subscribe})

	case "/取消":
		msg := tgbotapi.NewMessage(chatID, `取消成功`)
		if _, err := bot.Send(msg); err != nil {
			return err
		}
		lib.Set(&db.Session{Username: username, ChatID: chatID, Status: Unsubscribe})

	default:

	}

	return nil
}

func Send(bot *Bot, username string, text string) error {
	session, ok := lib.Get(username)
	if !ok {
		return fmt.Errorf("not found session with [%s]", username)
	}
	if session.Status == Unsubscribe {
		return fmt.Errorf("not subscribe")
	}

	msg := tgbotapi.NewMessage(session.ChatID, text)
	if _, err := bot.bot.Send(msg); err != nil {
		return err
	}
	return nil
}
