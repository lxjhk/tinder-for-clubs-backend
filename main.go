package main

import (
	"crypto/sha512"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	set "github.com/deckarep/golang-set"
	"github.com/gin-contrib/secure"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"net/http"
	"path"
	"reflect"
	"strings"
	"time"
	"tinder-for-clubs-backend/common"
	"tinder-for-clubs-backend/config"
	"tinder-for-clubs-backend/db"
	"tinder-for-clubs-backend/httpserver"
)

var router *gin.Engine
var globalConfig config.GlobalConfiguration

const (
	USER = "USER"
)

func main() {
	// Reading configuration file
	globalConfig = config.ReadConfig()

	// Setting up database connection
	db.Init(globalConfig.DBCredential)

	//Deferred Closed
	defer db.Close()

	// Initialise HTTP framework and Session Store
	router = gin.Default()

	// Initialise mem session storage
	store := memstore.NewStore([]byte(uuid.New().String()))
	router.Use(sessions.Sessions("AdminSession", store))

	gob.Register(db.AdminAccount{})

	initRouter(router)
	err := router.Run() // listen and serve on 0.0.0.0:8080
	common.ErrFatalLog(err)
}

func initRouter(engine *gin.Engine) {
	// Disable inline scripts
	router.Use(secure.New(secure.Config{
		ContentSecurityPolicy: "default-src 'self'",
	}))

	router.Use(CORSMiddleware())

	// Register Handler
	initializeRoutes()
}

// CORS allow domains so that we can serve the website from different subdomains.
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, user-id")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func initializeRoutes() {
	//Ping endpoint for testing
	router.GET("/ping", Pong)

	//For admin and club managers to login
	router.POST("/login", Login)
	router.DELETE("/logout", logout)
	router.GET("/authorized",ifAuthorized)

	// Admin only endpoints
	router.POST("/admin/account/create", createNewClubAccount)
	router.GET("/admin/account/all", listAccounts)
	router.GET("/admin/account/user/:userId", getAccountByUserId)

	// Club manager endpoints
	router.GET("/account", getCurrUser)
	router.POST("/club/uploadpicture", uploadSinglePicture)
	router.POST("/club/info", updateClubInfo)
	router.GET("/club/info", getClubInfo)
	router.GET("/club/tags", adminGetAllTags)

	// MiniApp endpoints
	router.GET("/static/clubphoto/:pictureID", serveStaticPicture)
	router.Static("/clubphoto",globalConfig.General.PictureStoragePath)

	router.POST("/app/register", registerAppUser)
	router.GET("/app/userinfo", getAppUserInfo)
	router.GET("/app/favourite", getFavouriteClubList)
	router.PUT("/app/favourite/:clubID", setFavouriteClub)
	router.PUT("/app/unfavourite/:clubID", setUnfavouriteClub)
	router.GET("/app/clubs/all", GetAllClubs)
	router.GET("/app/tagfilter", getClubInfoOfGivenTags)
	router.GET("/app/tages", appGetAllTags)
	router.GET("/app/viewlist/unreadlist/:viewListID", getUnreadViewList)
	router.GET("/app/viewlist/new", getNewViewList)
	router.PUT("/app/viewlist/markread", markClubReadInViewList)
}

func logout(ctx *gin.Context) {
	// if user exist in the session
	session := sessions.Default(ctx)
	result := session.Get(USER)
	if result == nil {
		ctx.JSON(http.StatusUnauthorized, httpserver.ConstructResponse(httpserver.NOT_AUTHORIZED, false))
		return
	}

	// delete the user in the session
	session.Delete(USER)
	if err := session.Save(); err != nil {
		ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
		return
	}

	ctx.JSON(http.StatusOK, httpserver.SuccessResponse(nil))
}

func getUnreadViewList(ctx *gin.Context) {
	//check user
	user, err := getAppUser(ctx)
	if err != nil {
		log.Error(err)
		return
	}

	viewListID := ctx.Param("viewListID")
	if viewListID == "" {
		ctx.JSON(http.StatusBadRequest, httpserver.ConstructResponse(httpserver.INVALID_PARAMS, nil))
		return
	}

	//Get all club infos attached with current user favourite or not
	favouriteClubInfos, err := db.GetAllPublishedFavouriteClubInfo(user.LoopUID)
	if err != nil {
		log.Error(err)
		ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
		return
	}

	//construct response club info from DB query result
	responseClubs, err := getResponseFromFavouriteClubInfos(favouriteClubInfos)
	if err != nil {
		log.Error(err)
		ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
		return
	}

	//shuffle response club infos
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(responseClubs), func(i, j int) { responseClubs[i], responseClubs[j] = responseClubs[j], responseClubs[i] })

	//get read club ids
	logs, err := db.GetViewedListByID(user.LoopUID, viewListID)
	if err != nil {
		log.Error(err)
		ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
		return
	}

	//remove clubs already read in response clubs
	responseUnreadClubs := make([]FavouriteClubInfo, 0)
	for _, club :=range responseClubs {
		var read bool
		for _, l :=range logs {
			if club.ClubID == l.ClubID {
				read = true
				break
			}
		}
		if read {
			continue
		}
		responseUnreadClubs = append(responseUnreadClubs, club)
	}

	//construct response unread view list
	viewListPost := ViewListPost{
		ViewListId: viewListID,
		ViewList: responseUnreadClubs,
	}
	ctx.JSON(http.StatusOK, httpserver.SuccessResponse(viewListPost))
}

type MarkClubReadInViewListRequest struct {
	ViewListId  string `json:"view_list_id"`
	ClubId      string `json:"club_id"`
}

//Marks club already read by user in current view list.
func markClubReadInViewList(ctx *gin.Context) {
	//get request user id
	userId := ctx.GetHeader("user-id")
	if userId == "" {
		ctx.JSON(http.StatusUnauthorized, httpserver.ConstructResponse(httpserver.NOT_AUTHORIZED, nil))
		return
	}

	//get request params
	markReq := new(MarkClubReadInViewListRequest)
	if err := ctx.ShouldBindJSON(markReq); err != nil {
		ctx.JSON(http.StatusBadRequest, httpserver.ConstructResponse(httpserver.INVALID_PARAMS, nil))
		log.Error(err)
		return
	}
	//check user and view list
	_, err := db.GetViewList(userId, markReq.ViewListId)
	if gorm.IsRecordNotFoundError(err) {
		ctx.JSON(http.StatusBadRequest, httpserver.ConstructResponse(httpserver.INVALID_PARAMS, nil))
		return
	}
	if err != nil {
		log.Error(err)
		ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
		return
	}
	//check club
	_, err = db.GetClubInfoByClubId(markReq.ClubId)
	if gorm.IsRecordNotFoundError(err) {
		ctx.JSON(http.StatusBadRequest, httpserver.ConstructResponse(httpserver.INVALID_PARAMS, nil))
		return
	}
	if err != nil {
		log.Error(err)
		ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
		return
	}

	//save read info into DB
	viewLog := db.ViewListLog{
		ViewListID: markReq.ViewListId,
		LoopUID: userId,
		ClubID: markReq.ClubId,
	}
	err = viewLog.Insert()
	if err != nil {
		log.Error(err)
		ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
		return
	}

	ctx.JSON(http.StatusOK, httpserver.SuccessResponse(nil))
}

//Returns a new club view list and corresponding id.
// Club view list is current all published clubs that sequence shuffled.
func getNewViewList(ctx *gin.Context) {
	//check user
	user, err := getAppUser(ctx)
	if err != nil {
		log.Error(err)
		return
	}

	//create new view list info into
	viewList := db.ViewList{
		LoopUID: user.LoopUID,
		ViewListID: uuid.New().String(),
	}
	err = viewList.Insert()
	if err != nil {
		log.Error(err)
		ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
		return
	}

	//Get all club infos attached with current user favourite or not
	favouriteClubInfos, err := db.GetAllPublishedFavouriteClubInfo(user.LoopUID)
	if err != nil {
		log.Error(err)
		ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
		return
	}

	//construct response club info from DB query result
	responseClubs, err := getResponseFromFavouriteClubInfos(favouriteClubInfos)
	if err != nil {
		log.Error(err)
		ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
		return
	}

	//shuffle response club infos
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(responseClubs), func(i, j int) { responseClubs[i], responseClubs[j] = responseClubs[j], responseClubs[i] })

	//construct response view list
	viewListPost := ViewListPost{
		ViewListId: viewList.ViewListID,
		ViewList: responseClubs,
	}

	ctx.JSON(http.StatusOK, httpserver.SuccessResponse(viewListPost))
}

type ViewListPost struct {
	ViewListId  string `json:"view_list_id"`
	ViewList []FavouriteClubInfo `json:"view_list"`
}

func getClubInfoOfGivenTags(ctx *gin.Context) {
	//check user
	user, err := getAppUser(ctx)
	if err != nil {
		log.Error(err)
		return
	}

	//get request tags
	tagStr := ctx.Query("tag_id")
	if tagStr == "" {
		ctx.JSON(http.StatusBadRequest, httpserver.ConstructResponse(httpserver.INVALID_PARAMS, "tags not specified"))
		return
	}
	tagIDs := strings.Split(tagStr, ";")

	//validate tags
	tags, err := db.GetClubTagsByTagIds(tagIDs)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
		return
	}
	if len(tags) != len(tagIDs) {
		ctx.JSON(http.StatusBadRequest, httpserver.ConstructResponse(httpserver.INVALID_PARAMS, "invalid tag id"))
		return
	}

	//get clubs having given tag id.
	relationships, err := db.GetTagRelationshipsByTagIDs(tagIDs)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
		return
	}
	//unite club ids from clubs
	clubIDs := make([]string, 0)
	for _, relation :=range relationships {
		clubIDs = append(clubIDs, relation.ClubID)
	}

	favouriteClubInfos, err := db.GetAllPublishedFavouriteClubInfoByClubIDs(user.LoopUID, clubIDs)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
		return
	}

	responseClubs, err := getResponseFromFavouriteClubInfos(favouriteClubInfos)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
	}

	ctx.JSON(http.StatusOK, httpserver.SuccessResponse(responseClubs))
}

//FavouriteClubInfo is a assist struct to response club info to app user.
type FavouriteClubInfo struct {
	ClubInfoPost
	//is user favourite to this club
	Favourite bool `json:"favourite"`
}

func GetAllClubs(ctx *gin.Context) {
	//check user
	user, err := getAppUser(ctx)
	if err != nil {
		log.Error(err)
		return
	}

	favouriteClubInfos, err := db.GetAllPublishedFavouriteClubInfo(user.LoopUID)
	if err != nil {
		log.Error(err)
		ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
		return
	}

	responseInfo, err := getResponseFromFavouriteClubInfos(favouriteClubInfos)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
	}

	ctx.JSON(http.StatusOK, httpserver.SuccessResponse(responseInfo))
}

//Constructs club response info from DB query result
func getResponseFromFavouriteClubInfos(favouriteClubInfos []db.FavouriteClubInfo) ([]FavouriteClubInfo, error) {
	clubInfos := make([]FavouriteClubInfo, 0)

	for _, clubInfo :=range favouriteClubInfos {
		tagIDs, pictureIDs, err := getClubTagIdsAndPictureIds(clubInfo.ClubID,
			clubInfo.Pic1ID, clubInfo.Pic2ID, clubInfo.Pic3ID, clubInfo.Pic4ID, clubInfo.Pic5ID, clubInfo.Pic6ID)
		if err != nil {
			return clubInfos, err
		}

		infoPost := ClubInfoPost{
			ClubID:      clubInfo.ClubID,
			Name:        clubInfo.Name,
			Website:     clubInfo.Website,
			Email:       clubInfo.Email,
			GroupLink:   clubInfo.GroupLink,
			VideoLink:   clubInfo.VideoLink,
			Published:   clubInfo.Published,
			Description: clubInfo.Description,
			LogoId:      clubInfo.LogoID,
			TagIds:      tagIDs,
			PictureIds:  pictureIDs,
		}
		responseInfo := FavouriteClubInfo{
			ClubInfoPost: infoPost,
			Favourite: clubInfo.Favourite,
		}
		clubInfos = append(clubInfos, responseInfo)
	}

	return clubInfos, nil
}

//Sets club is unfavourite to user
func setUnfavouriteClub(ctx *gin.Context) {
	doFavouriteClub(ctx, false)
}

//Sets club is favourite to user
func setFavouriteClub(ctx *gin.Context) {
	doFavouriteClub(ctx, true)
}

//Sets club is whatever favourite or unfavourite to user
func doFavouriteClub(ctx *gin.Context, favourite bool) {
	//check user
	user, err := getAppUser(ctx)
	if err != nil {
		log.Error(err)
		return
	}
	//check club id
	clubID := ctx.Param("clubID")
	_, err = db.GetClubInfoByClubId(clubID)
	if gorm.IsRecordNotFoundError(err) {
		ctx.JSON(http.StatusBadRequest, httpserver.ConstructResponse(httpserver.INVALID_PARAMS, nil))
		return
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
		log.Error(err)
		return
	}

	err = setFavouriteStateAndLogIntoDB(user.LoopUID, clubID, favourite)
	if err != nil {
		log.Error(err)
		ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
		return
	}

	ctx.JSON(http.StatusOK, httpserver.SuccessResponse(nil))
}

func setFavouriteStateAndLogIntoDB(loopUID, clubID string, like bool) error {
	//insert favourite club, rolls back when error occurs
	favourite := db.UserFavourite{
		LoopUID: loopUID,
		ClubID: clubID,
		Favourite: like,
	}

	txDb := db.DB.Begin()
	err := favourite.InsertOrUpdate(txDb)
	if err != nil {
		txDb.Rollback()
		return err
	}

	//what action user is doing
	var action string
	if like {
		action = db.FAVORITE_ACTION
	}else {
		action = db.UNFAVORITE_ACTION
	}

	//insert log, rolls back when error occurs
	l := db.UserFavouriteLog{
		LoopUID: loopUID,
		ClubID: clubID,
		Action: action,
	}
	err = l.Insert(txDb)
	if err != nil {
		txDb.Rollback()
		return err
	}

	txDb.Commit()
	return nil
}

//Returns the clubs that user favourite, whether published or not.
func getFavouriteClubList(ctx *gin.Context) {
	user, err := getAppUser(ctx)
	if err != nil {
		log.Error(err)
		return
	}

	//get favorite club ids
	favourites, err := db.GetUserFavouritesByUID(user.LoopUID)
	clubIds := make([]string, 0)
	for _, favourite :=range favourites {
		clubIds = append(clubIds, favourite.ClubID)
	}

	//get club infos
	clubInfos, err := db.GetPublishedClubInfosByClubIds(clubIds)

	//construct response info
	clubInfoResponses := make([]ClubInfoPost, 0)
	for _,clubInfo :=range clubInfos {
		tagIDs, pictureIDs, err := getClubTagIdsAndPictureIds(clubInfo.ClubID,
			clubInfo.Pic1ID, clubInfo.Pic2ID, clubInfo.Pic3ID, clubInfo.Pic4ID, clubInfo.Pic5ID, clubInfo.Pic6ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
			return
		}
		clubInfoResponse := constructClubInfoPost(&clubInfo, tagIDs, pictureIDs)
		clubInfoResponses = append(clubInfoResponses, *clubInfoResponse)
	}

	ctx.JSON(http.StatusOK, httpserver.SuccessResponse(clubInfoResponses))
}

func getAppUserInfo(ctx *gin.Context) {
	user, err := getAppUser(ctx)
	if err != nil {
		log.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, httpserver.SuccessResponse(user))
}

//Get user from DB by uid set in request header.
func getAppUser(ctx *gin.Context) (*db.UserList, error) {
	userId := ctx.GetHeader("user-id")
	if userId == "" {
		ctx.JSON(http.StatusUnauthorized, httpserver.ConstructResponse(httpserver.NOT_AUTHORIZED, nil))
		return nil, errors.New("header not found")
	}
	user, err := db.GetAppUserByUid(userId)
	if gorm.IsRecordNotFoundError(err) {
		ctx.JSON(http.StatusUnauthorized, httpserver.ConstructResponse(httpserver.NOT_AUTHORIZED, nil))
		return nil, errors.New("user not found")
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
		return nil, err
	}

	return user, nil
}

type UserPost struct {
	LoopUID      string `json:"loop_uid"`
	LoopUserName string `json:"loop_user_name"`
}

//Used to register for LOOP user.
func registerAppUser(ctx *gin.Context) {
	userPost := new(UserPost)
	if err := ctx.ShouldBindJSON(userPost); err != nil {
		ctx.JSON(http.StatusBadRequest, httpserver.ConstructResponse(httpserver.INVALID_PARAMS, nil))
		log.Error(err)
		return
	}
	if len(userPost.LoopUID) != 64 || len(userPost.LoopUserName) == 0 {
		ctx.JSON(http.StatusBadRequest, httpserver.ConstructResponse(httpserver.INVALID_PARAMS, nil))
		return
	}
	if len(userPost.LoopUserName) == 0 {
		ctx.JSON(http.StatusBadRequest, httpserver.ConstructResponse(httpserver.INVALID_PARAMS, nil))
		return
	}

	foundUser, err := db.GetAppUserByUid(userPost.LoopUID)
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
		log.Print(err)
		return
	}
	if foundUser.LoopUID != "" {
		ctx.JSON(http.StatusBadRequest, httpserver.ConstructResponse(httpserver.USER_ALREADY_REGISTERED, nil))
		return
	}

	user := db.UserList{
		LoopUID: userPost.LoopUID,
		LoopUserName: userPost.LoopUserName,
		JoinTime: time.Now(),
	}
	err = user.Insert()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
		log.Print(err)
		return
	}

	ctx.JSON(http.StatusOK, httpserver.SuccessResponse(nil))
}

func ifAuthorized(ctx *gin.Context) {
	_, err := getAdminUser(ctx)
	if err != nil {
		return
	}
	ctx.JSON(http.StatusOK, httpserver.SuccessResponse(true))
}

//Returns current login user
func getCurrUser(ctx *gin.Context) {
	account, err := getAdminUser(ctx)
	if err != nil {
		return
	}

	ctx.JSON(http.StatusOK, httpserver.SuccessResponse(account))
}

func appGetAllTags(ctx *gin.Context)  {
	_, err := getAppUser(ctx)
	if err != nil {
		return
	}

	tags, err := db.GetAllClubTags()
	if err != nil {
		log.Error(err)
		ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
		return
	}

	ctx.JSON(http.StatusOK, httpserver.SuccessResponse(tags))
}

func adminGetAllTags(ctx *gin.Context)  {
	_, err := getAdminUser(ctx)
	if err != nil {
		return
	}

	tags, err := db.GetAllClubTags()
	if err != nil {
		log.Error(err)
		ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
		return
	}

	ctx.JSON(http.StatusOK, httpserver.SuccessResponse(tags))
}

func serveStaticPicture(ctx *gin.Context) {
	pictureID := ctx.Param("pictureID")
	if pictureID == "" {
		ctx.JSON(http.StatusBadRequest, httpserver.ConstructResponse(httpserver.INVALID_PICTURE_ID, nil))
		return
	}
	picName, err := db.GetPictureNameById(pictureID)
	if gorm.IsRecordNotFoundError(err) {
		ctx.JSON(http.StatusBadRequest, httpserver.ConstructResponse(httpserver.INVALID_PICTURE_ID, nil))
		return
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
		return
	}
	relativePath := "/clubphoto/" + picName
	ctx.JSON(http.StatusOK, httpserver.SuccessResponse(relativePath))
}

// Get User information from request context.
func getAdminUser(ctx *gin.Context) (*db.AdminAccount, error) {
	session := sessions.Default(ctx)
	result := session.Get(USER)
	if result == nil {
		ctx.JSON(http.StatusUnauthorized, httpserver.ConstructResponse(httpserver.NOT_AUTHORIZED, false))
		return nil, errors.New("user invalid")
	}

	user := result.(db.AdminAccount)

	return &user, nil
}

// Returns the club info of current login user.
func getClubInfo(ctx *gin.Context) {
	account, err := getAdminUser(ctx)
	if err != nil {
		return
	}

	if account.IsAdmin {
		emptyClubInfo := ClubInfoPost{
			TagIds:     make([]string, 0),
			PictureIds: make([]string, 0),
		}
		ctx.JSON(http.StatusOK, httpserver.SuccessResponse(emptyClubInfo))
		return
	}

	clubInfo, err := db.GetClubInfoByClubId(account.ClubID)
	if err != nil {
		log.Error(err)
		ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
		return
	}

	tagIDs, pictureIDs, err := getClubTagIdsAndPictureIds(clubInfo.ClubID,
		clubInfo.Pic1ID, clubInfo.Pic2ID, clubInfo.Pic3ID, clubInfo.Pic4ID, clubInfo.Pic5ID, clubInfo.Pic6ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
		return
	}

	clubInfoResponse := constructClubInfoPost(clubInfo, tagIDs, pictureIDs)

	ctx.JSON(http.StatusOK, httpserver.SuccessResponse(clubInfoResponse))
}

func constructClubInfoPost(clubInfo *db.ClubInfo, tagIDs []string, pictureIDs []string) *ClubInfoPost {
	clubInfoResponse := ClubInfoPost{
		ClubID:      clubInfo.ClubID,
		Name:        clubInfo.Name,
		Website:     clubInfo.Website,
		Email:       clubInfo.Email,
		GroupLink:   clubInfo.GroupLink,
		VideoLink:   clubInfo.VideoLink,
		Published:   clubInfo.Published,
		Description: clubInfo.Description,
		LogoId:      clubInfo.LogoID,
		TagIds:      tagIDs,
		PictureIds:  pictureIDs,
	}
	return &clubInfoResponse
}

//Returns club tag id list and picture id list.
func getClubTagIdsAndPictureIds(clubID string, picIds ...string) ([]string, []string, error) {
	//Get club tags relationships from DB
	tagRelationships, err := db.GetTagRelationshipsByClubID(clubID)
	if err != nil {
		log.Error(err)
		return nil, nil, err
	}

	//Unite tag ids from relationships
	tagIDs := make([]string, 0)
	for _, tagRelationship := range tagRelationships {
		tagIDs = append(tagIDs, tagRelationship.TagID)
	}

	//Unite non-nil pic ids
	pictureIds := make([]string, 0)
	for _, picId :=range picIds {
		if picId == "" {
			break
		}
		pictureIds = append(pictureIds, picId)
	}

	return tagIDs, pictureIds, nil
}

func getAccountByUserId(ctx *gin.Context) {
	account, err := getAdminUser(ctx)
	if err != nil {
		return
	}

	if !account.IsAdmin {
		ctx.JSON(http.StatusUnauthorized, httpserver.ConstructResponse(httpserver.NO_PERMISSION, nil))
		return
	}

	userId := ctx.Param("userId")
	account, err = db.GetAccountByUserId(userId)
	if gorm.IsRecordNotFoundError(err) {
		ctx.JSON(http.StatusBadRequest, httpserver.ConstructResponse(httpserver.NOT_FOUND, nil))
		return
	}
	if err != nil {
		log.Error(err)
		ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
		return
	}
	ctx.JSON(http.StatusOK, httpserver.SuccessResponse(account))
}

func listAccounts(ctx *gin.Context) {
	account, err := getAdminUser(ctx)
	if err != nil {
		return
	}

	if !account.IsAdmin {
		ctx.JSON(http.StatusUnauthorized, httpserver.ConstructResponse(httpserver.NO_PERMISSION, nil))
		return
	}

	accounts, err := db.GetAllAccounts()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
		return
	}

	ctx.JSON(http.StatusOK, httpserver.SuccessResponse(accounts))
}

// Generate a new auth string
func genAuthString() string {
	h := sha512.New()
	h.Write([]byte(uuid.New().String()))
	str1 := hex.EncodeToString(h.Sum(nil))
	h.Write([]byte(uuid.New().String()))
	str2 := hex.EncodeToString(h.Sum(nil))
	return str1 + str2
}

//creates a club account and its club info.
func createNewClubAccount(ctx *gin.Context) {
	account, err := getAdminUser(ctx)
	if err != nil {
		return
	}

	if !account.IsAdmin {
		ctx.JSON(http.StatusUnauthorized, httpserver.ConstructResponse(httpserver.NO_PERMISSION, nil))
		return
	}

	//construct account and club info
	clubAccount := db.AdminAccount{
		AccountID:  uuid.New().String(),
		AuthString: genAuthString(),
		ClubID:     uuid.New().String(),
		IsAdmin:    false,
	}

	clubInfo := db.ClubInfo{
		ClubID: clubAccount.ClubID,
	}

	//create transaction to insert account and club info
	txDb := db.DB.Begin()
	err = clubAccount.Insert(txDb)
	if err != nil {
		txDb.Rollback()
		log.Error(err)
		ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
		return
	}
	err = clubInfo.Insert(txDb)
	if err != nil {
		txDb.Rollback()
		log.Error(err)
		ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
		return
	}
	txDb.Commit()

	//response account created
	ctx.JSON(http.StatusOK, httpserver.SuccessResponse(clubAccount))
}

func Pong(c *gin.Context) {
	c.JSON(
		http.StatusOK,
		httpserver.ConstructResponse(httpserver.SUCCESS, gin.H{
			"message": "pong",
		}))
}

type LoginPost struct {
	AuthToken string `json:"auth_token"`
}

// Handles Admin Login
func Login(c *gin.Context) {
	loginPost := new(LoginPost)
	if err := c.ShouldBindJSON(loginPost); err != nil {
		c.JSON(http.StatusUnauthorized, httpserver.ConstructResponse(httpserver.AUTH_FAILED, nil))
		log.Print(err)
		return
	}

	var Account db.AdminAccount
	if err := db.DB.Where("auth_string = ?", loginPost.AuthToken).First(&Account).Error; err != nil {
		c.JSON(http.StatusUnauthorized, httpserver.ConstructResponse(httpserver.AUTH_FAILED, nil))
		log.Print(err)
		return
	}

	// Save the username in the session
	session := sessions.Default(c)
	session.Set(USER, Account)
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
		return
	}

	c.JSON(http.StatusOK, httpserver.SuccessResponse(Account))
}

type ClubInfoPost struct {
	ClubID      string   `json:"club_id" binding:"required"`
	Name        string   `json:"name"`
	Website     string   `json:"website"`
	Email       string   `json:"email"`
	GroupLink   string   `json:"group_link"`
	VideoLink   string   `json:"video_link"`
	Published   bool     `json:"published"`
	Description string   `json:"description"`
	LogoId      string   `json:"logo_id"`
	TagIds      []string `json:"tag_ids"`
	PictureIds  []string `json:"picture_ids"`
}

type ClubInfoCountPost struct {
	ClubInfoPost
	FavouriteNum int64 `json:"favourite_num"`
	ViewNum     int64 `json:"view_num"`
}

const (
	CLUB_NAME_MAX_LEN = 50
	CLUB_PIC_MAX_NUM  = 6
	CLUB_TAG_MAX_NUM  = 4
)

//Club user updates their club info.
func updateClubInfo(ctx *gin.Context) {
	account, err := getAdminUser(ctx)
	if err != nil {
		return
	}

	//obtain and check request params
	var clubInfoPost ClubInfoPost
	if err := ctx.ShouldBindJSON(&clubInfoPost); err != nil {
		ctx.JSON(http.StatusBadRequest, httpserver.ConstructResponse(httpserver.INVALID_PARAMS, nil))
		return
	}
	if len(clubInfoPost.Name) == 0 || len(clubInfoPost.Name) > CLUB_NAME_MAX_LEN {
		ctx.JSON(http.StatusBadRequest, httpserver.ConstructResponse(httpserver.INVALID_PARAMS, nil))
		return
	}
	if len(clubInfoPost.PictureIds) > CLUB_PIC_MAX_NUM {
		ctx.JSON(http.StatusBadRequest, httpserver.ConstructResponse(httpserver.CLUB_PIC_NUM_ABOVE_LIMIT, nil))
		return
	}
	if len(clubInfoPost.TagIds) > CLUB_TAG_MAX_NUM {
		ctx.JSON(http.StatusBadRequest, httpserver.ConstructResponse(httpserver.CLUB_TAG_NUM_ABOVE_LIMIT, nil))
		return
	}
	//TODO check the params rest

	// Club tags
	if len(clubInfoPost.TagIds) > 0 {
		tags, err := db.GetClubTagsByTagIds(clubInfoPost.TagIds)
		if err != nil {
			log.Error(err)
			ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
			return
		}
		// Check for invalid tag IDs
		if len(clubInfoPost.TagIds) != len(tags) {
			ctx.JSON(http.StatusBadRequest, httpserver.ConstructResponse(httpserver.INVALID_PARAMS, nil))
			return
		}
	}
	log.Println(clubInfoPost.PictureIds)

	// Club picture upload
	if len(clubInfoPost.PictureIds) > 0 {
		dbPictureIDs, err := db.GetAccPictureIDS(account.AccountID)

		dbPictureIDsSet := set.NewSet()
		for _, accPic := range dbPictureIDs {
			dbPictureIDsSet.Add(accPic.PictureID)
		}

		if err != nil {
			log.Error(err)
			ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
			return
		}
		// Check for invalid IDs
		for _, pid := range clubInfoPost.PictureIds {
			if !dbPictureIDsSet.Contains(pid) {
				ctx.JSON(http.StatusBadRequest, httpserver.ConstructResponse(httpserver.INVALID_PARAMS, "does not contain this picture"))
				return
			}
		}
	}

	// Construct club information
	clubInfo := db.ClubInfo{
		ClubID:      clubInfoPost.ClubID,
		Name:        clubInfoPost.Name,
		Website:     clubInfoPost.Website,
		Email:       clubInfoPost.Email,
		GroupLink:   clubInfoPost.GroupLink,
		VideoLink:   clubInfoPost.VideoLink,
		Published:   clubInfoPost.Published,
		Description: clubInfoPost.Description,
		LogoID:      clubInfoPost.LogoId,
	}

	for idx, pid := range clubInfoPost.PictureIds {
		// TODO 最好不要这么写，现在为了方便先这样
		reflect.ValueOf(&clubInfo).Elem().FieldByName(fmt.Sprintf("Pic%dID", idx+1)).SetString(pid)
	}

	txDb := db.DB.Begin()
	err = clubInfo.Update(txDb)
	if err != nil {
		log.Error(err)
		txDb.Rollback()
		ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
		return
	}

	// Update club tags relationship
	if len(clubInfoPost.TagIds) > 0 {
		// Clean up old associations
		err := db.CleanAllTags(txDb, clubInfoPost.ClubID)
		if err != nil {
			log.Error(err)
			txDb.Rollback()
			ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
			return
		}

		// Then insert latest relationship
		for _, tagId := range clubInfoPost.TagIds {
			relationship := db.ClubTagRelationship{
				ClubID: clubInfo.ClubID,
				TagID:  tagId,
			}
			err := relationship.Insert(txDb)
			if err != nil {
				log.Error(err)
				txDb.Rollback()
				ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
				return
			}
		}
	}

	txDb.Commit()
	ctx.JSON(http.StatusOK, httpserver.SuccessResponse(nil))
}

type UploadPicResponse struct {
	Pid string `json:"pid"`
}

// Upload a picture and return an ID
func uploadSinglePicture(ctx *gin.Context) {
	account, err := getAdminUser(ctx)
	if err != nil {
		return
	}

	file, err := ctx.FormFile("file")
	if err != nil {
		log.Error(err)
		ctx.JSON(http.StatusBadRequest, httpserver.ConstructResponse(httpserver.INVALID_PARAMS, nil))
		return
	}

	// ensures what's uploaded is a picture
	if !strings.HasSuffix(file.Filename, ".jpg") && !strings.HasSuffix(file.Filename, ".jpeg") {
		log.Println(file.Filename)
		log.Errorf("Uploaded file is %v. The extension does noe match jpg ir jpeg", file.Filename)
		ctx.JSON(http.StatusBadRequest, httpserver.ConstructResponse(httpserver.UPLOAD_TYPE_NOT_SUPPORTED, nil))
		return
	}

	MaxFileSize := int64(1 << 20)
	// Check file size limit
	if file.Size > MaxFileSize {
		log.Errorf("File size is %v MB > than 1MB", file.Size/1<<20)
		ctx.JSON(http.StatusBadRequest, httpserver.ConstructResponse(httpserver.PIC_TOO_LARGE, nil))
		return
	}

	fileUUID := uuid.New().String()
	fileName := fileUUID + path.Ext(file.Filename)
	basePath := path.Join(globalConfig.General.PictureStoragePath, fileName)

	err = ctx.SaveUploadedFile(file, basePath)
	if err != nil {
		log.Error(err)
		ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
		return
	}

	// Sanity check done. Save picture info into db,
	pictureEntry := db.AccountPicture{
		AccountID:   account.AccountID,
		PictureID:   fileUUID,
		PictureName: fileName,
	}

	err = db.DB.Create(&pictureEntry).Error
	if err != nil {
		log.Error(err)
		ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
		return
	}

	ctx.JSON(http.StatusOK, httpserver.SuccessResponse(UploadPicResponse{Pid: fileUUID}))
}
