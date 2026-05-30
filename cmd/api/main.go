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
	userRepo := repository.NewUserRepo(db)
	dexRepo := repository.NewDexRepo(db)

	authSvc := service.NewAuthService(userRepo, cfg.JWTSecret)
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

	authH := handler.NewAuthHandler(authSvc)
	dexH := handler.NewDexHandler(dexSvc)
	tickerHub := ws.NewTickerHub()
	tickerHub.StartDemoBroadcast()

	if cfg.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}
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
