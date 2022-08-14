package repo

import (
	"time"

	tele "gopkg.in/telebot.v3"

	log "github.com/sirupsen/logrus"
)

var Bot *tele.Bot = InitBot()

func InitBot() *tele.Bot {
	pref := tele.Settings{
		Token:  Config.BotToken,
		Poller: &tele.LongPoller{Timeout: time.Duration(Config.PollingInterval) * time.Second},
	}

	bot, err := tele.NewBot(pref)
	if err != nil {
		log.Error(err)
		panic(err)
	}
	return bot

}
