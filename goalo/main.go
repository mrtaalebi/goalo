package main

import (
	"goalo/goalo/repo"
	"goalo/goalo/user"

	"github.com/robfig/cron"
)

func AddCron() {
	c := cron.New()
	c.AddFunc("0 18 * * *", func() {
		users := []user.User{}
		repo.DB.Find(&users)
		for _, u := range users {
			u.CheckCredit()
		}
	})
}

func main() {
	user.MigrateUser()
	AddCron()

	AddHandlers()
	repo.Bot.Start()
}
