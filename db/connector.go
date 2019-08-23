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
	DB.SingularTable(true)

	if err != nil {
		log.Fatalf("DB connection failed %s", err.Error())
	}
	log.Printf("DB connection established successfully!")

	// Struct AutoMigrate
	err = DB.AutoMigrate(&AdminAccount{}).Error
	common.ErrFatalLog(err)
	err = DB.AutoMigrate(&LoginHistory{}).Error
	common.ErrFatalLog(err)
	err = DB.AutoMigrate(&ClubInfo{}).Error
	common.ErrFatalLog(err)
	err = DB.AutoMigrate(&UserList{}).Error
	common.ErrFatalLog(err)
	err = DB.AutoMigrate(&ClubTags{}).Error
	common.ErrFatalLog(err)
	err = DB.AutoMigrate(&ClubTagRelationship{}).Error
	common.ErrFatalLog(err)
	err = DB.AutoMigrate(&AccountPicture{}).Error
	common.ErrFatalLog(err)
}

func Close() {
	err := DB.Close()
	common.ErrFatalLog(err)
}
