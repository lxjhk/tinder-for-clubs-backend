package db

import (
	"github.com/jinzhu/gorm"
	"time"
)

// Admin Accounts
type AdminAccount struct {
	gorm.Model
	AccountID     string `gorm:"type:varchar(40);unique_index" json:"account_id"`
	AuthString string `gorm:"type:varchar(256);"            json:"auth_string"`
	ClubID     string `gorm:"type:varchar(40);"             json:"club_id"`
	// For managers of Tinder for Clubs
	IsAdmin bool `gorm:"type:tinyint(1);" json:"is_admin"`
}

func (ac *AdminAccount) Insert(txDb *gorm.DB) error {
	return txDb.Create(ac).Error
}

func (ac *AdminAccount) Update() error {
	return DB.Model(&AdminAccount{}).Where("account_id = ?", ac.AccountID).Updates(*ac).Error
}

func GetAccountById(id int64) (*AdminAccount, error) {
	var account AdminAccount
	return &account, DB.Where("id = ?", id).First(&account).Error
}

func GetAllAccounts() ([]AdminAccount, error) {
	accounts := make([]AdminAccount, 0)
	return accounts, DB.Find(&accounts).Error
}

func GetAccountByUserId(userId string) (*AdminAccount, error) {
	var account AdminAccount
	return &account, DB.Where("user_id = ?", userId).First(&account).Error
}

// Admin Account Login History
type LoginHistory struct {
	gorm.Model
	Username string `gorm:"not null;index:username"`
	IP       string
	// Whether the login attempt is successful
	AttemptResult string `gorm:"not null;"`
	// Dump the header info associated with this session
	HeaderDump string `gorm:"type:varchar(1000);"`
}

// Club Information
type ClubInfo struct {
	gorm.Model
	ClubID    string `gorm:"type:varchar(40);unique_index"  json:"club_id" binding:"required"`
	Name      string `gorm:"not null;type:varchar(1000);"   json:"name"`
	Website   string `gorm:"type:varchar(500);"             json:"website"`
	Email     string `gorm:"type:varchar(500);"             json:"email"`
	GroupLink string `gorm:"type:varchar(500);"             json:"group_link"`
	VideoLink string `gorm:"type:varchar(500);"             json:"video_link"`
	// Whether the club is viewable
	Published   bool   `gorm:"type:tinyint(1);" json:"published"`
	Description string `gorm:"type:varchar(4000);" json:"description"`

	PictureIds  string `gorm:"type:varchar(500)"`

	// Stores the ID of the pictures. The first picture will also be the cover photo
	Pic1ID string `gorm:"type:varchar(500);" json:"pic1_id"`
	Pic2ID string `gorm:"type:varchar(500);" json:"pic2_id"`
	Pic3ID string `gorm:"type:varchar(500);" json:"pic3_id"`
	Pic4ID string `gorm:"type:varchar(500);" json:"pic4_id"`
	Pic5ID string `gorm:"type:varchar(500);" json:"pic5_id"`
	Pic6ID string `gorm:"type:varchar(500);" json:"pic6_id"`
}

// Account pictures uploaded
type AccountPicture struct {
	gorm.Model
	AccountID string `gorm:"type:varchar(40);unique_index"  json:"account_id"`
	PictureID string `gorm:"type:varchar(40);unique_index"  json:"picture_id"`
	PictureName string `gorm:"type:varchar(60)"  json:"picture_name"`
}

func (ap *AccountPicture) Insert(txDb *gorm.DB) error {
	return txDb.Create(ap).Error
}

func getPictureNameById(pictureId string) (string, error) {
	var picture AccountPicture
	err := DB.Where("picture_id = ?", pictureId).First(&picture).Error
	if err != nil {
		return "", err
	}
	return picture.PictureName, nil
}

func GetAccountPicturesByPictureIds(accountId string, ids []string) ([]AccountPicture, error) {
	pictures := make([]AccountPicture, 0)
	return pictures, DB.Where("account_id = ?", accountId).Find(&pictures).Error
}

func (ci *ClubInfo) Insert(txDb *gorm.DB) error {
	return txDb.Create(ci).Error
}

func GetClubInfoByClubId(id string) (*ClubInfo, error) {
	clubInfo := &ClubInfo{}
	return clubInfo, DB.Where("club_id = ?", id).Find(clubInfo).Error
}

func (ci *ClubInfo) Update() error {
	return DB.Model(&ClubInfo{}).Where("club_id = ?", ci.ClubID).UpdateColumns(*ci).Error
}

func (ci *ClubInfo) UpdateAllPicIds() error {
	return DB.Model(&ClubInfo{}).Where("club_id = ?", ci.ClubID).Updates(ClubInfo{
		Pic1ID: ci.Pic1ID,
		Pic2ID: ci.Pic2ID,
		Pic3ID: ci.Pic3ID,
		Pic4ID: ci.Pic4ID,
		Pic5ID: ci.Pic5ID,
		Pic6ID: ci.Pic6ID,
	}).Error
}

type UserList struct {
	gorm.Model
	LoopUID      string    `gorm:"not null;"`
	LoopUserName string    `gorm:"not null;"`
	JoinTime     time.Time `gorm:"not null;"`
}

type ClubTags struct {
	gorm.Model
	TagID string `gorm:"type:varchar(40);unique_index:uni_tag"`
	Tag   string `gorm:"type:varchar(40);unique_index:uni_tag"`
}

func (ct *ClubTags) Insert() error {
	return DB.Create(&ct).Error
}

func SelectClubTags() ([]ClubTags, error) {
	tags := make([]ClubTags, 0)
	return tags, DB.Find(&tags).Error
}

type ClubTagRelationship struct {
	gorm.Model
	ClubID string `gorm:"type:varchar(40);unique_index:uni_tag"`
	TagID    string `gorm:"type:varchar(40);unique_index:uni_tag"`
}

func (cr *ClubTagRelationship) Insert() error {
	return DB.Create(&cr).Error
}

func (cr *ClubTagRelationship) Update() error {
	return DB.Model(&ClubTagRelationship{}).Where("club_id = ? and tag_id = ?", cr.ClubID, cr.TagID).UpdateColumns(*cr).Error
}
