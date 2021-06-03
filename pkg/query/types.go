package query

import (
	"counseling-system/pkg/common"
)

// pagination 表示分页属性
type pagination struct {
	PageNum  int `json:"pageNum"`
	PageSize int `json:"pageSize"`
	Total    int `json:"total"`
}

// filterOption 表示条件查询
type filterOption struct {
	Topic  *int `json:"topic"`
	Method *int `json:"method"`
	City   *int `json:"city"`
}

// counselor response data
type counselorRespData struct {
	pagination
	List []common.Counselor `json:"list"`
}

type articleList struct {
	pagination
	List []common.Article `json:"list"`
}

type articleQueryArgs struct {
	category *string
	cID      *int
}

type askList []common.AskItem
type counselorList []common.Counselor

type fuzzyList interface{}
