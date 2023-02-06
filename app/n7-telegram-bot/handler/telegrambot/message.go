package telegrambot

import (
	"fmt"
	"strings"

	"github.com/eviltomorrow/project-n7/lib/zlog"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

func reply(bot *tgbotapi.BotAPI, u tgbotapi.Update) error {
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
	case "/start", "订阅":
		msg := tgbotapi.NewMessage(chatID, `这是一个私人使用得股票提醒机器人，不构成任何投资建议，取消提醒请输入"取消订阅"`)
		if _, err := bot.Send(msg); err != nil {
			return err
		}
		lib.Set(&Session{Username: username, ChatId: chatID, Status: Subscribe})

	case "取消订阅":
		msg := tgbotapi.NewMessage(chatID, `取消成功`)
		if _, err := bot.Send(msg); err != nil {
			return err
		}
		lib.Set(&Session{Username: username, ChatId: chatID, Status: Unsubscribe})

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

	msg := tgbotapi.NewMessage(session.ChatId, text)
	if _, err := bot.bot.Send(msg); err != nil {
		return err
	}
	return nil
}
