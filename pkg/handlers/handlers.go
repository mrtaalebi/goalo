package handlers

import (
	"fmt"
	"goalo/pkg/repo"
	"goalo/pkg/user"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	tele "gopkg.in/telebot.v3"
)

func AddHandlers() {
	repo.Bot.Handle("/start", onStart)
	repo.Bot.Handle("/apps", onApps)
	repo.Bot.Handle("/setup", onSetup)
	repo.Bot.Handle("/info", onInfo)
	repo.Bot.Handle("/pay", onPay)
	repo.Bot.Handle(tele.OnPhoto, onPhoto)

	// admin only
	repo.Bot.Handle(tele.OnText, onText)
	repo.Bot.Handle("/charge", onCharge)
}

func onStart(c tele.Context) error {
	c.Send(repo.Config.Message.Start)
	if err := repo.DB.Where("tele_id = ?", c.Sender().ID).First(&user.User{}).Error; err != nil {
		repo.DB.Create(&user.User{TeleID: c.Sender().ID})
	}
	return nil
}

func onApps(c tele.Context) error {
	c.Send(repo.Config.Message.Apps)
	return nil
}

func onSetup(c tele.Context) error {
	user := user.User{}
	repo.DB.Where("tele_id = ?", c.Sender().ID).First(&user)
	args := strings.Split(c.Message().Text, " ")[1:]
	if len(args) != 2 {
		if user.VPNUsername != "" && user.VPNPassword != "" {
			c.Send("You already have a VPN account. use /info")
		} else {
			c.Send(repo.Config.Message.Setup)
		}
		return nil
	}
	user.VPNUsername = args[0]
	user.VPNPassword = args[1]
	err := user.Activate()
	if err != nil {
		c.Send("Setting up your account failed. Please contact admin.")
		log.WithFields(log.Fields{
			"event": "vpn",
			"user":  c.Sender().ID,
		}).Error("activate failed")
		return err
	}
	repo.DB.Save(user)
	return nil
}

func onInfo(c tele.Context) error {
	user := user.User{}
	repo.DB.Where("tele_id = ?", c.Sender().ID).First(&user)
	if user.VPNUsername != "" && user.VPNPassword != "" {
		c.Send(fmt.Sprintf(repo.Config.Message.Info, repo.Config.VPNHost, user.VPNUsername, user.VPNPassword))
	} else {
		c.Send("You have yet to setup your VPN account. use /setup")
	}
	return nil
}

func onPay(c tele.Context) error {
	user := user.User{}
	repo.DB.Where("tele_id = ?", c.Sender().ID).First(&user)
	haveToPay := "Don't "
	if user.Credit < 0 {
		haveToPay = ""
	}
	c.Send(fmt.Sprintf(repo.Config.Message.Pay, user.Credit, haveToPay, repo.Config.CardNumber))
	return nil
}

func onPhoto(c tele.Context) error {
	admin, err := repo.Bot.ChatByID(repo.Config.AdminID)
	if err != nil {
		log.WithFields(log.Fields{
			"event": "photo",
			"user":  c.Sender().ID,
		}).Error("failed to get admin by ChatByID")
		return err
	}
	c.ForwardTo(admin)
	repo.Bot.Send(admin, fmt.Sprintf("%d", c.Sender().ID))
	log.WithFields(log.Fields{
		"event": "photo",
		"user":  c.Sender().ID,
	}).Info("photo & userID sent to admin successfully")
	return nil
}

func onText(c tele.Context) error {
	if admin := c.Sender(); c.Message().ReplyTo != nil && admin.ID == repo.Config.AdminID {
		user := user.User{}
		repo.DB.Where("tele_id = ?", c.Message().ReplyTo.Text).First(&user)
		amount, err := strconv.ParseFloat(c.Message().Text, 64)
		if err != nil {
			log.Error(err)
		}
		err = user.Pay(amount)
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

func onCharge(c tele.Context) error {
	if admin := c.Sender(); admin.ID == repo.Config.AdminID {
		totalFee, err := strconv.ParseFloat(strings.Split(c.Message().Text, ",")[1], 64)
		if err != nil {
			log.Error(err)
		}
		users := []user.User{}
		repo.DB.Find(&users)
		userFee, err := strconv.ParseFloat(fmt.Sprintf("%0.2f", totalFee/float64(len(users))), 64)
		if err != nil {
			log.Error(err)
		}
		for _, u := range users {
			u.Pay(-1 * userFee)
		}
	}
	return nil
}
