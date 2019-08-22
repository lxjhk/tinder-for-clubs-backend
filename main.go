package main

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
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
	AUTH_ACCOUNT = "AUTH_ACCOUNT"
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
	store := memstore.NewStore([]byte(uuid.New().String()))
	router.Use(sessions.Sessions("AdminSession", store))

	initializeRoutes()
	err := router.Run() // listen and serve on 0.0.0.0:8080
	common.ErrFatalLog(err)
}

func initializeRoutes() {
	//Initialize authentication middleware
	router.Use(authInterceptor())

	//Ping endpoint for testing
	router.GET("/ping", Pong)

	//For admin and club users to login
	router.POST("/admin/login", AdminLogin)

	//For admin to manage account
	router.POST("/admin/account/create", createNewClubAccount)
	router.GET("/admin/account", listAccounts)
	router.GET("/admin/account/:userId", getAccount)

	//For club account to upload picture and update club info.
	router.POST("/admin/clubInfo/uploadPic", uploadSinglePicture)
	router.POST("/admin/clubInfo", updateClubInfo)
	router.GET("/admin/clubInfo", getClubInfo)
}

func authInterceptor() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		//user is signing in
		url := ctx.Request.RequestURI
		if "/admin/login" == url {
			ctx.Next()
			return
		}

		//get user id from session to get its detail info from DB.
		session := sessions.Default(ctx)
		result := session.Get(AUTH_ACCOUNT)
		if result == nil {
			ctx.JSON(http.StatusUnauthorized, httpserver.ConstructResponse(httpserver.NOT_AUTHORIZED, nil))
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}

// Get User information from session
func getAuthAccountFromSession(ctx *gin.Context) *db.AdminAccount {
	auth, ok := ctx.Get(AUTH_ACCOUNT)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, httpserver.ConstructResponse(httpserver.AUTH_FAILED, nil))
		return nil
	}
	account := auth.(*db.AdminAccount)
	return account
}

//目前根据登录club账户直接获取对应club info
func getClubInfo(ctx *gin.Context) {
	account := getAuthAccountFromSession(ctx)

	clubInfo, err := db.GetClubInfoByClubId(account.ClubID)
	if gorm.IsRecordNotFoundError(err) {
		ctx.JSON(http.StatusOK, httpserver.ConstructResponse(httpserver.NOT_FOUND, nil))
		return
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
		return
	}

	ctx.JSON(http.StatusOK, httpserver.SuccessResponse(clubInfo))
}

func getAccount(ctx *gin.Context) {
	account := getAuthAccountFromSession(ctx)

	if !account.IsAdmin {
		ctx.JSON(http.StatusUnauthorized, httpserver.ConstructResponse(httpserver.NO_PERMISSION, nil))
		return
	}

	if !account.IsAdmin {
		ctx.JSON(http.StatusUnauthorized, httpserver.ConstructResponse(httpserver.NO_PERMISSION, nil))
		return
	}

	userId := ctx.Param("userId")
	account, err := db.GetAccountByUserId(userId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
		return
	}
	ctx.JSON(http.StatusOK, httpserver.SuccessResponse(account))
}

func listAccounts(ctx *gin.Context) {
	account := getAuthAccountFromSession(ctx)
	if !account.IsAdmin {
		ctx.JSON(http.StatusUnauthorized, httpserver.ConstructResponse(httpserver.NO_PERMISSION, nil))
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
	account := getAuthAccountFromSession(ctx)
	if !account.IsAdmin {
		ctx.JSON(http.StatusUnauthorized, httpserver.ConstructResponse(httpserver.NO_PERMISSION, nil))
		return
	}

	authString := genAuthString()

	//construct account and club info
	clubAccount := db.AdminAccount{
		UserID:     uuid.New().String(),
		AuthString: authString,
		ClubID:     uuid.New().String(),
		IsAdmin:    false,
	}
	clubInfo := db.ClubInfo{
		ClubID: clubAccount.ClubID,
	}

	//create transaction to insert account and club info
	txDb := db.DB.Begin()
	err := clubAccount.Insert(txDb)
	if err != nil {
		txDb.Rollback()
		ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
		return
	}
	err = clubInfo.Insert(txDb)
	if err != nil {
		txDb.Rollback()
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
func AdminLogin(c *gin.Context) {
	loginPost := new(LoginPost)
	if err := c.ShouldBindJSON(loginPost); err != nil {
		c.JSON(http.StatusUnauthorized, httpserver.ConstructResponse(httpserver.AUTH_FAILED, nil))
		log.Print(err)
		return
	}

	var AdminAccount db.AdminAccount
	if err := db.DB.Where("auth_string = ?", loginPost.AuthToken).First(&AdminAccount).Error; err != nil {
		c.JSON(http.StatusUnauthorized, httpserver.ConstructResponse(httpserver.AUTH_FAILED, nil))
		log.Print(err)
		return
	}

	// Save the username in the session
	session := sessions.Default(c)
	session.Set(AUTH_ACCOUNT, AdminAccount)
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
		return
	}

	c.JSON(http.StatusOK, httpserver.SuccessResponse(nil))
}

//Club user updates their club info.
func updateClubInfo(ctx *gin.Context) {
	account := getAuthAccountFromSession(ctx)
	//get the source club info from DB.
	sourceClub, err := db.GetClubInfoByClubId(account.ClubID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
		return
	}

	var targetClub db.ClubInfo
	if err := ctx.ShouldBindJSON(&targetClub); err != nil {
		ctx.JSON(http.StatusBadRequest, httpserver.ConstructResponse(httpserver.INVALID_PARAMS, nil))
		return
	}

	//ensures user is updating his own club
	if sourceClub.ClubID != targetClub.ClubID {
		ctx.JSON(http.StatusBadRequest, httpserver.ConstructResponse(httpserver.NO_PERMISSION, nil))
		return
	}

	//update club info
	if err = targetClub.Update(); err != nil {
		ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
		return
	}

	ctx.JSON(http.StatusOK, httpserver.SuccessResponse(nil))
}

func uploadSinglePicture(ctx *gin.Context) {
	account := getAuthAccountFromSession(ctx)

	pid := ctx.Query("pid")
	if !(pid == "1" || pid == "2" || pid == "3" || pid == "4" || pid == "5" || pid == "6") {
		ctx.JSON(http.StatusBadRequest, httpserver.ConstructResponse(httpserver.INVALID_PARAMS, nil))
		return
	}

	//get multipart file from the multipart form data.
	multipart, err := ctx.MultipartForm()
	if err != nil {
		log.WithFields(log.Fields{"event": "uploadPicture"}).Error(err)
		ctx.JSON(http.StatusBadRequest, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
		return
	}
	//returns error when not one picture.
	files := multipart.File["image"]
	if len(files) > 1 || len(files) == 0 {
		ctx.JSON(http.StatusBadRequest, httpserver.ConstructResponse(httpserver.PIC_NUM_NOT_SUPPORTED, nil))
		return
	}

	//ensures what uploaded is a picture and save it to specified path.
	file := files[0]
	fileType := file.Header.Get("Content-Type")
	if !strings.HasPrefix(fileType, "image") {
		ctx.JSON(http.StatusBadRequest, httpserver.ConstructResponse(httpserver.UPLOAD_TYPE_NOT_SUPPORTED, nil))
		return
	}
	extName := path.Ext(file.Filename)
	picUid := fmt.Sprintf("%s.%s", uuid.New().String(), extName)
	err = ctx.SaveUploadedFile(file, fmt.Sprintf("%s/%s", globalConfig.General.PictureStoragePath, picUid))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SAVE_PICTURE_FAILED, nil))
		return
	}

	//get the source club info from DB.
	clubInfo, err := db.GetClubInfoByClubId(account.ClubID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
		return
	}

	//ensures the source picture has been removed when set the target picture
	switch pid {
	case "1":
		removePicIfExist(clubInfo.Pic1ID)
		clubInfo.Pic1ID = picUid
		break
	case "2":
		removePicIfExist(clubInfo.Pic2ID)
		clubInfo.Pic2ID = picUid
		break
	case "3":
		removePicIfExist(clubInfo.Pic3ID)
		clubInfo.Pic3ID = picUid
		break
	case "4":
		removePicIfExist(clubInfo.Pic4ID)
		clubInfo.Pic4ID = picUid
		break
	case "5":
		removePicIfExist(clubInfo.Pic5ID)
		clubInfo.Pic5ID = picUid
		break
	case "6":
		removePicIfExist(clubInfo.Pic6ID)
		clubInfo.Pic6ID = picUid
		break
	}

	//update the picture info of current club to the latest.
	err = clubInfo.UpdateAllPicIds()
	if err != nil {
		//remove the uploaded picture when error
		removePicIfExist(picUid)
		ctx.JSON(http.StatusInternalServerError, httpserver.ConstructResponse(httpserver.SYSTEM_ERROR, nil))
		return
	}

	ctx.JSON(http.StatusOK, httpserver.SuccessResponse(nil))
}

func removePicIfExist(picName string) {
	if picName != "" {
		err := os.Remove(fmt.Sprintf("%s/%s", globalConfig.General.PictureStoragePath, picName))
		log.Print(err)
	}
}
