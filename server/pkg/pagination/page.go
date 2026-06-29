package pagination

import (
	"gorm.io/gorm"
)

// Page 分页请求
type Page struct {
	Page     int `form:"page" json:"page" binding:"min=1"`           // 页码，从1开始
	PageSize int `form:"page_size" json:"page_size" binding:"min=1,max=100"` // 每页条数，上限100
}

// PageResult 分页结果
type PageResult struct {
	List     interface{} `json:"list"`      // 数据列表
	Total    int64       `json:"total"`     // 总条数
	Page     int         `json:"page"`      // 当前页码
	PageSize int         `json:"page_size"` // 每页条数
}

// Paginate 执行分页查询
func Paginate(db *gorm.DB, p *Page, dest interface{}) (*PageResult, error) {
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	offset := (p.Page - 1) * p.PageSize
	if err := db.Offset(offset).Limit(p.PageSize).Find(dest).Error; err != nil {
		return nil, err
	}

	return &PageResult{
		List:     dest,
		Total:    total,
		Page:     p.Page,
		PageSize: p.PageSize,
	}, nil
}
