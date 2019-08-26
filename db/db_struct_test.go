package db

import (
	"log"
	"testing"
	"tinder-for-clubs-backend/config"
)

func initTestConfiguration() *config.GlobalConfiguration {
	configuration := config.GlobalConfiguration{
		DBCredential: config.DBCredential{
			DBAddress: "127.0.0.1",
			DBName: "tinder-for-clubs",
			DBPort: "3306",
			DBUser: "root",
			DBPass: "root",
		},
	}
	return &configuration
}

func TestGetAllAccountInfoByCondition(t *testing.T)  {
	configuration := initTestConfiguration()
	Init(configuration.DBCredential)

	accounts, err := GetAllAccountInfoByCondition(nil)

	if err != nil {
		t.Fatal(err)
	}

	_ = accounts
}

func TestGetClubInfoCountsByCondition(t *testing.T) {
	configuration := initTestConfiguration()
	Init(configuration.DBCredential)

	condition := &ClubInfoCondition{
		Published:"true",
		SortBy: "created_at",
		SortOrder: "DESC",
	}

	counts, err := GetClubInfoCountsByCondition(condition)
	if err != nil {
		t.Fatal(err)
	}
	_ = counts

	totalSize, err := GetClubInfoNumByCondition(condition)
	if err != nil {
		t.Fatal(err)
	}
	log.Println(totalSize)
}

func TestClubInfo_Update(t *testing.T) {
	configuration := initTestConfiguration()
	Init(configuration.DBCredential)

	clubInfo := ClubInfo{
		ClubID: "7661a656-5ba6-4fc4-b43c-cd8b6cc09a6e",
	}

	//insert new club info
	err := DB.Create(&clubInfo).Error
	if err != nil {
		t.Fatal(err)
	}

	//update club info
	clubInfo.Pic1ID = "75cd39d5-baa0-40cd-8314-833d784cfc2d.png"
	err = DB.Model(&clubInfo).Where("club_id = ?", clubInfo.ClubID).Updates(map[string]interface{}{"pic1_id":clubInfo.Pic1ID}).Error
	if err != nil {
		t.Fatal(err)
	}

	//cover club info
	clubInfo.Pic1ID = ""
	err = DB.Model(&clubInfo).Where("club_id = ?", clubInfo.ClubID).Updates(map[string]interface{}{"pic1_id":clubInfo.Pic1ID}).Error
	if err != nil {
		t.Fatal(err)
	}
}
