package repo

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	log "github.com/sirupsen/logrus"
)

var DB *gorm.DB = InitDB()

func InitDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(Config.DB))
	if err != nil {
		log.Panic(err)
		panic(err)
	}
	return db
}
