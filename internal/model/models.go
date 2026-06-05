package model

import (
	"time"

	"gorm.io/gorm"
)

// User 系统账户，用于注册/登录；DexSwap 等业务的 UserID 指向本表 ID。
type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`                         // 用户唯一编号，JWT 与关联表外键
	Email        string         `gorm:"uniqueIndex;size:128;not null" json:"email"`   // 登录邮箱，全站唯一
	PasswordHash string         `gorm:"size:255;not null" json:"-"`                   // bcrypt 哈希，永不通过 API 返回
	CreatedAt    time.Time      `json:"created_at"`                                   // 注册时间
	UpdatedAt    time.Time      `json:"updated_at"`                                   // 最后更新时间
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`                               // 软删除时间，非空则普通查询不可见
}

// DexSwap 链下 Swap 流水：Indexer 扫链或前台上报入库；本服务不在链上签名。
type DexSwap struct {
	ID             uint      `gorm:"primaryKey" json:"id"`                              // 库内自增主键
	UserID         uint      `gorm:"index;not null" json:"user_id"`                     // 归属用户，对应 User.ID
	IdempotencyKey string    `gorm:"uniqueIndex;size:64;not null" json:"idempotency_key"` // 幂等键，防重复提交
	TxHash         string    `gorm:"index;size:128;not null" json:"tx_hash"`            // 链上交易哈希（如 Solana signature）
	Chain          string    `gorm:"size:32;not null;default:solana" json:"chain"`      // 所在链，如 solana / ethereum
	PoolID         string    `gorm:"size:64;not null" json:"pool_id"`                   // 交易池 ID，如 SOL-USDC
	AmountIn       string    `gorm:"type:decimal(36,18);not null" json:"amount_in"`       // 换入数量（string 避免浮点精度丢失）
	AmountOut      string    `gorm:"type:decimal(36,18);not null" json:"amount_out"`     // 换出数量
	Status         string    `gorm:"size:32;not null;default:pending" json:"status"`    // pending | confirmed | failed
	Slot           *uint64   `json:"slot,omitempty"`                                    // 可选，Solana 区块 slot；nil 表示未填
	CreatedAt      time.Time `json:"created_at"`                                        // 入库时间
}

// DexPool 流动性池快照，供池子列表与行情展示；启动时可 SeedDemoPools 灌 demo 数据。
type DexPool struct {
	ID        uint      `gorm:"primaryKey" json:"id"`                           // 库内自增主键
	PoolID    string    `gorm:"uniqueIndex;size:64;not null" json:"pool_id"`    // 业务池 ID，全站唯一，如 SOL-USDC
	Chain     string    `gorm:"size:32;not null;default:solana" json:"chain"`   // 池子所在链
	BaseMint  string    `gorm:"size:64;not null" json:"base_mint"`              // 基础代币（练习简写；链上一般为 mint 地址）
	QuoteMint string    `gorm:"size:64;not null" json:"quote_mint"`             // 计价代币
	Price     string    `gorm:"type:decimal(36,18);not null" json:"price"`      // 当前价格，如 1 SOL ≈ 142.50 USDC
	TVL       string    `gorm:"type:decimal(36,18)" json:"tvl"`                 // 池子总锁仓量（Total Value Locked）
	UpdatedAt time.Time `json:"updated_at"`                                     // 快照最后更新时间
}

// Asset CEX 向：用户资产余额（练习用，业务接口尚未实现）。
type Asset struct {
	ID        uint   `gorm:"primaryKey" json:"id"`                                    // 库内自增主键
	UserID    uint   `gorm:"uniqueIndex:idx_user_symbol;not null" json:"user_id"`     // 归属用户
	Symbol    string `gorm:"uniqueIndex:idx_user_symbol;size:16;not null" json:"symbol"` // 币种，与 UserID 联合唯一
	Available string `gorm:"type:decimal(36,18);not null;default:0" json:"available"` // 可用余额，可下单/提现
	Frozen    string `gorm:"type:decimal(36,18);not null;default:0" json:"frozen"`    // 冻结余额，挂单或处理中暂不可用
}

// CexOrder CEX 向：买卖订单 mock（练习用，业务接口尚未实现）。
type CexOrder struct {
	ID             uint      `gorm:"primaryKey" json:"id"`                              // 订单号
	UserID         uint      `gorm:"index;not null" json:"user_id"`                     // 下单用户
	IdempotencyKey string    `gorm:"uniqueIndex;size:64;not null" json:"idempotency_key"` // 幂等键，防重复下单
	Symbol         string    `gorm:"size:16;not null" json:"symbol"`                    // 交易对，如 BTC-USDT
	Side           string    `gorm:"size:8;not null" json:"side"`                       // buy | sell
	Price          string    `gorm:"type:decimal(36,18);not null" json:"price"`         // 限价单价格
	Quantity       string    `gorm:"type:decimal(36,18);not null" json:"quantity"`      // 委托数量
	Status         string    `gorm:"size:32;not null;default:pending" json:"status"`    // 订单状态，如 pending / filled / cancelled
	CreatedAt      time.Time `json:"created_at"`                                        // 下单时间
}
