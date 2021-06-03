package info

import (
	"counseling-system/pkg/common"
	"encoding/json"
	"fmt"
	"net/http"
)

// const
const (
	CITY   = 8
	METHOD = 10
	TOPIC  = 4
)

// CounselorFilterHandler get all filters about counselor
func CounselorFilterHandler(w http.ResponseWriter, r *http.Request) {
	var res common.Response
	var cities = common.GetAllDictInfoByTypeCode(CITY)
	var methods = common.GetAllDictInfoByTypeCode(METHOD)
	var topics = common.GetAllDictInfoByTypeCode(TOPIC)

	res.Code = 1
	res.Message = "ok"
	res.Data = filter{Topic: topics, City: cities, Method: methods}

	resJSON, _ := json.Marshal(res)
	fmt.Fprintln(w, string(resJSON))
}

// PreHandler return loggin user info
func PreHandler(w http.ResponseWriter, r *http.Request) {
	var uid int
	if uid, _ = common.IsUserLogin(r); uid == -1 {
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
		return
	}

	var res common.Response

	res.Data = getLogginUserInfo(uid)
	res.Code = 1
	res.Message = "ok"

	resJSON, _ := json.Marshal(res)
	fmt.Fprintln(w, string(resJSON))
}

// PreCounselorHandler return the loggin counselor info
func PreCounselorHandler(w http.ResponseWriter, r *http.Request) {
	var uid int
	var userType int
	if uid, userType = common.IsUserLogin(r); uid == -1 || userType != 1 {
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
		return
	}

	var res common.Response

	res.Data = getLogginCounselorInfo(uid)
	res.Code = 1
	res.Message = "ok"

	resJSON, _ := json.Marshal(res)
	fmt.Fprintln(w, string(resJSON))
}

// ArticleDraftHandler 保存为草稿的文章
func ArticleDraftHandler(w http.ResponseWriter, r *http.Request) {
	var uid int
	var userType int
	if uid, userType = common.IsUserLogin(r); uid == -1 || userType != 1 {
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
		return
	}

	var cID = common.GetCounselorIDByUID(uid)
	var article *(common.Article)
	var resp common.Response

	if article = getArticleDraft(cID); article != nil {
		resp.Code = 1
		resp.Message = "ok"
		resp.Data = *(article)
	} else {
		resp.Code = 0
		resp.Message = "无草稿"
	}

	resJSON, _ := json.Marshal(resp)
	fmt.Fprintln(w, string(resJSON))
}

// AskTagsHandler 获取问答标签
func AskTagsHandler(w http.ResponseWriter, r *http.Request) {
	var resp common.Response
	resp.Code = 1
	resp.Message = "ok"
	resp.Data = getAskTags()

	resJSON, _ := json.Marshal(resp)
	fmt.Fprintln(w, string(resJSON))
}

// MyArticleListHandler 获取我的问题
func MyArticleListHandler(w http.ResponseWriter, r *http.Request) {
	var uid int
	if uid, _ = common.IsUserLogin(r); uid == -1 {
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
		return
	}

	var resp common.Response
	resp.Code = 1
	resp.Message = "ok"
	resp.Data = getMyArticleList(uid)

	resJSON, _ := json.Marshal(resp)
	fmt.Fprintln(w, string(resJSON))
}

// MyAskListHandler 获取我的问题
func MyAskListHandler(w http.ResponseWriter, r *http.Request) {
	var uid int
	if uid, _ = common.IsUserLogin(r); uid == -1 {
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
		return
	}

	var resp common.Response
	resp.Code = 1
	resp.Message = "ok"
	resp.Data = getMyAskList(uid)

	resJSON, _ := json.Marshal(resp)
	fmt.Fprintln(w, string(resJSON))
}
