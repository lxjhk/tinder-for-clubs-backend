package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	log "github.com/sirupsen/logrus"
	"net/http"
	"path"
	"strings"
	"tinder-for-clubs-backend/common"
	"tinder-for-clubs-backend/config"
	"tinder-for-clubs-backend/db"
)

var router *gin.Engine

var conf *config.GlobalConfiguration

func main() {
	// Reading configuration file
	globalConfig := config.ReadConfig()

	conf = &globalConfig

	// Setting up database connection
	db.Init(globalConfig.DBCredential)

	//Deferred Closed
	defer db.Close()


	// Initialise HTTP framework
	router = gin.Default()
	initializeRoutes()
	err := router.Run() // listen and serve on 0.0.0.0:8080
	common.ErrFatalLog(err)
}

func initializeRoutes() {

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(
			http.StatusOK,
			gin.H{
				"message": "pong",
			})
	})

	router.POST("/uploadOne", uploadSinglePicture)

}

type ResResult struct {
	Code int64 `json:"code"`
	Msg string `json:"msg"`
	Data interface{} `json:"data"`
}

func uploadSinglePicture(ctx *gin.Context) {
	//获取用户信息，获取其club id
    var clubId = ""
	//获取club信息
	clubInfo, err := db.GetClubInfoById(clubId)
	if err != nil {
		log.WithFields(log.Fields{"event": "uploadPicture"}).Error(err)
		ctx.JSON(http.StatusInternalServerError, ResResult{
			Code: 2001,
			Msg: "internal error",
		})
	}

	pid := ctx.Query("pid")
	switch pid {
	case "1":
		break
	case "2":
		break
	case "3":
		break
	case "4":
		break
	case "5":
		break
	case "6":
		break
	default:
		ctx.JSON(http.StatusBadRequest, ResResult{
			Code: 2002,
			Msg: "invalid params",
		})
	}

	//获取上传图片
	multipart, err := ctx.MultipartForm()
	if err != nil {
		log.WithFields(log.Fields{"event": "uploadPicture"}).Error(err)
		ctx.JSON(http.StatusInternalServerError, ResResult{
			Code: 2001,
			Msg: "internal error",
		})
	}

	files := multipart.File["image"]
	if len(files) > 1 || len(files) == 0 {
		log.WithFields(log.Fields{"event": "uploadPicture"}).Error(err)
		ctx.JSON(http.StatusBadRequest, ResResult{
			Code: 2002,
			Msg: "invalid params",
		})
	}

	picUid := uuid.Must(uuid.NewV4()).String()

	for _, file :=range files {
		fileType := file.Header.Get("Content-Type")
		if !strings.HasPrefix(fileType, "image"){
			ctx.JSON(http.StatusBadRequest, ResResult{
				Code: 2004,
				Msg: "not picture",
			})
		}
		ext := path.Ext(file.Filename)
		err := ctx.SaveUploadedFile(file, fmt.Sprintf("%s/%s.%s", conf.General.PictureStoragePath, picUid, ext))
		if err != nil {
			log.WithFields(log.Fields{"event": "uploadPicture"}).Error(err)
			ctx.JSON(http.StatusInternalServerError, ResResult{
				Code: 2001,
				Msg: "internal error",
			})
		}
	}

	ctx.JSON(http.StatusOK, ResResult{
		Code: 1000,
		Msg: "success",
	})
}


