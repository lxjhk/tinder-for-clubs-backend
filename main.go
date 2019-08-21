package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
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
	router.POST("/clubInfo/update",updateClubInfo)
}

func updateClubInfo(ctx *gin.Context) {
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

	//图片id检查


}

type ResResult struct {
	Code int64 `json:"code"`
	Msg string `json:"msg"`
	Data interface{} `json:"data"`
}

func uploadSinglePicture(ctx *gin.Context) {
	pid := ctx.Query("pid")
	if !(pid == "1" || pid == "2" || pid == "3" || pid == "4" || pid == "5" || pid == "6") {
		ctx.JSON(http.StatusBadRequest, ResResult{
			Code: 2002,
			Msg: "invalid params",
		})
	}

	//上传图片
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

	newPicUid := uuid.Must(uuid.NewV4()).String()
	for _, file :=range files {
		fileType := file.Header.Get("Content-Type")
		if !strings.HasPrefix(fileType, "image"){
			ctx.JSON(http.StatusBadRequest, ResResult{
				Code: 2004,
				Msg: "not picture",
			})
		}
		ext := path.Ext(file.Filename)
		newPicUid = fmt.Sprintf("%s.%s", newPicUid, ext)
		err := ctx.SaveUploadedFile(file, fmt.Sprintf("%s/%s", conf.General.PictureStoragePath, newPicUid))
		if err != nil {
			log.WithFields(log.Fields{"event": "uploadPicture"}).Error(err)
			ctx.JSON(http.StatusInternalServerError, ResResult{
				Code: 2001,
				Msg: "internal error",
			})
		}
	}

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

	//原图删除检查
	switch pid {
	case "1":
		removePicIfExist(clubInfo.Pic1ID)
		clubInfo.Pic1ID = newPicUid
		break
	case "2":
		removePicIfExist(clubInfo.Pic2ID)
		clubInfo.Pic2ID = newPicUid
		break
	case "3":
		removePicIfExist(clubInfo.Pic3ID)
		clubInfo.Pic3ID = newPicUid
		break
	case "4":
		removePicIfExist(clubInfo.Pic4ID)
		clubInfo.Pic4ID = newPicUid
		break
	case "5":
		removePicIfExist(clubInfo.Pic5ID)
		clubInfo.Pic5ID = newPicUid
		break
	case "6":
		removePicIfExist(clubInfo.Pic6ID)
		clubInfo.Pic6ID = newPicUid
		break
	}

	//更新最新的club 信息
	err = clubInfo.UpdateAllPicIds()
	if err != nil {
		removePicIfExist(newPicUid)
		ctx.JSON(http.StatusInternalServerError, ResResult{
			Code: 2001,
			Msg: "internal error",
		})
	}

	ctx.JSON(http.StatusOK, ResResult{
		Code: 1000,
		Msg: "success",
	})
}

func removePicIfExist(picName string) {
	if picName != "" {
		os.Remove(fmt.Sprintf("%s/%s", conf.General.PictureStoragePath, picName))
	}
}


