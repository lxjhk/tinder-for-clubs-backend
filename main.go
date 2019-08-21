package main

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	log "github.com/sirupsen/logrus"
	"net/http"
	"tinder-for-clubs-backend/common"
	"tinder-for-clubs-backend/config"
	"tinder-for-clubs-backend/db"
	"tinder-for-clubs-backend/httpserver"
)

var router *gin.Engine

const (
	SESSION_USER_KEY = "SESSION_USER_KEY"
)

func main() {
	// Reading configuration file
	globalConfig := config.ReadConfig()

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
	router.GET("/ping", Pong)
	router.POST("/login", AdminLogin)
	router.POST("/uploadOne", uploadSinglePicture)

}

func Pong(c *gin.Context) {
	c.JSON(
		http.StatusOK,
		httpserver.ConstructResponse(httpserver.SUCCESS, gin.H{
			"message": "pong",
		}))
}

// Handles Admin Login
func AdminLogin(c *gin.Context) {
	session := sessions.Default(c)
	authToken := c.PostForm("authToken")

	var AdminAccount db.AdminAccount
	if err := db.DB.Where("auth_string = ?", authToken).First(&AdminAccount).Error; err != nil {
		c.JSON(http.StatusUnauthorized, httpserver.ConstructResponse(httpserver.AUTH_FAILED, nil))
		log.Print(err)
		return
	}

	// Save the username in the session
	session.Set(SESSION_USER_KEY, AdminAccount.ID) // In real world usage you'd set this to the users ID
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
		return
	}
	c.JSON(http.StatusOK, httpserver.SuccessResponse(nil))

}

func uploadSinglePicture(ctx *gin.Context) {
	////获取用户信息，获取其club id
	//var clubId = ""
	////获取club信息
	//clubInfo, err := db.GetClubInfoById(clubId)
	//if err != nil {
	//	log.WithFields(log.Fields{"event": "uploadPicture"}).Error(err)
	//	ctx.JSON(http.StatusInternalServerError, ResResult{
	//		Code: 2001,
	//		Msg: "internal error",
	//	})
	//}
	//
	//pid := ctx.Query("pid")
	//switch pid {
	//case "1":
	//	break
	//case "2":
	//	break
	//case "3":
	//	break
	//case "4":
	//	break
	//case "5":
	//	break
	//case "6":
	//	break
	//default:
	//	ctx.JSON(http.StatusBadRequest, ResResult{
	//		Code: 2002,
	//		Msg: "invalid params",
	//	})
	//}
	//
	////获取上传图片
	//multipart, err := ctx.MultipartForm()
	//if err != nil {
	//	log.WithFields(log.Fields{"event": "uploadPicture"}).Error(err)
	//	ctx.JSON(http.StatusInternalServerError, ResResult{
	//		Code: 2001,
	//		Msg: "internal error",
	//	})
	//}
	//
	//files := multipart.File["image"]
	//if len(files) > 1 || len(files) == 0 {
	//	log.WithFields(log.Fields{"event": "uploadPicture"}).Error(err)
	//	ctx.JSON(http.StatusBadRequest, ResResult{
	//		Code: 2002,
	//		Msg: "invalid params",
	//	})
	//}
	//
	//picUid := uuid.Must(uuid.NewV4()).String()
	//
	//for _, file :=range files {
	//	fileType := file.Header.Get("Content-Type")
	//	if !strings.HasPrefix(fileType, "image"){
	//		ctx.JSON(http.StatusBadRequest, ResResult{
	//			Code: 2004,
	//			Msg: "not picture",
	//		})
	//	}
	//	ext := path.Ext(file.Filename)
	//	err := ctx.SaveUploadedFile(file, fmt.Sprintf("%s/%s.%s", conf.General.PictureStoragePath, picUid, ext))
	//	if err != nil {
	//		log.WithFields(log.Fields{"event": "uploadPicture"}).Error(err)
	//		ctx.JSON(http.StatusInternalServerError, ResResult{
	//			Code: 2001,
	//			Msg: "internal error",
	//		})
	//	}
	//}
	//
	//ctx.JSON(http.StatusOK, ResResult{
	//	Code: 1000,
	//	Msg: "success",
	//})
}
