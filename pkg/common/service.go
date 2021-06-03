package common

import (
	"counseling-system/pkg/utils"
	"fmt"
	"net/http"
	"strconv"
)

// GetAllDictInfoByTypeCode return the counseling cities, methods, topics
func GetAllDictInfoByTypeCode(dictType int) []DictInfo {
	var infos []DictInfo
	var queryStr = fmt.Sprintf("select id, info_name from dict_info where type_code='%v'", dictType)

	infoRows := utils.QueryDB(queryStr)
	for infoRows.Next() {
		var info DictInfo
		infoRows.Scan(&info.ID, &info.Name)
		infos = append(infos, info)
	}
	infoRows.Close()
	return infos
}

// GetDictInfoByID return the dictInfo by id
func GetDictInfoByID(id int) *DictInfo {
	var info DictInfo
	var queryStr = fmt.Sprintf("select info_code, info_name from dict_info where id='%v'", id)

	infoRow := utils.QueryDB(queryStr)
	if infoRow.Next() {
		infoRow.Scan(&info.ID, &info.Name)
	}
	infoRow.Close()
	return &info
}

// IsUserLogin to check the user auth and return the user id and type if user logged in
func IsUserLogin(r *http.Request) (int, int) {
	session, _ := utils.Store.Get(r, "user_session")
	if auth, ok := session.Values["auth"].(bool); !ok || !auth {
		return -1, -1
	}

	// queryUserType
	var userType = -1
	uid, _ := session.Values["uid"].(int)
	rows := utils.QueryDB(fmt.Sprintf("select type from user where id='%v'", uid))
	if rows.Next() {
		rows.Scan(&userType)
	}
	rows.Close()
	return uid, userType
}

// AddNotification to notification table
func AddNotification(uid int, no Notification) {
	var insertStr = fmt.Sprintf("insert notification set u_id=%v, type=?, title=?, `desc`=?, payload=?", uid)

	if _, success := utils.InsertDB(insertStr, no.Type, no.Title, no.Desc, no.Payload); !success {
		fmt.Println("新增通知失败！")
	}
}

// GetUserIDByCID return the user ID for counselor
func GetUserIDByCID(cID int) int {
	var queryStr = fmt.Sprintf("select u_id from counselor where id=%v", cID)
	var uID int

	rows := utils.QueryDB(queryStr)
	if rows.Next() {
		rows.Scan(&uID)
	} else {
		fmt.Println("查找咨询师u_id失败！")
	}
	rows.Close()
	return uID
}

// GetCounselorIDByUID return the c_id for counselor
func GetCounselorIDByUID(id int) int {
	var queryStr = fmt.Sprintf("select id from counselor where u_id=%v", id)
	var cID int

	rows := utils.QueryDB(queryStr)
	if rows.Next() {
		rows.Scan(&cID)
	} else {
		cID = -1
	}
	rows.Close()
	return cID
}

// GetCounselorNameByCID return counselor name by CID
func GetCounselorNameByCID(id int) string {
	var queryStr = fmt.Sprintf("select name from counselor where id=%v", id)
	var name string

	rows := utils.QueryDB(queryStr)
	if rows.Next() {
		rows.Scan(&name)
	} else {
		fmt.Println("查找咨询师姓名失败！")
	}
	rows.Close()
	return name
}

// GetUserNameByID return the username
func GetUserNameByID(id int) string {
	var queryStr = fmt.Sprintf("select username from user where id=%v", id)
	var username string

	rows := utils.QueryDB(queryStr)
	if rows.Next() {
		rows.Scan(&username)
	}
	rows.Close()
	return username
}

// GetNameByUID return the username or counselor name according to userID
func GetNameByUID(uid int) string {
	var name string
	var queryStr = fmt.Sprintf("select username, type from user where id=%v", uid)

	rows := utils.QueryDB(queryStr)
	if rows.Next() {
		var userType int
		rows.Scan(&name, &userType)
		rows.Close()

		if userType == 1 {
			queryStr = fmt.Sprintf("select name from counselor where u_id=%v", uid)
			crows := utils.QueryDB(queryStr)
			if crows.Next() {
				crows.Scan(&name)
				crows.Close()
			} else {
				fmt.Println("根据UID查找咨询师姓名失败！")
			}
		}
	} else {
		fmt.Println("根据UID查找用户名失败！")
	}
	return name
}

// HandleApplyCity 面对面咨询城市的插入更新处理
func HandleApplyCity(city string, uid int) {
	if city == "" {
		return
	}
	var cityID int
	var queryStr = fmt.Sprintf("select id from dict_info where `type_code`=8 and `info_name`='%v'", city)
	existRows := utils.QueryDB(queryStr)

	if existRows.Next() {
		existRows.Scan(&cityID)
		existRows.Close()
	} else {
		infoCode := utils.QueryDBRow("select count(*) from dict_info where `type_code`=8") + 1
		if cID, success := utils.InsertDB("insert dict_info set type_code=?, info_code=?, info_name=?", 8, infoCode, city); success {
			cityID = int(cID)
		} else {
			fmt.Println("新增咨询城市出错！")
			return
		}
	}

	updateStr := fmt.Sprintf("update counselor set city=? where u_id='%v'", uid)
	if success := utils.UpdateDB(updateStr, cityID); !success {
		fmt.Println("更新咨询师所在城市出错")
	}
}

// CheckReadStarLike 检查状态(已读，已收藏，已点赞，已关注), t1: read & like & star & follow; t2: article & article_comment & ask
func CheckReadStarLike(uID int, refID int, t1 string, t2 string) bool {
	var queryStr = "select * from "
	switch t1 {
	case "read":
		queryStr += fmt.Sprintf("read_count where u_id=%v and ref_id=%v and type='%v'", uID, refID, t2)
	case "like":
		queryStr += fmt.Sprintf("star_like where u_id=%v and ref_id=%v and type1='like' and type2='%v' and is_cancel=0", uID, refID, t2)
	case "star":
		queryStr += fmt.Sprintf("star_like where u_id=%v and ref_id=%v and type1='star' and type2='%v' and is_cancel=0", uID, refID, t2)
	default:
		queryStr = ""
	}

	if queryStr == "" {
		return false
	}
	rows := utils.QueryDB(queryStr)
	if rows.Next() {
		rows.Close()
		return true
	}
	rows.Close()
	return false
}

// GetCountByID 获得数量(阅读，点赞，收藏，关注), t1: read & like & star & follow; t2: article & article_comment & ask
func GetCountByID(id int, t1 string, t2 string) int {
	var queryCountStr = "select count(*) from "
	switch t1 {
	case "read":
		queryCountStr += fmt.Sprintf("read_count where ref_id=%v and type='%v'", id, t2)
	case "like":
		queryCountStr += fmt.Sprintf("star_like where ref_id=%v and type1='like' and type2='%v' and is_cancel=0", id, t2)
	case "star":
		queryCountStr += fmt.Sprintf("star_like where ref_id=%v and type1='star' and type2='%v' and is_cancel=0", id, t2)
	default:
		queryCountStr = ""
	}
	if queryCountStr == "" {
		return 0
	}

	var count = utils.QueryDBRow(queryCountStr)
	return count
}

// GetTagByID 获得Tag
func GetTagByID(id int) *AskTag {
	var at AskTag
	var queryStr = fmt.Sprintf("select name, parent_id, parent_name from ask_tag where id=%v", id)
	rows := utils.QueryDB(queryStr)
	if rows.Next() {
		var subTag AskTag
		subTag.ID = strconv.Itoa(id)
		rows.Scan(&subTag.Name, &at.ID, &at.Name)
		at.SubTags = &([]AskTag{subTag})
		rows.Close()
		return &at
	}
	rows.Close()
	return nil
}

// QueryArticleComment 查询文章留言
func QueryArticleComment(aid int, uID int) []ArticleComment {
	var queryStr = fmt.Sprintf("select id, text, create_time, author, ref from article_comment where a_id=%v", aid)
	var cmts []ArticleComment

	rows := utils.QueryDB(queryStr)
	for rows.Next() {
		var cmt ArticleComment
		var ref *int
		rows.Scan(&cmt.ID, &cmt.Text, &cmt.PostTime, &cmt.AuthorID, ref)
		// handle ref
		if ref == nil {
			cmt.Ref = nil
		} else {
			cmt.Ref = QueryArticleCommentRefByID(*(ref))
		}
		// handle authorName
		cmt.AuthorName = GetUserNameByID(cmt.AuthorID)
		// handle isLike
		if uID != -1 {
			isLike := CheckReadStarLike(uID, cmt.ID, "like", "article_comment")
			cmt.IsLike = &isLike
		} else {
			cmt.IsLike = nil
		}
		// handle like count
		cmt.LikeCount = GetCountByID(cmt.ID, "like", "article_comment")

		cmts = append(cmts, cmt)
	}
	rows.Close()
	return cmts
}

// QueryArticleCommentRefByID 查询文章留言，按ID查询
func QueryArticleCommentRefByID(id int) *(ArticleComment) {
	var queryStr = fmt.Sprintf("select id, text, create_time, author from article_comment where id=%v", id)
	var cmt ArticleComment

	rows := utils.QueryDB(queryStr)
	if rows.Next() {
		rows.Scan(&cmt.ID, &cmt.Text, &cmt.PostTime, cmt.AuthorID)
		rows.Close()
		return &cmt
	}
	rows.Close()
	return nil
}

// QueryArticle 查询文章，按id查询
func QueryArticle(id int, uID int) *Article {
	var queryStr = fmt.Sprintf("select id, cover, title, excerpt, content, category, tags, c_id, update_time from article where is_draft=0 and id=%v", id)
	var a Article

	rows := utils.QueryDB(queryStr)
	if rows.Next() {
		rows.Scan(&a.ID, &a.Cover, &a.Title, &a.Excerpt, &a.Content, &a.Category, &a.Tags, &a.CID, &a.PostTime)
		rows.Close()
		a.AuthorName = GetCounselorNameByCID(a.CID)
		// handle comment
		a.Comment = QueryArticleComment(*(a.ID), uID)
		// handle isRead, isStar isLike
		if uID != -1 {
			isRead := CheckReadStarLike(uID, *(a.ID), "read", "article")
			isLike := CheckReadStarLike(uID, *(a.ID), "like", "article")
			isStar := CheckReadStarLike(uID, *(a.ID), "star", "article")
			a.IsRead = &isRead
			a.IsLike = &isLike
			a.IsStar = &isStar
		} else {
			a.IsRead = nil
			a.IsLike = nil
			a.IsStar = nil
		}
		// handle readCount
		a.ReadCount = GetCountByID(*(a.ID), "read", "article")
		// handle like count
		a.LikeCount = GetCountByID(*(a.ID), "like", "article")
		return &a
	}
	rows.Close()
	return nil
}

// GetArticleListByCID 获得咨询师专栏文章
func GetArticleListByCID(cid int) []Article {
	var list []Article
	var queryStr = fmt.Sprintf("select id from article where is_draft=0 and c_id=%v", cid)

	rows := utils.QueryDB(queryStr)
	for rows.Next() {
		var aid int
		var a Article
		rows.Scan(&aid)
		if p := QueryArticle(aid, -1); p != nil {
			a = *p
		}
		list = append(list, a)
	}

	rows.Close()
	return list
}

// GetAuthorID 获得作者userID，文章或问答
func GetAuthorID(t string, id int) int {
	var queryStr string
	var authorID int
	if t == "article" {
		queryStr = fmt.Sprintf("select c_id from article where id=%v", id)
	} else if t == "ask" {
		queryStr = fmt.Sprintf("select user_id from ask where id=%v", id)
	} else {
		return -1
	}

	rows := utils.QueryDB(queryStr)
	if rows.Next() {
		rows.Scan(&authorID)
	} else {
		authorID = -1
	}

	rows.Close()
	return authorID
}
