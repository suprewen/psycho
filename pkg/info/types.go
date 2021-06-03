package info

import (
	"counseling-system/pkg/common"
)

type filter struct {
	Topic  []common.DictInfo `json:"topic"`
	Method []common.DictInfo `json:"method"`
	City   []common.DictInfo `json:"city"`
}

type preAsk struct {
	AskCmtCount  int `json:"askCmtCount"`
	AskPostCount int `json:"askPostCount"`
}

type preInfo struct {
	common.User
	preAsk
}

type myArticleList struct {
	PostArticleList *[]common.Article `json:"postArticleList"`
	CmtArticleList  []common.Article  `json:"cmtArticleList"`
}

type myAskList struct {
	PostAskList []common.AskItem `json:"postAskList"`
	CmtAskList  []common.AskItem `json:"cmtAskList"`
}
