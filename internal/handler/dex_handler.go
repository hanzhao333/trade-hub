package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zhanghanzhao/trade-hub/internal/middleware"
	"github.com/zhanghanzhao/trade-hub/internal/service"
	"github.com/zhanghanzhao/trade-hub/pkg/response"
)

type DexHandler struct {
	dex *service.DexService
}

func NewDexHandler(dex *service.DexService) *DexHandler {
	return &DexHandler{dex: dex}
}

type createSwapReq struct {
	// 幂等键 防止同一笔 swap 被重复入库
	IdempotencyKey string `json:"idempotency_key" binding:"required"`
	TxHash         string `json:"tx_hash" binding:"required"`
	Chain          string `json:"chain"`
	PoolID         string `json:"pool_id" binding:"required"`
	// SOL 换 USDC
	// 换入数量 1.5 SOL
	AmountIn string `json:"amount_in" binding:"required"`
	// 换出数量 142.3 USDC
	AmountOut string `json:"amount_out" binding:"required"`
	Status    string `json:"status"`
}

func (h *DexHandler) CreateSwap(c *gin.Context) {
	var req createSwapReq
	// 将请求 Body 里的 JSON赋值给req
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	// 从上下文里获取当前用户ID
	userID := c.GetUint(middleware.UserIDKey)
	// 在handle层调用service层里的RecordSwap方法，
	// RecordSwap方法会调用repo层里的CreateSwap方法
	swap, err := h.dex.RecordSwap(userID, service.CreateSwapInput{
		IdempotencyKey: req.IdempotencyKey, // 幂等键 防止同一笔 swap 被重复入库
		TxHash:         req.TxHash,         // 链上交易哈希（如 Solana signature）
		Chain:          req.Chain,          // 所在链，如 solana / ethereum
		PoolID:         req.PoolID,         // 交易池 ID，如 SOL-USDC
		AmountIn:       req.AmountIn,       // 换入数量（string 避免浮点精度丢失）
		AmountOut:      req.AmountOut,      // 换出数量
		Status:         req.Status,         // pending | confirmed | failed
	})
	if err != nil {
		if errors.Is(err, service.ErrDuplicateSwap) {
			response.Fail(c, 409, 409, "duplicate idempotency_key")
			return
		}
		response.Internal(c, err.Error())
		return
	}
	response.OK(c, swap)
}

func (h *DexHandler) ListSwaps(c *gin.Context) {
	userID := c.GetUint(middleware.UserIDKey)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	// 在handle层调用service层里的ListSwaps方法，
	list, total, err := h.dex.ListSwaps(userID, page, pageSize)
	if err != nil {
		response.Internal(c, err.Error())
		return
	}
	response.OK(c, gin.H{"list": list, "total": total, "page": page, "page_size": pageSize})
}

func (h *DexHandler) ListPools(c *gin.Context) {
	list, err := h.dex.ListPools()
	if err != nil {
		response.Internal(c, err.Error())
		return
	}
	response.OK(c, list)
}
