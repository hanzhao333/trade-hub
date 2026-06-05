package main

import (
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zhanghanzhao/trade-hub/internal/config"
	"github.com/zhanghanzhao/trade-hub/internal/handler"
	"github.com/zhanghanzhao/trade-hub/internal/middleware"
	redisclient "github.com/zhanghanzhao/trade-hub/internal/pkg/redis"
	"github.com/zhanghanzhao/trade-hub/internal/repository"
	"github.com/zhanghanzhao/trade-hub/internal/service"
	"github.com/zhanghanzhao/trade-hub/internal/ws"
)

func main() {
	cfg := config.Load()

	db := repository.NewDB(cfg.DBDSN)
	// 有点不明白，上面已经拿到db了，这里为什么还要用userRepo再包一层，
	// 因为我看NewUserRepo其实也没做什么啊
	// 我可以理解成，你要去访问数据库的话，其实就是要有这一行代码的，就是关于授权这一块的话
	// 只要用户发起请求，都需要执行repository.NewUserRepo(db)，所以为了防止写重复代码
	// 干脆将这一行代码写在最外层，可以这么理解吗？
	userRepo := repository.NewUserRepo(db) // 这个userRepo是结构体UserRepo{db: db}
	dexRepo := repository.NewDexRepo(db)
	// 这个authSvc是结构体AuthService{users: userRepo, jwtSecret: []byte(cfg.JWTSecret)}
	authSvc := service.NewAuthService(userRepo, cfg.JWTSecret)
	// 这个dexSvc是结构体DexService{dex: dexRepo}
	dexSvc := service.NewDexService(dexRepo)

	if err := dexSvc.SeedDemoPools(); err != nil {
		log.Printf("seed pools: %v", err)
	}

	rdb := redisclient.NewClient(cfg.RedisAddr)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := redisclient.Ping(ctx, rdb); err != nil {
		log.Printf("redis ping failed (optional for week1): %v", err)
	} else {
		log.Println("redis connected")
	}

	authH := handler.NewAuthHandler(authSvc) // 这个authH是结构体AuthHandler{auth: authSvc}
	dexH := handler.NewDexHandler(dexSvc)
	tickerHub := ws.NewTickerHub()
	// 启动demo广播，每3秒推送一次demo价格
	// 这个时候clients可能是空的，因为前端必须通过ws/ticket接口，
	// 来触发ticket.go里面的register函数，才会将自己添加到clients中
	tickerHub.StartDemoBroadcast()

	// 如果环境是生产环境，则设置为生产模式,否则设置为开发模式
	// 生产模式下，gin会自动优化性能，开发模式下，gin会自动打印更多调试信息
	if cfg.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}
	// 创建一个gin的HTTP引擎，后面所有的路由都挂在这个r上面
	r := gin.Default()

	r.GET("/health", handler.Health)
	r.GET("/ws/ticker", tickerHub.HandleWS)

	v1 := r.Group("/api/v1")
	{
		v1.POST("/auth/register", authH.Register)
		v1.POST("/auth/login", authH.Login)
		v1.GET("/dex/pools", dexH.ListPools) // 池子列表公开（行情展示）

		authed := v1.Group("")
		authed.Use(middleware.Auth(authSvc))
		{
			authed.GET("/profile", authH.Profile)
			authed.POST("/dex/swaps", dexH.CreateSwap)
			authed.GET("/dex/swaps", dexH.ListSwaps)
		}
	}

	log.Printf("trade-hub listening on %s", cfg.HTTPAddr)
	if err := r.Run(cfg.HTTPAddr); err != nil {
		log.Fatal(err)
	}
}
