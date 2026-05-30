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
	IdempotencyKey string `json:"idempotency_key" binding:"required"`
	TxHash         string `json:"tx_hash" binding:"required"`
	Chain          string `json:"chain"`
	PoolID         string `json:"pool_id" binding:"required"`
	AmountIn       string `json:"amount_in" binding:"required"`
	AmountOut      string `json:"amount_out" binding:"required"`
	Status         string `json:"status"`
}

func (h *DexHandler) CreateSwap(c *gin.Context) {
	var req createSwapReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	userID := c.GetUint(middleware.UserIDKey)
	swap, err := h.dex.RecordSwap(userID, service.CreateSwapInput{
		IdempotencyKey: req.IdempotencyKey,
		TxHash:         req.TxHash,
		Chain:          req.Chain,
		PoolID:         req.PoolID,
		AmountIn:       req.AmountIn,
		AmountOut:      req.AmountOut,
		Status:         req.Status,
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
