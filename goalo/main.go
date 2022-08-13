package main

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	tele "gopkg.in/telebot.v3"
)

const (
	HelpMessage = `
		/info
		/account
		/setup
	`
	InfoMessage = `
		apps:
			android: bit.ly/3rECytR
			ios: apple.co/3rIOCtW
			windows: bit.ly/3bDZ8gI
			unix: install openconnect

			tips:
				- use batch mode
				- enable no-dtls (disable UDP)

		--------------------------------------------------------------------------------
		vpn host: %s
	`
	AccountMessage = `
		credit: %0.2f
		card number: %s
	`
	PayMessage = `
		please pay %0.2f Tomans. type in or press /accounts for more info
	`
)

type Config struct {
	AdminID         int64
	BotToken        string
	PollingInterval int64
	DB              string
	VPNHost         string
	CardNumber      string
}

func initConfig() Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()

	if err != nil {
		log.Panic(err)
		panic(err)
	}

	c := Config{}
	viper.Unmarshal(&c)
	return c
}

func initBot(config *Config) *tele.Bot {
	pref := tele.Settings{
		Token:  config.BotToken,
		Poller: &tele.LongPoller{Timeout: time.Duration(config.PollingInterval) * time.Second},
	}

	bot, err := tele.NewBot(pref)
	if err != nil {
		log.Error(err)
		panic(err)
	}
	return bot

}

func initDB(config *Config) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(config.DB))
	if err != nil {
		log.Panic(err)
		panic(err)
	}
	db.AutoMigrate(&User{})
	return db
}

func InitCron(bot *tele.Bot, db *gorm.DB) {
	c := cron.New()
	c.AddFunc("0 18 * * *", func() {
		users := []User{}
		db.Find(&users)
		for _, u := range users {
			if debtDays := time.Duration(time.Now().Unix()-u.LastChargeUnix) / time.Duration(24*time.Hour); u.LastChargeUnix != 0 && u.Credit < 0 && debtDays > 7 {
				err := u.Deactivate()
				if err != nil {
					log.Error(err)
				}
			} else if debtDays <= 7 {
				user, err := bot.ChatByID(u.TeleID)
				if err != nil {
					log.Error(err)
				}
				bot.Send(user, fmt.Sprintf(PayMessage, -1*u.Credit))
			}
		}
	})
}

type User struct {
	gorm.Model
	TeleID         int64 `gorm:primaryKey`
	Credit         float64
	LastChargeUnix int64
	VPNUsername    string
	VPNPassword    string
}

func (u *User) Pay(amount float64, db *gorm.DB) error {
	u.Credit += amount
	db.Save(u)
	if u.Credit >= 0 {
		return u.Activate()
	}
	return errors.New("user has negative cerdit")
}

func (u *User) Activate() error {
	cmd := exec.Command("ocpasswd", u.VPNUsername, "-c", "/etc/ocserv/passwd")
	return cmd.Err
}

func (u *User) Deactivate() error {
	cmd := exec.Command("ocpasswd", "-d", u.VPNUsername, "-c", "/etc/ocserv/passwd")
	return cmd.Err

}

func main() {
	config := initConfig()
	bot := initBot(&config)
	db := initDB(&config)
	InitCron(bot, db)

	onHelp := func(c tele.Context) error {
		c.Send(HelpMessage)
		return nil
	}
	bot.Handle("/start", onHelp)
	bot.Handle("/help", onHelp)

	onInfo := func(c tele.Context) error {
		c.Send(fmt.Sprintf(InfoMessage, config.VPNHost))
		return nil
	}
	bot.Handle("/info", onInfo)

	onAccount := func(c tele.Context) error {
		user := User{}
		if err := db.Where("tele_id = ?", c.Sender().ID).First(&user).Error; err != nil {
			db.Create(&User{TeleID: c.Sender().ID})
		}
		c.Send(fmt.Sprintf(AccountMessage, user.Credit, config.CardNumber))
		return nil
	}
	bot.Handle("/account", onAccount)

	onPhoto := func(c tele.Context) error {
		admin, err := bot.ChatByID(config.AdminID)
		if err != nil {
			log.WithFields(log.Fields{
				"event":  "photo",
				"sender": c.Sender().ID,
			}).Error("failed to get admin by ChatByID")
			return err
		}
		c.ForwardTo(admin)
		bot.Send(admin, fmt.Sprintf("%d", c.Sender().ID))
		log.WithFields(log.Fields{
			"event":  "photo",
			"sender": c.Sender().ID,
		}).Info("photo & userID sent to admin successfully")
		return nil
	}
	bot.Handle(tele.OnPhoto, onPhoto)

	onText := func(c tele.Context) error {
		if admin := c.Sender(); c.Message().ReplyTo != nil && admin.ID == config.AdminID {
			user := User{}
			if err := db.Where("tele_id = ?", c.Message().ReplyTo.Text).First(&user).Error; err != nil {
				log.Error(err)
			}
			amount, err := strconv.ParseFloat(c.Message().Text, 64)
			if err != nil {
				log.Error(err)
			}
			err = user.Pay(amount, db)
			if err != nil {
				log.WithFields(log.Fields{
					"event":  "payment",
					"user":   user.TeleID,
					"amount": c.Message().Text,
					"credit": user.Credit,
				}).Error("payment failed")
				return err
			}
			log.WithFields(log.Fields{
				"event":  "payment",
				"user":   user.TeleID,
				"amount": c.Message().Text,
				"credit": user.Credit,
			}).Info("payed successfully")
		}
		return nil
	}
	bot.Handle(tele.OnText, onText)

	onCharge := func(c tele.Context) error {
		if admin := c.Sender(); admin.ID == config.AdminID {
			totalFee, err := strconv.ParseFloat(strings.Split(c.Message().Text, ",")[1], 64)
			if err != nil {
				log.Error(err)
			}
			users := []User{}
			db.Find(&users)
			userFee, err := strconv.ParseFloat(fmt.Sprintf("%0.2f", totalFee/float64(len(users))), 64)
			if err != nil {
				log.Error(err)
			}
			for _, u := range users {
				u.Pay(-1*userFee, db)
			}
		}
		return nil
	}
	bot.Handle("/charge", onCharge)

	bot.Start()
}
