package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/buglot/postAPI/lib"
	"github.com/buglot/postAPI/orm"
	"github.com/buglot/websocketPostweb/middleware"
	"github.com/buglot/websocketPostweb/model"
	"github.com/buglot/websocketPostweb/socket"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	model.InitDB()
	if err != nil {
		fmt.Println("Error loading .env file")
	}
	go func() {
		mux := http.NewServeMux()
		socket.SetupSocketRoutes(mux, model.Db)
		fmt.Println("WebSocket server running at 0.0.0.0:8081")
		if err := http.ListenAndServe("0.0.0.0:8081", mux); err != nil {
			log.Fatal("WebSocket server error:", err)
		}
	}()
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	Authen := router.Group("/auth", middleware.Auth())
	Authen.GET("/friend", func(ctx *gin.Context) {
		userID := lib.AnyToUInt(ctx.MustGet("userID"))

		var follows []orm.Follow
		model.Db.Preload("Followee").Preload("User").Where("follower_id = ?", userID).Find(&follows)

		statusMap := socket.GetFolloweesOnlineStatus(model.Db, userID)

		var result []gin.H
		for _, f := range follows {
			followee := f.Followee
			result = append(result, gin.H{
				"username": followee.Username,
				"avatar":   followee.Avatar,
				"online":   statusMap[followee.ID],
				"url":      followee.Url,
			})
		}
		ctx.JSON(200, result)
	})
	router.Run("0.0.0.0:8082")
}
