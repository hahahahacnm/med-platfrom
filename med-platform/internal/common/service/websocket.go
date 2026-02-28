package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WsHub 管理所有活跃的 WebSocket 连接
type WsHub struct {
	Clients    map[uint]*websocket.Conn // userID -> connection
	Register   chan *WsClient           // 注册通道
	Unregister chan *WsClient           // 注销通道
	mu         sync.RWMutex
}

type WsClient struct {
	UserID uint
	Conn   *websocket.Conn
}

// 全局唯一的 Hub 实例
var Hub = &WsHub{
	Clients:    make(map[uint]*websocket.Conn),
	Register:   make(chan *WsClient),
	Unregister: make(chan *WsClient),
}

// 升级器：处理 HTTP 协议升级为 WebSocket
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许跨域
	},
}

// 🔥 核心修复：添加 Run 方法处理并发注册与注销
func (h *WsHub) Run() {
	fmt.Println("🚀 [WebSocket] 事件中枢已启动")
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.Clients[client.UserID] = client.Conn
			h.mu.Unlock()
			fmt.Printf("🔵 [WebSocket] 用户 %d 已上线\n", client.UserID)

		case client := <-h.Unregister:
			h.mu.Lock()
			if conn, ok := h.Clients[client.UserID]; ok {
				if conn == client.Conn { // 确保关闭的是当前的连接
					conn.Close()
					delete(h.Clients, client.UserID)
					fmt.Printf("🔴 [WebSocket] 用户 %d 已离线\n", client.UserID)
				}
			}
			h.mu.Unlock()
		}
	}
}

// SendToUser 向指定用户发送实时消息
func (h *WsHub) SendToUser(userID uint, message interface{}) {
	h.mu.RLock()
	conn, ok := h.Clients[userID]
	h.mu.RUnlock()

	if ok {
		msgBytes, _ := json.Marshal(message)
		err := conn.WriteMessage(websocket.TextMessage, msgBytes)
		if err != nil {
			fmt.Printf("⚠️ [WebSocket] 向用户 %d 推送失败: %v\n", userID, err)
			h.Unregister <- &WsClient{UserID: userID, Conn: conn}
		}
	}
}

// WsHandler WebSocket 路由处理函数
func WsHandler(c *gin.Context) {
	uidStr := c.Query("uid")
	userID, _ := strconv.Atoi(uidStr)

	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的用户ID"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Printf("❌ [WebSocket] 升级失败: %v\n", err)
		return
	}

	client := &WsClient{UserID: uint(userID), Conn: conn}
	Hub.Register <- client

	// 阻塞监听连接关闭状态
	go func() {
		defer func() {
			Hub.Unregister <- client
		}()
		for {
			// 持续读取以检测连接是否存活
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}()
}