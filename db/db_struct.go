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
	ClubID     string
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

type ClubTagRelationship struct {
	gorm.Model
	ClubUUID string `gorm:"not null;"`
	TagID    string `gorm:"not null;"`
}
