package db

import (
	"github.com/jinzhu/gorm"
	"time"
)

// Admin Accounts
type AdminAccount struct {
	gorm.Model
	AccountID  string `gorm:"type:varchar(40);unique_index" json:"account_id"`
	AuthString string `gorm:"type:varchar(256);"            json:"auth_string"`
	ClubID     string `gorm:"type:varchar(40);"             json:"club_id"`
	// For managers of Tinder for Clubs
	IsAdmin bool `gorm:"type:tinyint(1);" json:"is_admin"`
}

func (ac *AdminAccount) Insert(txDb *gorm.DB) error {
	err := txDb.Create(ac).Error
	return err
}

func (ac *AdminAccount) Update() error {
	err := DB.Model(&AdminAccount{}).Where("account_id = ?", ac.AccountID).Updates(*ac).Error
	return err
}

func GetAccountById(id int64) (*AdminAccount, error) {
	var account AdminAccount
	err := DB.Where("id = ?", id).First(&account).Error
	return &account, err
}

func GetAllAccounts() ([]AdminAccount, error) {
	accounts := make([]AdminAccount, 0)
	err := DB.Find(&accounts).Error
	return accounts, err
}

func GetAccountByUserId(userId string) (*AdminAccount, error) {
	var account AdminAccount
	err := DB.Where("account_id = ?", userId).First(&account).Error
	return &account, err
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
	AccountID   string `gorm:"type:varchar(40);"  json:"account_id"`
	PictureID   string `gorm:"type:varchar(40);"  json:"picture_id"`
	PictureName string `gorm:"type:varchar(60)"   json:"picture_name"`
}

func (ap *AccountPicture) Insert(txDb *gorm.DB) error {
	err := txDb.Create(ap).Error
	return err
}

func GetPictureNameById(pictureId string) (string, error) {
	var picture AccountPicture
	err := DB.Where("picture_id = ?", pictureId).First(&picture).Error
	if err != nil {
		return "", err
	}
	return picture.PictureName, nil
}

func GetAccPictureIDS(accountId string) ([]AccountPicture, error) {
	pictures := make([]AccountPicture, 0)
	err := DB.Select("picture_id").Where("account_id = ?", accountId).Find(&pictures).Error
	return pictures, err
}

func (ci *ClubInfo) Insert(txDb *gorm.DB) error {
	err := txDb.Create(ci).Error
	return err
}

func GetClubInfoByClubId(id string) (*ClubInfo, error) {
	clubInfo := &ClubInfo{}
	err := DB.Where("club_id = ?", id).Find(clubInfo).Error
	return clubInfo, err
}

//Club manager may add or remove info in club, so we must update all columns, even the columns have default value.
func (ci *ClubInfo) Update(txDb *gorm.DB) error {
	err := txDb.Model(&ClubInfo{}).Where("club_id = ?", ci.ClubID).
		Updates(map[string]interface{}{"name":ci.Name, "website":ci.Website, "email":ci.Email, "group_link":ci.GroupLink, "video_link":ci.VideoLink, "published":ci.Published, "description":ci.Description,
			"pic1_id":ci.Pic1ID, "pic2_id":ci.Pic2ID, "pic3_id":ci.Pic3ID, "pic4_id":ci.Pic4ID, "pic5_id":ci.Pic5ID, "pic6_id":ci.Pic6ID}).Error
	return err
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

func GetClubTagsByTagIds(ids []string) ([]ClubTags, error) {
	tags := make([]ClubTags, 0)
	err := DB.Where("tag_id in (?)", ids).Find(&tags).Error
	return tags, err
}

func (ct *ClubTags) Insert() error {
	err := DB.Create(&ct).Error
	return err
}

func GetAllClubTags() ([]ClubTags, error) {
	tags := make([]ClubTags, 0)
	err := DB.Find(&tags).Error
	return tags, err
}

type ClubTagRelationship struct {
	gorm.Model
	ClubID string `gorm:"type:varchar(40);unique_index:uni_tag"`
	TagID  string `gorm:"type:varchar(40);unique_index:uni_tag"`
}

func (cr *ClubTagRelationship) Insert(txDb *gorm.DB) error {
	err := txDb.Create(&cr).Error
	return err
}

func CleanAllTags(txDb *gorm.DB, clubId string) error {
	err := txDb.Where("clubId = ?", clubId).Delete(&ClubTagRelationship{}).Error
	return err
}
