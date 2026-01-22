package serializer

// DataList 基础的分页响应包装
type DataList struct {
	Items interface{} `json:"items"`            // 具体的数组数据
	Total int64       `json:"total"`            // 总条数
	Page  int         `json:"page" form:"page"` // 当前页
	Size  int         `json:"size" form:"size"` // 每页数量
}

// BuildDataList 通用构建函数
func BuildDataList(items interface{}, total int64, page, size int) DataList {
	return DataList{
		Items: items,
		Total: total,
		Page:  page,
		Size:  size,
	}
}
