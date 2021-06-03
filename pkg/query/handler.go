package query

import (
	"counseling-system/pkg/common"
	"counseling-system/pkg/utils"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

// CounselorListHandler return the Counselor list
func CounselorListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// pagination
		pageNum, err := strconv.Atoi(r.URL.Query().Get("pageNum"))
		pageSize, err := strconv.Atoi(r.URL.Query().Get("pageSize"))
		var p = pagination{
			PageNum:  pageNum,
			PageSize: pageSize,
		}

		like := r.URL.Query().Get("like")

		var pp pagination
		var counselors []common.Counselor

		// filters data body
		res, _ := ioutil.ReadAll(r.Body)
		if string(res) == "" {
			pp, counselors = queryCounselors(p, nil, false, like)
		} else {
			var option *filterOption
			err = json.Unmarshal(res, &option)
			utils.CheckErr(err)
			pp, counselors = queryCounselors(p, option, false, like)
		}

		var resp common.Response
		resp.Code = 1
		resp.Message = "ok"
		resp.Data = counselorRespData{
			pagination: pp,
			List:       counselors,
		}

		resJSON, _ := json.Marshal(resp)
		fmt.Fprintf(w, string(resJSON))
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

// NewlyCounselorsHandler return the Counselor list ordered by join_time
func NewlyCounselorsHandler(w http.ResponseWriter, r *http.Request) {
	// pagination
	pageNum, _ := strconv.Atoi(r.URL.Query().Get("pageNum"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	var p = pagination{
		PageNum:  pageNum,
		PageSize: pageSize,
	}

	var pp pagination
	var counselors []common.Counselor

	pp, counselors = queryCounselors(p, nil, true, "")

	var resp common.Response
	resp.Code = 1
	resp.Message = "ok"
	resp.Data = counselorRespData{
		pagination: pp,
		List:       counselors,
	}

	resJSON, _ := json.Marshal(resp)
	fmt.Fprintf(w, string(resJSON))
}

// CounselorInfoHandler return the info about a single Counselor
func CounselorInfoHandler(w http.ResponseWriter, r *http.Request) {
	var idStr = r.URL.Query().Get("id")
	var resp common.Response
	if idStr == "" {
		resp.Code = 0
		resp.Message = "咨询师不存在"
	} else {
		id, _ := strconv.Atoi(idStr)
		csl := queryCounselor(id)
		if csl == nil {
			resp.Code = 0
			resp.Message = "咨询师不存在"
		} else {
			resp.Code = 1
			resp.Message = "ok"
			resp.Data = *csl
		}
	}

	resJSON, _ := json.Marshal(resp)
	fmt.Fprintf(w, string(resJSON))
}

// NotificationHandler return the messages and notifications of user
func NotificationHandler(w http.ResponseWriter, r *http.Request) {
	var uid int
	if uid, _ = common.IsUserLogin(r); uid == -1 {
		http.Error(w, "Not Loggin", http.StatusUnauthorized)
		return
	}

	var resp common.Response
	resp.Code = 1
	resp.Message = "ok"
	resp.Data = queryNotifications(uid)

	resJSON, _ := json.Marshal(resp)
	fmt.Fprintf(w, string(resJSON))
}

// CounselingRecordHandler return all the counseling record
func CounselingRecordHandler(w http.ResponseWriter, r *http.Request) {
	var uid int
	var cid int
	var userType int
	var resp common.Response

	rid, _ := strconv.Atoi(r.URL.Query().Get("id"))
	if rid == -1 {
		records := queryCounselingRecords(userType, rid)
		resp.Code = 1
		resp.Message = "ok"
		resp.Data = records
		resJSON, _ := json.Marshal(resp)
		fmt.Fprintf(w, string(resJSON))
		return
	}

	if uid, userType = common.IsUserLogin(r); uid == -1 {
		http.Error(w, "Not Loggin", http.StatusUnauthorized)
		return
	}

	// query by id
	if rid != 0 {
		var uuid int
		var record *(common.RecordForm)
		if userType == 1 {
			uuid = common.GetCounselorIDByUID(uid)
		} else {
			uuid = uid
		}
		if record = queryCounselingRecordByID(userType, uuid, rid); record != nil {
			resp.Code = 1
			resp.Message = "ok"
			resp.Data = *(record)
		} else {
			resp.Code = 0
			resp.Message = "非法查询ID"
		}
	} else {
		var records []common.RecordForm
		if userType == 1 {
			cid = common.GetCounselorIDByUID(uid)
			records = queryCounselingRecords(userType, cid)
		} else {
			records = queryCounselingRecords(userType, uid)
		}
		resp.Code = 1
		resp.Message = "ok"
		resp.Data = records
	}

	resJSON, _ := json.Marshal(resp)
	fmt.Fprintf(w, string(resJSON))
}

// MessageHandler return all the message
func MessageHandler(w http.ResponseWriter, r *http.Request) {
	var uid int
	if uid, _ = common.IsUserLogin(r); uid == -1 {
		http.Error(w, "Not Loggin", http.StatusUnauthorized)
		return
	}

	var resp common.Response
	resp.Code = 1
	resp.Message = "ok"
	resp.Data = queryMessage(uid)
	resJSON, _ := json.Marshal(resp)

	fmt.Fprintf(w, string(resJSON))
}

// ArticleHandler 返回文章信息
// 带c_id查询需要校验权限
// 带id查询单个文章，否则查询列表(支持翻页)
func ArticleHandler(w http.ResponseWriter, r *http.Request) {
	aID, _ := strconv.Atoi(r.URL.Query().Get("id"))
	var resp common.Response

	// 查询List
	if aID <= 0 {
		// pagination
		pageNum, _ := strconv.Atoi(r.URL.Query().Get("pageNum"))
		pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
		var p = pagination{
			PageNum:  pageNum,
			PageSize: pageSize,
		}

		// 按c_id查询，需要校验权限
		cID, _ := strconv.Atoi(r.URL.Query().Get("cid"))
		category := r.URL.Query().Get("category")
		var queryArgs articleQueryArgs
		if category != "" {
			queryArgs.category = &category
		}
		if cID > 0 {
			if uid, _ := common.IsUserLogin(r); cID != common.GetCounselorIDByUID(uid) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			queryArgs.cID = &cID
		}

		// query
		resp.Code = 1
		resp.Message = "ok"
		resp.Data = queryArticleList(queryArgs, p)

		// result
		resJSON, _ := json.Marshal(resp)
		fmt.Fprintln(w, string(resJSON))
		return
	}

	// 查询单个
	var a *(common.Article)
	uid, _ := common.IsUserLogin(r) // 用于查询文章已读状态
	if a = common.QueryArticle(aID, uid); a != nil {
		resp.Code = 1
		resp.Message = "ok"
		resp.Data = *a
	} else {
		resp.Code = 0
		resp.Message = "未知文章ID"
	}

	w.Header().Set("x-text-x", "test")

	// result
	resJSON, _ := json.Marshal(resp)
	fmt.Fprintln(w, string(resJSON))
	return
}

// AskHandler 查询问答帖子
func AskHandler(w http.ResponseWriter, r *http.Request) {
	isAnswerStr := r.URL.Query().Get("isAnswer")
	featuredStr := r.URL.Query().Get("featured")
	var isAnswer, featured bool
	if isAnswerStr == "true" {
		isAnswer = true
	} else {
		isAnswer = false
	}
	if featuredStr == "true" {
		featured = true
	} else {
		featured = false
	}

	var resp common.Response
	resp.Code = 1
	resp.Message = "ok"
	resp.Data = queryAskList(featured, isAnswer)
	resJSON, _ := json.Marshal(resp)

	fmt.Fprintln(w, string(resJSON))
}

// AskItemHandler 查询问答项
func AskItemHandler(w http.ResponseWriter, r *http.Request) {
	askID, _ := strconv.Atoi(r.URL.Query().Get("id"))
	var resp common.Response
	if askID > 0 {
		uid, _ := common.IsUserLogin(r)
		if askItem := QueryAskItem(askID, uid); askItem != nil {
			resp.Code = 1
			resp.Message = "ok"
			resp.Data = &askItem
		} else {
			resp.Code = 0
			resp.Message = "未找到数据"
		}
	} else {
		resp.Code = 0
		resp.Message = "非法查询ID"
	}

	resJSON, _ := json.Marshal(resp)
	fmt.Fprintln(w, string(resJSON))
}

// FuzzyQueryHandler 模糊全局搜索
func FuzzyQueryHandler(w http.ResponseWriter, r *http.Request) {
	var uid int
	if uid, _ = common.IsUserLogin(r); uid == -1 {
		http.Error(w, "Not Loggin", http.StatusUnauthorized)
		return
	}

	var ttype = r.URL.Query().Get("type")
	var keyword = r.URL.Query().Get("keyword")

	var resp common.Response
	resp.Code = 1
	resp.Message = "ok"
	resp.Data = fuzzyQuery(keyword, ttype, uid)
	resJSON, _ := json.Marshal(resp)

	fmt.Fprintln(w, string(resJSON))
}

// PopularListHandler 阅读最多文章
func PopularListHandler(w http.ResponseWriter, r *http.Request) {
	var resp common.Response
	resp.Code = 1
	resp.Message = "ok"
	resp.Data = queryPopularList()

	resJSON, _ := json.Marshal(resp)
	fmt.Fprintln(w, string(resJSON))
}

// UserListHandler return the User list
func UserListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// pagination
		pageNum, err := strconv.Atoi(r.URL.Query().Get("pageNum"))
		pageSize, err := strconv.Atoi(r.URL.Query().Get("pageSize"))
		if err != nil {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		}
		var p = pagination{
			PageNum:  pageNum,
			PageSize: pageSize,
		}

		var pp pagination
		var users []common.User
		pp, users = queryUsers(p)

		var resp common.Response
		resp.Code = 1
		resp.Message = "ok"
		resp.Data = userRespData{
			pagination: pp,
			List:       users,
		}

		resJSON, _ := json.Marshal(resp)
		fmt.Fprintf(w, string(resJSON))
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

// ArticleListHandler return the Article list
func ArticleListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// pagination
		pageNum, err := strconv.Atoi(r.URL.Query().Get("pageNum"))
		pageSize, err := strconv.Atoi(r.URL.Query().Get("pageSize"))
		if err != nil {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		}
		var p = pagination{
			PageNum:  pageNum,
			PageSize: pageSize,
		}

		var pp pagination
		var articles []common.Article
		pp, articles = queryArticle(p)

		var resp common.Response
		resp.Code = 1
		resp.Message = "ok"
		resp.Data = articleList{
			pagination: pp,
			List:       articles,
		}

		resJSON, _ := json.Marshal(resp)
		fmt.Fprintf(w, string(resJSON))
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}
