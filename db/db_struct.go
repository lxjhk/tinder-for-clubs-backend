package db

import (
	"github.com/jinzhu/gorm"
	"time"
)

// Admin Accounts
type AdminAccount struct {
	gorm.Model
	UserID     int    `gorm:"not null;"`
	AuthString string `gorm:"type:varchar(1000);"`
	ClubID     string `gorm:"type:varchar(40);"`
	// For managers of Tinder for Clubs
	IsAdmin string
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
	Timestamp  time.Time
}

// Club Information
type ClubInfo struct {
	gorm.Model
	ClubID    string `gorm:"type:varchar(100);"`
	Name      string `gorm:"not null;type:varchar(1000);"`
	Website   string `gorm:"type:varchar(500);"`
	Email     string `gorm:"type:varchar(500);"`
	GroupLink string `gorm:"type:varchar(500);"`
	VideoLink string `gorm:"type:varchar(500);"`
	// Whether the club is viewable
	Published   bool   `gorm:"not null;"`
	Description string `gorm:"type:varchar(5000);"`

	// Stores the ID of the pictures. The first picture will also be the cover photo
	Pic1ID string `gorm:"type:varchar(100);"`
	Pic2ID string `gorm:"type:varchar(100);"`
	Pic3ID string `gorm:"type:varchar(100);"`
	Pic4ID string `gorm:"type:varchar(100);"`
	Pic5ID string `gorm:"type:varchar(100);"`
	Pic6ID string `gorm:"type:varchar(100);"`

	// Last update time of this entry
	LastUpdateTime int64
}

func GetClubInfoById(id string) (*ClubInfo, error) {
	clubInfo := &ClubInfo{}
	return clubInfo, DB.Where("id = ?", id).Find(clubInfo).Error
}

func (ci ClubInfo) UpdateAllPicIds() error {
	return DB.Model(&ClubInfo{}).Where("club_uuid = ?", ci.ID).UpdateColumns(ClubInfo{
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
	ID  string `gorm:"AUTO_INCREMENT"`
	Tag string `gorm:"not null;"`
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
	ClubUUID string `gorm:"not null;"`
	TagID    string `gorm:"not null;"`
}

func (cr *ClubTagRelationship) Insert() error {
	return DB.Create(&cr).Error
}

func (cr *ClubTagRelationship) Update() error {
	return DB.Model(&ClubTagRelationship{}).Where("club_uuid = ? and tag_id = ?", cr.ClubUUID, cr.TagID).UpdateColumns(*cr).Error
}
