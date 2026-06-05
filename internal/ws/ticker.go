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
	// mu 是锁，保护 clients 这张表
	mu sync.Mutex
	// 一个连接 = 通常一个浏览器标签页 / 一个 WS 客户端
	clients map[*websocket.Conn]struct{}
}

func NewTickerHub() *TickerHub {
	// Mutex 零值就能用，map 零值是 nil，写也可以，写的话就是sync.Mutex

	// // 情况 1：没 make（零值）
	// ** var m map[string]int   // m == nil
	// m["a"] = 1 → panic 崩溃

	// // 情况 2：make 过了
	// ** m := make(map[string]int)   // m 是空 map，但不是 nil
	// m["a"] = 1 正常

	// 如果不用make初始化，那么下面的代码h.clients[conn] = struct{}{}就会报错了
	return &TickerHub{clients: make(map[*websocket.Conn]struct{})}
}

// 注册WS连接，并且将连接添加到clients中
func (h *TickerHub) HandleWS(c *gin.Context) {
	// 将当次请求升级成 WebSocket 长连接
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	h.register(conn)
	defer h.unregister(conn)
	// for { } // 无限循环，只能靠 break/return/panic 退出
	// 如果一直可以读到消息，那么就会一直循环下去
	// go会为本次的这个HandleWS函数，创建一个goroutine，
	// 通过一直ReadMessage来判断客户端是否能成功收到消息
	// 直到客户端关闭连接，或者出现错误，才会退出循环
	// 然后unregister，将这个客户端从clients中删除
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

// json:"pool_id转成json的时候用pool_id传给前端
// 反方向也一样：前端 POST 过来 {"pool_id":"..."} 时，Gin
// 用 json:"pool_id" 绑回 Go 的 PoolID 字段
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
			// 没办法，模拟数组只有两项，只能取0或1
			msg := prices[i%len(prices)] // len(prices)是2，所以i%2的值是0或1
			msg.Ts = time.Now().Unix()
			i++
			h.broadcast(msg)
		}
	}()
}

// clients当前在线的websocket连接名单
func (h *TickerHub) broadcast(msg tickerMsg) {
	// `Marshal(v any) ([]byte, error)`：struct → JSON 字节
	// 将msg转换为json
	data, _ := json.Marshal(msg)
	h.mu.Lock()
	defer h.mu.Unlock()
	// 遍历clients，将msg发送给每个客户端
	for conn := range h.clients {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("ws write: %v", err)
			// 如果广播失败了，就关闭连接，并从clients中删除
			_ = conn.Close()
			delete(h.clients, conn)
		}
	}
}
