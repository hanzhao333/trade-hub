package repository

import (
	"github.com/zhanghanzhao/trade-hub/internal/model"
	"gorm.io/gorm"
)

type DexRepo struct {
	db *gorm.DB
}

func NewDexRepo(db *gorm.DB) *DexRepo {
	return &DexRepo{db: db}
}

func (r *DexRepo) CreateSwap(s *model.DexSwap) error {
	return r.db.Create(s).Error
}

func (r *DexRepo) ListSwaps(userID uint, page, pageSize int) ([]model.DexSwap, int64, error) {
	var list []model.DexSwap
	var total int64
	// r.db.Model(&model.DexSwap{}) // 查哪张表
	// 查user_id = userID的记录
	q := r.db.Model(&model.DexSwap{}).Where("user_id = ?", userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	// 排序，按id降序，分页查询
	err := q.Order("id desc").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *DexRepo) ListPools() ([]model.DexPool, error) {
	var list []model.DexPool
	err := r.db.Order("pool_id asc").Find(&list).Error
	return list, err
}

func (r *DexRepo) UpsertPool(p *model.DexPool) error {
	return r.db.Where("pool_id = ?", p.PoolID).Assign(p).FirstOrCreate(p).Error
}
