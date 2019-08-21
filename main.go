package main

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"log"
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

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(
			http.StatusOK,
			gin.H{
				"message": "pong",
			})
	})

	router.POST("/login", AdminLogin)
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
