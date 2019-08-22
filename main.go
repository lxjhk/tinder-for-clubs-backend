package main

import (
	"crypto/sha512"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/gin-contrib/secure"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	log "github.com/sirupsen/logrus"
	"net/http"
	"path"
	"strings"
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

	// Disable inline scripts
	router.Use(secure.New(secure.Config{
		ContentSecurityPolicy: "default-src 'self'",
	}))

	router.Use(CORSMiddleware())

	store := memstore.NewStore([]byte(uuid.New().String()))
	router.Use(sessions.Sessions("AdminSession", store))

	initializeRoutes()
	registerObjToGob()
	err := router.Run() // listen and serve on 0.0.0.0:8080
	common.ErrFatalLog(err)
}

func registerObjToGob() {
	gob.Register(db.AdminAccount{})
}

// CORS allow domains so that we can serve the website from different subdomains.
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

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

	// Admin only endpoints
	router.POST("/admin/account/create", createNewClubAccount)
	router.GET("/admin/account/all", listAccounts)
	router.GET("/admin/account/user/:userId", getAccountByUserId)

	// Club manager endpoints
	//TODO: router.GET("/account", getUser)
	router.POST("/club/uploadpicture", uploadSinglePicture)
	router.POST("/club/info", updateClubInfo)
	router.GET("/club/info", getClubInfo)
	// TODO router.GET("/club/tags", getTags)

	// Public endpoints
	router.GET("/static/clubphoto/:pictureID", serveStaticPicture)
}

func serveStaticPicture(ctx *gin.Context) {

}

// Get User information from request context.
func getUser(ctx *gin.Context) (*db.AdminAccount, error) {
	session := sessions.Default(ctx)
	result := session.Get(USER)
	if result == nil {
		ctx.JSON(http.StatusUnauthorized, httpserver.ConstructResponse(httpserver.NOT_AUTHORIZED, nil))
		return nil, errors.New("user invalid")
	}

	user := result.(db.AdminAccount)

	return &user, nil
}

// Returns the club info of current login user.
func getClubInfo(ctx *gin.Context) {
	account, err := getUser(ctx)
	if err != nil {
		return
	}

	clubInfo, err := db.GetClubInfoByClubId(account.ClubID)
	if err != nil {
		log.Error(err)
		ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
		return
	}

	ctx.JSON(http.StatusOK, httpserver.SuccessResponse(clubInfo))
}

func getAccountByUserId(ctx *gin.Context) {
	account, err := getUser(ctx)
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
		ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
		return
	}
	ctx.JSON(http.StatusOK, httpserver.SuccessResponse(account))
}

func listAccounts(ctx *gin.Context) {
	account, err := getUser(ctx)
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
	account, err := getUser(ctx)
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

type updateClubInfoRequest struct {
	ClubID      string `json:"club_id" binding:"required"`
	Name        string `json:"name"`
	Website     string `json:"website"`
	Email       string `json:"email"`
	GroupLink   string `json:"group_link"`
	VideoLink   string `json:"video_link"`
	Published   bool   `json:"published"`
	Description string `json:"description"`
	Tags        []string `json:"tags"`
	PictureIds   []string `json:"picture_ids"`
}

//Club user updates their club info.
func updateClubInfo(ctx *gin.Context) {
	account, err := getUser(ctx)
	if err != nil {
		return
	}

	_ = account

	// TODO 拿到 Tag List， 拿到 Picture ID List， 检查是否存在数据库
	// TODO If 都存在 OK， 任何的不存在， 直接返回错误

	////get the source club info from DB.
	//sourceClub, err := db.GetClubInfoByClubId(account.ClubID)
	//if err != nil {
	//	ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
	//	return
	//}
	//
	////get request params from json
	//var targetClub db.ClubInfo
	//if err := ctx.ShouldBindJSON(&targetClub); err != nil {
	//	ctx.JSON(http.StatusBadRequest, httpserver.ConstructResponse(httpserver.INVALID_PARAMS, nil))
	//	return
	//}
	//
	////ensures user is updating his own club
	//if sourceClub.ClubID != targetClub.ClubID {
	//	ctx.JSON(http.StatusBadRequest, httpserver.ConstructResponse(httpserver.NO_PERMISSION, nil))
	//	return
	//}
	//
	////removes decreased pic info of target club
	//removeDecreasedPics(sourceClub, &targetClub)
	//
	////update club info
	//if err = targetClub.Update(); err != nil {
	//	ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
	//	return
	//}
	//
	//ctx.JSON(http.StatusOK, httpserver.SuccessResponse(nil))
}

func uploadSinglePicture(ctx *gin.Context) {
	account, err := getUser(ctx)
	if err != nil {
		return
	}

	_ = account
	//get multipart file from the multipart form data.
	multipart, err := ctx.MultipartForm()
	if err != nil {
		log.Error(err)
		ctx.JSON(http.StatusBadRequest, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
		return
	}
	//returns error when not one picture.
	files := multipart.File["image"]
	if len(files) > 1 || len(files) == 0 {
		ctx.JSON(http.StatusBadRequest, httpserver.ConstructResponse(httpserver.PIC_NUM_NOT_SUPPORTED, nil))
		return
	}

	//ensures what uploaded is a picture
	file := files[0]
	fileType := file.Header.Get("Content-Type")
	if !strings.HasPrefix(fileType, "image") {
		log.Error(err)
		ctx.JSON(http.StatusBadRequest, httpserver.ConstructResponse(httpserver.UPLOAD_TYPE_NOT_SUPPORTED, nil))
		return
	}

	//All is detected, save picture info into db,
	// rolls back DB transaction when fails to save picture to the disk.
	picture := db.AccountPicture{
		AccountID:   account.AccountID,
		PictureID:   uuid.New().String(),
		PictureName: uuid.New().String() + path.Ext(file.Filename),
	}

	txDb := db.DB.Begin()
	err = picture.Insert(txDb)
	if err != nil {
		log.Error(err)
		txDb.Rollback()
		ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
		return
	}

	//save picture to disk
	err = ctx.SaveUploadedFile(file, fmt.Sprintf("%s/%s", globalConfig.General.PictureStoragePath, picture.PictureName))
	if err != nil {
		log.Error(err)
		txDb.Rollback()
		ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
		return
	}
	txDb.Commit()

	ctx.JSON(http.StatusOK, httpserver.SuccessResponse(picture.PictureID))

	// TODO 接受上传的图片，检查图片（尺寸和大小—），成功并且返回 Picture ID, Add to DB [userID, PictureID, PictureNameOnDisk]
	// TODO PictureID = UUID  NAME = {UUID}.jpg
}
