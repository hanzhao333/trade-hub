package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// TickerHub 简易行情推送（演示用；生产由 Indexer / 行情服务写入 Redis 再广播）
type TickerHub struct {
	mu      sync.Mutex
	clients map[*websocket.Conn]struct{}
}

func NewTickerHub() *TickerHub {
	return &TickerHub{clients: make(map[*websocket.Conn]struct{})}
}

func (h *TickerHub) HandleWS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	h.register(conn)
	defer h.unregister(conn)

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (h *TickerHub) register(conn *websocket.Conn) {
	h.mu.Lock()
	h.clients[conn] = struct{}{}
	h.mu.Unlock()
}

func (h *TickerHub) unregister(conn *websocket.Conn) {
	h.mu.Lock()
	delete(h.clients, conn)
	h.mu.Unlock()
	_ = conn.Close()
}

type tickerMsg struct {
	PoolID string `json:"pool_id"`
	Price  string `json:"price"`
	Ts     int64  `json:"ts"`
}

// StartDemoBroadcast 每 3 秒推送 demo 价格（学习用；第 3 周可改为读 Redis）
func (h *TickerHub) StartDemoBroadcast() {
	ticker := time.NewTicker(3 * time.Second)
	prices := []tickerMsg{
		{PoolID: "SOL-USDC", Price: "142.50"},
		{PoolID: "ETH-USDC", Price: "3200.00"},
	}
	i := 0
	go func() {
		for range ticker.C {
			msg := prices[i%len(prices)]
			msg.Ts = time.Now().Unix()
			i++
			h.broadcast(msg)
		}
	}()
}

func (h *TickerHub) broadcast(msg tickerMsg) {
	data, _ := json.Marshal(msg)
	h.mu.Lock()
	defer h.mu.Unlock()
	for conn := range h.clients {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("ws write: %v", err)
			_ = conn.Close()
			delete(h.clients, conn)
		}
	}
}
