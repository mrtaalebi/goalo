package user

import (
	"errors"
	"fmt"
	"goalo/goalo/repo"
	"os/exec"
	"time"

	"gorm.io/gorm"

	log "github.com/sirupsen/logrus"
)

type User struct {
	gorm.Model
	TeleID         int64 `gorm:primaryKey`
	Credit         float64
	LastChargeUnix int64
	VPNUsername    string
	VPNPassword    string
}

func MigrateUser() {
	repo.DB.AutoMigrate(&User{})
}

func (u *User) Pay(amount float64) error {
	u.Credit += amount
	repo.DB.Save(u)
	recipient, _ := repo.Bot.ChatByID(u.TeleID)
	repo.Bot.Send(recipient, fmt.Sprintf("You have successfully paid %0.2f Tomans. Your credit is: %0.2f Tomans.", amount, u.Credit))
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

func (u *User) CheckCredit() {
	debtDays := int64(time.Duration(time.Now().Unix()-u.LastChargeUnix) / time.Duration(24*time.Hour))
	if u.LastChargeUnix != 0 && u.Credit < 0 && debtDays > 7 {
		err := u.Deactivate()
		if err != nil {
			log.Error(err)
		}
	} else if debtDays <= 7 {
		user, err := repo.Bot.ChatByID(u.TeleID)
		if err != nil {
			log.Error(err)
		}
		repo.Bot.Send(user, fmt.Sprintf(repo.Config.Message.Pay, -1*u.Credit))
	}
}
