package util

// PaginationReq 分页请求参数绑定
type PaginationReq struct {
	Page     int `form:"page"`      // 页码
	PageSize int `form:"page_size"` // 每页数量
}

// GetPage 获取页码 (默认 1)
func (p *PaginationReq) GetPage() int {
	if p.Page <= 0 {
		return 1
	}
	return p.Page
}

// GetSize 获取每页数量 (默认 10，最大 100)
func (p *PaginationReq) GetSize() int {
	if p.PageSize <= 0 {
		return 10
	}
	if p.PageSize > 100 {
		return 100
	}
	return p.PageSize
}

// GetOffset 计算数据库偏移量
func (p *PaginationReq) GetOffset() int {
	return (p.GetPage() - 1) * p.GetSize()
}
