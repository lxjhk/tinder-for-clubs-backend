package db

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"log"
	"tinder-for-clubs-backend/common"
	"tinder-for-clubs-backend/config"
)

// Global DB connection instance
var DB *gorm.DB

func Init(dbCred config.DBCredential) {
	var err error
	log.Printf("Connection Info: %v", dbCred.ConnectionCredentialLogString())
	DB, err = gorm.Open("mysql", dbCred.GetConnectionString()) // QH Dev Env
	common.ErrFatalLog(err)

	// Checking connection status
	err = DB.DB().Ping()
	if err != nil {
		log.Fatalf("DB connection failed %s", err.Error())
	}
	log.Printf("DB connection established successfully!")
}

func Close() {
	err := DB.Close()
	common.ErrFatalLog(err)
}