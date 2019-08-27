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
	Email      string `gorm:"type:varchar(100);"            json:"email"`
	PhoneNum   string `gorm:"type:varchar(20);"             json:"phone_num"`
	Note       string `gorm:"type:varchar(200);"            json:"note"`
	// For managers of Tinder for Clubs
	IsAdmin bool `gorm:"type:tinyint(1);" json:"is_admin"`
}

func (ac *AdminAccount) Insert(txDb *gorm.DB) error {
	err := txDb.Create(ac).Error
	return err
}

func (ac *AdminAccount) Update() error {
	err := DB.Model(&AdminAccount{}).Where("account_id = ?", ac.AccountID).
		Updates(map[string]interface{}{"email":ac.Email,"phone_num":ac.PhoneNum,"note":ac.Note}).
		Error
	return err
}

func GetAccountByUserId(userId string) (*AdminAccount, error) {
	var account AdminAccount
	err := DB.Where("account_id = ?", userId).First(&account).Error
	return &account, err
}

func GetTotalAccountNum() (int64, error) {
	var num int64
	err := DB.Table("admin_account").Count(&num).Error
	return num, err
}

type AccountInfo struct {
	AdminAccount
	ClubName string `json:"club_name"`
}

type AccountInfoCondition struct {
	PageRequest
	SortBy   string
	SortOrder string
}

func GetAllAccountInfoByCondition(condition *AccountInfoCondition) ([]AccountInfo, error) {
	var accounts []AccountInfo

	baseQuery := DB.Select("a.*, c.name club_name").Table("admin_account a").
		Joins("LEFT JOIN club_info c ON c.club_id = a.club_id")

	if condition != nil {
		if condition.SortBy != "" {
			baseQuery = baseQuery.Order("c.created_at " + condition.SortOrder)
		}
		//pagination
		if condition.Offset != 0 {
			baseQuery = baseQuery.Offset(condition.Offset)
		}
		if condition.Limit != 0 {
			baseQuery = baseQuery.Limit(condition.Limit)
		}
	}

	err := baseQuery.Scan(&accounts).Error
	return accounts, err
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

	LogoID string `gorm:"type:varchar(40)" json:"logo_id"`
	// Stores the ID of the pictures. The first picture will also be the cover photo
	Pic1ID string `gorm:"type:varchar(500);" json:"pic1_id"`
	Pic2ID string `gorm:"type:varchar(500);" json:"pic2_id"`
	Pic3ID string `gorm:"type:varchar(500);" json:"pic3_id"`
	Pic4ID string `gorm:"type:varchar(500);" json:"pic4_id"`
	Pic5ID string `gorm:"type:varchar(500);" json:"pic5_id"`
	Pic6ID string `gorm:"type:varchar(500);" json:"pic6_id"`
}

type ClubInfoCount struct {
	ClubInfo
	FavouriteNum int64 `json:"favourite_num"`
	ViewNum int64 `json:"view_num"`
}

type PageRequest struct {
	CurrPage int64
	PageSize int64
	Offset int64
	Limit int64
}

type ClubInfoCondition struct {
	PageRequest
	SortBy   string
	SortOrder string
	Published string
}

//Returns given condition club info and their count of favourite num and view num.
func GetClubInfoCountsByCondition(condition *ClubInfoCondition) ([]ClubInfoCount, error) {
	var clubInfos []ClubInfoCount

	favouriteNumQuery := DB.Select("club_id, count(*) favourite_num").Table("user_favourite").Group("club_id").SubQuery()
	viewNumQuery := DB.Select("club_id, count(*) view_num").Table("view_list_log").Group("club_id").SubQuery()
	baseQuery := DB.Table("club_info c").Select("c.*, f.favourite_num, v.view_num").
		Joins("LEFT JOIN ? f ON c.club_id = f.club_id", favouriteNumQuery).
		Joins("LEFT JOIN ? v ON c.club_id = v.club_id", viewNumQuery)

	//search by conditions
	if condition != nil {
		if condition.Published == "true" {
			baseQuery = baseQuery.Where("c.published = 1")
		}
		if condition.Published == "false" {
			baseQuery = baseQuery.Where("c.published = 0")
		}
		if condition.SortBy != "" {
			baseQuery = baseQuery.Order("c.created_at " + condition.SortOrder)
		}
		//pagination
		if condition.Offset != 0 {
			baseQuery = baseQuery.Offset(condition.Offset)
		}
		if condition.Limit != 0 {
			baseQuery = baseQuery.Limit(condition.Limit)
		}
	}

	err := baseQuery.Scan(&clubInfos).Error
	return clubInfos, err
}

func GetClubInfoNumByCondition(condition *ClubInfoCondition) (int64, error) {
	var num int64

	favouriteNumQuery := DB.Select("club_id, count(*) favourite_num").Table("user_favourite").Group("club_id").SubQuery()
	viewNumQuery := DB.Select("club_id, count(*) view_num").Table("view_list_log").Group("club_id").SubQuery()
	baseQuery := DB.Table("club_info c").
		Joins("LEFT JOIN ? f ON c.club_id = f.club_id", favouriteNumQuery).
		Joins("LEFT JOIN ? v ON c.club_id = v.club_id", viewNumQuery)

	if condition != nil {
		if condition.Published == "true" {
			baseQuery = baseQuery.Where("c.published = 1")
		}
		if condition.Published == "false" {
			baseQuery = baseQuery.Where("c.published = 0")
		}
	}

	err := baseQuery.Count(&num).Error
	return num, err
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

func GetPublishedClubInfosByClubIds(ids []string) ([]ClubInfo, error) {
	clubInfos := make([]ClubInfo, 0)
	err := DB.Where("club_id in (?) AND published = 1", ids).Find(&clubInfos).Error
	return clubInfos, err
}

//Club manager may add or remove info in club, so we must update all columns, even the columns have default value.
func (ci *ClubInfo) Update(txDb *gorm.DB) error {
	err := txDb.Model(&ClubInfo{}).Where("club_id = ?", ci.ClubID).
		Updates(map[string]interface{}{"name": ci.Name, "website": ci.Website, "email": ci.Email, "group_link": ci.GroupLink, "video_link": ci.VideoLink, "published": ci.Published, "description": ci.Description,
			"logo_id": ci.LogoID, "pic1_id": ci.Pic1ID, "pic2_id": ci.Pic2ID, "pic3_id": ci.Pic3ID, "pic4_id": ci.Pic4ID, "pic5_id": ci.Pic5ID, "pic6_id": ci.Pic6ID}).Error
	return err
}

//FavouriteClubInfo is a assist struct to query club info to app user.
type FavouriteClubInfo struct {
	ClubInfo
	Favourite bool
}

//Get all club infos attached with current user favourite or not
func GetAllPublishedFavouriteClubInfo(uid string) ([]FavouriteClubInfo, error) {
	favouriteClubInfos := make([]FavouriteClubInfo, 0)
	err := DB.Table("club_info c").Select("c.*, f.favourite").
		Joins("LEFT JOIN (SELECT * FROM user_favourite WHERE loop_uid = ?) f ON c.club_id = f.club_id WHERE c.published = 1", uid).
		Scan(&favouriteClubInfos).
		Error
	return favouriteClubInfos, err
}

func GetAllPublishedFavouriteClubInfoByClubIDs(uid string, ids []string) ([]FavouriteClubInfo, error) {
	favouriteClubInfos := make([]FavouriteClubInfo, 0)
	err := DB.Table("club_info c").Select("c.*, f.favourite").
		Joins("LEFT JOIN (SELECT * FROM user_favourite WHERE loop_uid = ?) f ON c.club_id = f.club_id WHERE c.club_id in (?) and c.published = 1", uid, ids).
		Scan(&favouriteClubInfos).
		Error
	return favouriteClubInfos, err
}

// Account pictures uploaded
type AccountPicture struct {
	gorm.Model
	AccountID   string `gorm:"type:varchar(40);index"  json:"account_id"`
	PictureID   string `gorm:"type:varchar(40);unique_index"  json:"picture_id"`
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
	ClubID string `gorm:"type:varchar(40);index"`
	TagID  string `gorm:"type:varchar(40);index"`
}

func GetTagRelationshipsByTagIDs(tagIDs []string) ([]ClubTagRelationship, error) {
	relations := make([]ClubTagRelationship, 0)
	err := DB.Select("DISTINCT(club_id),tag_id").Where("tag_id in (?)", tagIDs).Find(&relations).Error
	return relations, err
}

func GetTagRelationshipsByClubID(clubID string) ([]ClubTagRelationship, error) {
	relations := make([]ClubTagRelationship, 0)
	err := DB.Where("club_id = ?", clubID).Find(&relations).Error
	return relations, err
}

func (cr *ClubTagRelationship) Insert(txDb *gorm.DB) error {
	err := txDb.Create(&cr).Error
	return err
}

func CleanAllTags(txDb *gorm.DB, clubId string) error {
	err := txDb.Where("club_id = ?", clubId).Delete(&ClubTagRelationship{}).Error
	return err
}

type UserList struct {
	gorm.Model
	//consider user may cancel authorization, use index rather than unique index here.
	LoopUID      string `gorm:"type:varchar(70);index"`
	LoopUserName string `gorm:"type:varchar(50)"`
	JoinTime     time.Time
}

func (ul *UserList) Insert() error {
	err := DB.Create(ul).Error
	return err
}

func GetAppUserByUid(uid string) (*UserList, error) {
	var user UserList
	err := DB.Where("loop_uid = ?", uid).First(&user).Error
	return &user, err
}

type ViewList struct {
	//append only
	gorm.Model
	LoopUID    string `gorm:"type:varchar(70);unique_index:uni_view"`
	ViewListID string `gorm:"type:varchar(40);unique_index:uni_view"`
}

func GetLatestViewListByUID(uid string) (*ViewList, error) {
	var viewList ViewList
	err := DB.Where("loop_uid = ?", uid).Last(&viewList).Error
	return &viewList, err
}

func (vl *ViewList) Insert() error {
	err := DB.Create(vl).Error
	return err
}

type ViewListLog struct {
	//append only
	gorm.Model
	ViewListID string `gorm:"type:varchar(40);index"`
	LoopUID    string `gorm:"type:varchar(70);index"`
	ClubID     string `gorm:"type:varchar(40);index"`
}

func (l *ViewListLog) Insert() error {
	err := DB.Create(l).Error
	return err
}

func GetViewedListByID(uid, viewId string) ([]ViewListLog, error) {
	logs := make([]ViewListLog, 0)
	err := DB.Where("loop_uid = ? and view_list_id = ?", uid, viewId).Find(&logs).Error
	return logs, err
}

type UserFavourite struct {
	//append and delete
	gorm.Model
	LoopUID   string `gorm:"type:varchar(70);index"`
	ClubID    string `gorm:"type:varchar(40);index"`
	Favourite bool   `gorm:"type:tinyint(1)"`
}

func (f *UserFavourite) InsertOrUpdate(txDb *gorm.DB) error {
	err := txDb.Where("Loop_uid = ? and club_id = ?", f.LoopUID, f.ClubID).Assign(map[string]interface{}{"favourite": f.Favourite}).FirstOrCreate(f).Error
	return err
}

func GetUserFavouritesByUID(uid string) ([]UserFavourite, error) {
	favourites := make([]UserFavourite, 0)
	err := DB.Where("loop_uid = ? and favourite = 1", uid).Find(&favourites).Error
	return favourites, err
}

const (
	FAVORITE_ACTION   = "FAVORITE"
	UNFAVORITE_ACTION = "UNFAVORITE"
)

type UserFavouriteLog struct {
	// append only log
	gorm.Model
	LoopUID string `gorm:"type:varchar(70);index"`
	ClubID  string `gorm:"type:varchar(40);index"`
	Action  string `gorm:"type:varchar(20)"` // "favourite" "unfavourite"
}

func (l *UserFavouriteLog) Insert(txDb *gorm.DB) error {
	err := txDb.Create(l).Error
	return err
}
