package main

import (
	"github.com/robfig/cron"
	"goalo/pkg/handlers"
	"goalo/pkg/repo"
	"goalo/pkg/user"
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

	handlers.AddHandlers()

	repo.Bot.Start()
}
