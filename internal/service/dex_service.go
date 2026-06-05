package service

import (
	"errors"
	"strings"

	"github.com/zhanghanzhao/trade-hub/internal/model"
	"github.com/zhanghanzhao/trade-hub/internal/repository"
)

var ErrDuplicateSwap = errors.New("duplicate idempotency_key")

type CreateSwapInput struct {
	IdempotencyKey string
	TxHash         string
	Chain          string
	PoolID         string
	AmountIn       string
	AmountOut      string
	Status         string
}

type DexService struct {
	dex *repository.DexRepo
}

func NewDexService(dex *repository.DexRepo) *DexService {
	return &DexService{dex: dex}
}

func (s *DexService) RecordSwap(userID uint, in CreateSwapInput) (*model.DexSwap, error) {
	if in.Chain == "" {
		in.Chain = "solana"
	}
	if in.Status == "" {
		in.Status = "pending"
	}
	swap := &model.DexSwap{
		UserID:         userID,
		IdempotencyKey: in.IdempotencyKey,
		TxHash:         in.TxHash,
		Chain:          in.Chain,
		PoolID:         in.PoolID,
		AmountIn:       in.AmountIn,
		AmountOut:      in.AmountOut,
		Status:         in.Status,
	}
	// 入库，调用repo里的CreateSwap方法
	if err := s.dex.CreateSwap(swap); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return nil, ErrDuplicateSwap
		}
		return nil, err
	}
	return swap, nil
}

func (s *DexService) ListSwaps(userID uint, page, pageSize int) ([]model.DexSwap, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	// 又去调用repo层里的ListSwaps方法
	return s.dex.ListSwaps(userID, page, pageSize)
}

func (s *DexService) ListPools() ([]model.DexPool, error) {
	// 去调用repo层里的ListPools方法
	// 得到一个pool列表 []model.DexPool
	return s.dex.ListPools()
}

func (s *DexService) SeedDemoPools() error {
	demos := []model.DexPool{
		{PoolID: "SOL-USDC", Chain: "solana", BaseMint: "SOL", QuoteMint: "USDC", Price: "142.50", TVL: "1200000"},
		{PoolID: "ETH-USDC", Chain: "ethereum", BaseMint: "ETH", QuoteMint: "USDC", Price: "3200.00", TVL: "5000000"},
	}
	for i := range demos {
		if err := s.dex.UpsertPool(&demos[i]); err != nil {
			return err
		}
	}
	return nil
}
