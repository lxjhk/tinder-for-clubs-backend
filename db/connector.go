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
	DB, err = gorm.Open("mysql", dbCred.GetConnectionString())
	common.ErrFatalLog(err)

	// Checking connection status
	err = DB.DB().Ping()
	DB.LogMode(true)

	if err != nil {
		log.Fatalf("DB connection failed %s", err.Error())
	}
	log.Printf("DB connection established successfully!")

	// Struct AutoMigrate
	err = DB.AutoMigrate(&AdminAccount{}).Error
	common.ErrFatalLog(err)
	DB.AutoMigrate(&LoginHistory{})
	common.ErrFatalLog(err)
	DB.AutoMigrate(&ClubInfo{})
	common.ErrFatalLog(err)
	DB.AutoMigrate(&UserList{})
	common.ErrFatalLog(err)
	DB.AutoMigrate(&ClubTags{})
	common.ErrFatalLog(err)
	DB.AutoMigrate(&ClubTagRelationship{})
	common.ErrFatalLog(err)

}

func Close() {
	err := DB.Close()
	common.ErrFatalLog(err)
}
