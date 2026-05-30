package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Email        string         `gorm:"uniqueIndex;size:128;not null" json:"email"`
	PasswordHash string         `gorm:"size:255;not null" json:"-"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// DexSwap 链下 Swap 记录（Indexer / 前台上报入库）
type DexSwap struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	UserID          uint      `gorm:"index;not null" json:"user_id"`
	IdempotencyKey  string    `gorm:"uniqueIndex;size:64;not null" json:"idempotency_key"`
	TxHash          string    `gorm:"index;size:128;not null" json:"tx_hash"`
	Chain           string    `gorm:"size:32;not null;default:solana" json:"chain"`
	PoolID          string    `gorm:"size:64;not null" json:"pool_id"`
	AmountIn        string    `gorm:"type:decimal(36,18);not null" json:"amount_in"`
	AmountOut       string    `gorm:"type:decimal(36,18);not null" json:"amount_out"`
	Status          string    `gorm:"size:32;not null;default:pending" json:"status"` // pending | confirmed | failed
	Slot            *uint64   `json:"slot,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

// DexPool 流动性池快照（可由定时任务或手动录入）
type DexPool struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	PoolID    string    `gorm:"uniqueIndex;size:64;not null" json:"pool_id"`
	Chain     string    `gorm:"size:32;not null;default:solana" json:"chain"`
	BaseMint  string    `gorm:"size:64;not null" json:"base_mint"`
	QuoteMint string    `gorm:"size:64;not null" json:"quote_mint"`
	Price     string    `gorm:"type:decimal(36,18);not null" json:"price"`
	TVL       string    `gorm:"type:decimal(36,18)" json:"tvl"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Asset CEX 向：用户资产（练习用）
type Asset struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	UserID    uint   `gorm:"uniqueIndex:idx_user_symbol;not null" json:"user_id"`
	Symbol    string `gorm:"uniqueIndex:idx_user_symbol;size:16;not null" json:"symbol"`
	Available string `gorm:"type:decimal(36,18);not null;default:0" json:"available"`
	Frozen    string `gorm:"type:decimal(36,18);not null;default:0" json:"frozen"`
}

// CexOrder CEX 向：订单 mock
type CexOrder struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	UserID         uint      `gorm:"index;not null" json:"user_id"`
	IdempotencyKey string    `gorm:"uniqueIndex;size:64;not null" json:"idempotency_key"`
	Symbol         string    `gorm:"size:16;not null" json:"symbol"`
	Side           string    `gorm:"size:8;not null" json:"side"` // buy | sell
	Price          string    `gorm:"type:decimal(36,18);not null" json:"price"`
	Quantity       string    `gorm:"type:decimal(36,18);not null" json:"quantity"`
	Status         string    `gorm:"size:32;not null;default:pending" json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}
