package service

import (
	"fmt"
	"strconv"
	"time"

	"github.com/agandreev/tfs-go-hw/CourseWork/internal/domain"
	tb "gopkg.in/tucnak/telebot.v2"
)

type TelegramBot struct {
	bot *tb.Bot
}

func NewTelegramBot(token string) (*TelegramBot, error) {
	bot, err := tb.NewBot(tb.Settings{
		Token:       token,
		Poller:      &tb.LongPoller{Timeout: 10 * time.Second},
		Synchronous: false,
	})
	if err != nil {
		return nil, err
	}
	tg := &TelegramBot{bot: bot}
	tg.InitRoutes()
	return tg, nil
}

func (tg *TelegramBot) InitRoutes() {
	tg.bot.Handle("/id", func(m *tb.Message) {
		id := strconv.FormatInt(int64(m.Sender.ID), 10)
		_, err := tg.bot.Send(m.Sender, id)
		if err != nil {
			_, _ = tg.bot.Send(m.Sender, err)
			return
		}
	})
	go tg.bot.Start()
}

func (tg TelegramBot) WriteMessage(message domain.OrderInfo, user domain.User) error {
	if user.TelegramID == 0 {
		return fmt.Errorf("user hasn't telegram id")
	}
	if _, err := tg.bot.Send(&tb.User{ID: int(user.TelegramID)}, message.String()); err != nil {
		return err
	}
	return nil
}

func (tg TelegramBot) WriteError(message string, user domain.User) error {
	fmt.Println(user)
	if _, err := tg.bot.Send(&tb.User{ID: int(user.TelegramID)}, message); err != nil {
		return fmt.Errorf("can't send message to tg user: <%w>", err)
	}
	return nil
}

func (tg TelegramBot) Shutdown() {
	tg.bot.Stop()
}
