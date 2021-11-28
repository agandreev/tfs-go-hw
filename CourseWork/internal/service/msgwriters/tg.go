package msgwriters

import (
	"github.com/agandreev/tfs-go-hw/CourseWork/internal/domain"
	tb "gopkg.in/tucnak/telebot.v2"
	"time"
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
	tg.bot.Handle("/start", func(m *tb.Message) {
		tg.bot.Send(m.Sender, m.Sender.ID)
	})
	go tg.bot.Start()
}

func (tg TelegramBot) WriteMessage(message domain.Message, user domain.User) error {
	if _, err := tg.bot.Send(&tb.User{ID: int(user.TelegramID)}, message.String()); err != nil {
		return err
	}
	return nil
}

func (tg *TelegramBot) ShutDown() {
	tg.bot.Stop()
}