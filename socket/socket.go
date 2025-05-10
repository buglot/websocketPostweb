package socket

import (
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/buglot/postAPI/lib"
	"github.com/buglot/postAPI/orm"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/net/websocket"
	"gorm.io/gorm"
)

type UserWs struct {
	User   *websocket.Conn
	Friend map[uint]*websocket.Conn
}

var onlineUsers = struct {
	sync.RWMutex
	Users map[uint]UserWs
}{Users: make(map[uint]UserWs)}

func parseToken(tokenStr string) (uint, error) {
	hmacSampleSecret := []byte(os.Getenv("JWT_SECRAT_KEY"))
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return hmacSampleSecret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if uidFloat, ok := claims["userID"]; ok {
			return lib.AnyToUInt(uidFloat), nil
		}
	}
	return 0, fmt.Errorf("invalid token claims")
}
func SetupSocketRoutes(mux *http.ServeMux, db *gorm.DB) {
	mux.Handle("/ws", websocket.Handler(func(ws *websocket.Conn) {
		token := ws.Request().URL.Query().Get("token")
		userID, _ := parseToken(token)
		onlineUsers.Lock()
		if _, exists := onlineUsers.Users[uint(userID)]; !exists {
			onlineUsers.Users[uint(userID)] = &UserWs{
				User:    ws,
				Friends: []*websocket.Conn{}, // Initialize with empty friends list
			}
		} else {
			// Update existing user with new WebSocket connection
			onlineUsers.Users[uint(userID)].User = ws
		}
		onlineUsers.Unlock()

		fmt.Printf("User %d connected\n", userID)

		for {
			var msg string
			if err := websocket.Message.Receive(ws, &msg); err != nil {
				break
			}
			fmt.Println("Message:", msg)

		}

		onlineUsers.Lock()
		delete(onlineUsers.Users, uint(userID))
		onlineUsers.Unlock()
		fmt.Printf("User %d disconnected\n", userID)
	}))
}

func GetFolloweesOnlineStatus(db *gorm.DB, userID uint) map[uint]bool {
	var follows []orm.Follow
	db.Preload("Followee").Where("follower_id = ?", userID).Find(&follows)
	result := make(map[uint]bool)
	onlineUsers.RLock()
	for _, f := range follows {
		_, ok := onlineUsers.Users[f.FolloweeID]
		result[f.FolloweeID] = ok
	}
	onlineUsers.RUnlock()
	return result
}
