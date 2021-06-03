package info

import (
	"counseling-system/pkg/common"
	"counseling-system/pkg/query"
	"counseling-system/pkg/utils"
	"fmt"
	"strconv"
)

func getLogginUserInfo(uid int) preInfo {
	var info preInfo
	var queryStr = fmt.Sprintf("select id, username, phone, email, create_time from user where id=%v", uid)

	rows := utils.QueryDB(queryStr)
	if rows.Next() {
		rows.Scan(&info.ID, &info.UserName, &info.Phone, &info.Email, &info.CreateTime)
	}
	rows.Close()

	if cid := common.GetCounselorIDByUID(uid); cid != -1 {
		info.CID = &cid
	}

	// about ask
	getUserAskCountInfo(uid, &info)
	return info
}

func getUserAskCountInfo(uID int, pre *preInfo) {
	var queryStr = fmt.Sprintf("select count(*) from ask where user_id=%v", uID)
	var postCount, cmtCount int

	postCount = utils.QueryDBRow(queryStr)
	queryStr = fmt.Sprintf("select count(*) from ask_comment where author=%v and reply_to=0", uID)
	cmtCount = utils.QueryDBRow(queryStr)

	pre.AskCmtCount = cmtCount
	pre.AskPostCount = postCount
}

func getLogginCounselorInfo(uid int) common.Counselor {
	var c common.Counselor
	var queryStr = fmt.Sprintf("select id, u_id, name, gender, description, work_years, good_rate, motto, audio_price, video_price, ftf_price, city, topic, topic_other, create_time from counselor where u_id=%v", uid)

	rows := utils.QueryDB(queryStr)
	if rows.Next() {
		var cityID *int
		var topicID int
		rows.Scan(&c.ID, &c.UID, &c.Name, &c.Gender, &c.Description, &c.WorkYears, &c.GoodRate, &c.Motto, &c.AudioPrice, &c.VideoPrice, &c.FtfPrice, &cityID, &topicID, &c.TopicOther, &c.ApplyTime)

		if cityID != nil {
			c.City = common.GetDictInfoByID(*(cityID))
		} else {
			c.City = nil
		}
		c.Topic = *(common.GetDictInfoByID(topicID))
	}
	rows.Close()
	return c
}

// 文章草稿
func getArticleDraft(cID int) *(common.Article) {
	var queryStr = fmt.Sprintf("select id, cover, title, content, category, tags from article where is_draft=1 and c_id=%v", cID)
	var a common.Article

	rows := utils.QueryDB(queryStr)
	if rows.Next() {
		rows.Scan(&a.ID, &a.Cover, &a.Title, &a.Content, &a.Category, &a.Tags)
		rows.Close()
		return &a
	}
	rows.Close()
	return nil
}

func getAskTags() []common.AskTag {
	var queryTagParentStr = "select parent_id, ANY_VALUE(parent_name) from ask_tag GROUP BY `parent_id`"
	var queryTagStr string
	var at []common.AskTag

	rows1 := utils.QueryDB(queryTagParentStr)
	for rows1.Next() {
		var att common.AskTag
		rows1.Scan(&att.ID, &att.Name)
		at = append(at, att)
	}
	rows1.Close()

	for index, p := range at {
		queryTagStr = fmt.Sprintf("select id, name from ask_tag where parent_id='%v'", p.ID)
		rows2 := utils.QueryDB(queryTagStr)
		var subAt []common.AskTag
		for rows2.Next() {
			var subAtt common.AskTag
			var id int
			rows2.Scan(&id, &subAtt.Name)
			subAtt.ID = strconv.Itoa(id)
			subAt = append(subAt, subAtt)
		}
		rows2.Close()
		at[index].SubTags = &subAt
	}
	return at
}

func getMyArticleList(uid int) myArticleList {
	var cID int
	var list myArticleList
	cID = common.GetCounselorIDByUID(uid)

	// counselor
	if cID != -1 {
		var cmtList []common.Article
		var postList []common.Article
		var queryStr = fmt.Sprintf("select id from article where is_draft=0 and c_id=%v", cID)

		// posted article
		prows := utils.QueryDB(queryStr)
		for prows.Next() {
			var aid int
			var a common.Article
			prows.Scan(&aid)
			if p := common.QueryArticle(aid, -1); p != nil {
				a = *p
			}
			postList = append(postList, a)
		}
		list.PostArticleList = &postList
		prows.Close()

		// cmted list
		queryStr = fmt.Sprintf("select a_id from article_comment where author=%v GROUP BY a_id", uid)
		crows := utils.QueryDB(queryStr)
		for crows.Next() {
			var aid int
			var a common.Article
			crows.Scan(&aid)
			if p := common.QueryArticle(aid, -1); p != nil {
				a = *p
			}
			cmtList = append(cmtList, a)
		}
		list.CmtArticleList = cmtList
		crows.Close()
		return list
	}

	// user
	var userList []common.Article
	// cmted list
	queryStr := fmt.Sprintf("select a_id from article_comment where author=%v GROUP BY a_id", uid)
	rows := utils.QueryDB(queryStr)
	for rows.Next() {
		var aid int
		var a common.Article
		rows.Scan(&aid)
		if p := common.QueryArticle(aid, -1); p != nil {
			a = *p
		}
		userList = append(userList, a)
	}
	list.CmtArticleList = userList
	rows.Close()
	return list
}

func getMyAskList(uid int) myAskList {
	var queryStr string
	var list myAskList

	// post
	queryStr = fmt.Sprintf("select id from ask where user_id=%v", uid)
	prows := utils.QueryDB(queryStr)
	var plist []common.AskItem
	for prows.Next() {
		var a common.AskItem
		var aid int
		prows.Scan(&aid)
		if p := query.QueryAskItem(aid, -1); p != nil {
			a = *p
			plist = append(plist, a)
		}
	}
	list.PostAskList = plist
	prows.Close()

	// cmt
	queryStr = fmt.Sprintf("select ask_id from ask_comment where author=%v GROUP BY ask_id", uid)
	crows := utils.QueryDB(queryStr)
	var clist []common.AskItem
	for crows.Next() {
		var a common.AskItem
		var aid int
		crows.Scan(&aid)
		if p := query.QueryAskItem(aid, -1); p != nil {
			a = *p
			clist = append(clist, a)
		}
	}
	list.CmtAskList = clist
	crows.Close()

	return list
}
