package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/zhanghanzhao/trade-hub/pkg/response"
)

func Health(c *gin.Context) {
	response.OK(c, gin.H{"status": "ok", "service": "trade-hub"})
}
