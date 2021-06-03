package query

import (
	"counseling-system/pkg/common"
	"counseling-system/pkg/utils"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// sort askList
func (al askList) Len() int {
	return len(al)
}
func (al askList) Less(i, j int) bool {
	return al[i].AnswerCount+al[i].StarCount > al[j].AnswerCount+al[j].StarCount
}
func (al askList) Swap(i, j int) {
	al[i], al[j] = al[j], al[i]
}

// sort counselor list
func (cl counselorList) Len() int {
	return len(cl)
}
func (cl counselorList) Less(i, j int) bool {
	var iRate = *(cl[i].GoodRate)
	var jRate = *(cl[j].GoodRate)
	var iServe = float64(cl[i].ServiceCount)
	var jServe = float64(cl[j].ServiceCount)

	return iRate*iServe > jRate*jServe
}
func (cl counselorList) Swap(i, j int) {
	cl[i], cl[j] = cl[j], cl[i]
}

// 咨询师列表
func queryCounselors(p pagination, option *filterOption, isNew bool, like string) (pagination, counselorList) {
	var queryCountStr = "select count(*) from counselor"
	var queryStr = "select id, u_id, name, gender, description, work_years, good_rate, motto, audio_price, video_price, ftf_price, city, topic, topic_other, create_time from counselor"

	var firstRecordIndex = (p.PageNum - 1) * p.PageSize

	// append with query params
	if option != nil {
		var base string
		if option.Method != nil || option.City != nil || option.Topic != nil {
			queryCountStr += " where "
			queryStr += " where "
		}
		// method
		if option.Method != nil {
			methodID := *(option.Method)
			methodCode := (*(common.GetDictInfoByID(methodID))).ID
			if methodCode == 1 {
				base = "ftf_price>0 and "
			} else if methodCode == 2 {
				base = "video_price>0 and "
			} else if methodCode == 3 {
				base = "audio_price>0 and "
			} else {
				base = ""
			}
			queryCountStr += base
			queryStr += base
		}
		// city
		if option.City != nil {
			city := *(option.City)
			if city > 0 {
				base = fmt.Sprintf("city='%v' and ", city)
			} else {
				base = ""
			}
			queryCountStr += base
			queryStr += base
		}
		// topic
		if option.Topic != nil {
			topic := *(option.Topic)
			if topic > 0 {
				base = fmt.Sprintf("topic='%v' and ", topic)
			} else {
				base = ""
			}
			queryCountStr += base
			queryStr += base
		}
	}

	// fuzzy query
	if like != "" {
		var likeStr string
		if strings.Contains(queryStr, "where") {
			likeStr = "name LIKE '%" + like + "%'"
			queryCountStr += likeStr
			queryStr += likeStr
		} else {
			likeStr = " where name LIKE '%" + like + "%'"
			queryCountStr += likeStr
			queryStr += likeStr
		}
	}
	queryCountStr = strings.TrimSuffix(queryCountStr, " and ")
	queryStr = strings.TrimSuffix(queryStr, " and ")

	// orderby
	if isNew {
		queryStr += " ORDER BY create_time DESC"
	}

	// db op
	count := utils.QueryDBRow(queryCountStr)
	var pp = pagination{PageNum: p.PageNum, PageSize: p.PageSize, Total: count}

	// empty
	if count == 0 || count <= firstRecordIndex {
		var emptyCounslor = []common.Counselor{}
		return pp, emptyCounslor
	}

	var cl counselorList
	var concatList counselorList
	rows := utils.QueryDB(queryStr)
	for rows.Next() {
		var c common.Counselor
		var cityID *int
		var topicID int
		rows.Scan(&c.ID, &c.UID, &c.Name, &c.Gender, &c.Description, &c.WorkYears, &c.GoodRate, &c.Motto, &c.AudioPrice, &c.VideoPrice, &c.FtfPrice, &cityID, &topicID, &c.TopicOther, &c.ApplyTime)

		// city和topic单独处理
		if cityID == nil {
			c.City = nil
		} else {
			c.City = common.GetDictInfoByID(*cityID)
		}
		c.Topic = *(common.GetDictInfoByID(topicID))

		// rate
		rate := queryServiceRate(c.ID)
		c.GoodRate = &rate

		// serviceCount
		c.ServiceCount = queryServiceCount(c.ID)

		cl = append(cl, c)
	}
	rows.Close()

	if !isNew {
		sort.Sort(cl)
	}

	var start = firstRecordIndex
	var end = firstRecordIndex + p.PageSize
	if end > count {
		end = count
	}
	concatList = cl[start:end]
	return pp, concatList
}

func queryCounselor(id int) *(common.Counselor) {
	var queryStr = fmt.Sprintf("select id, u_id, name, gender, description, work_years, good_rate, motto, audio_price, video_price, ftf_price, city, topic, topic_other, create_time from counselor where id='%v'", id)

	rows := utils.QueryDB(queryStr)
	if rows.Next() {
		var c common.Counselor
		var cityID *int
		var topicID int
		rows.Scan(&c.ID, &c.UID, &c.Name, &c.Gender, &c.Description, &c.WorkYears, &c.GoodRate, &c.Motto, &c.AudioPrice, &c.VideoPrice, &c.FtfPrice, &cityID, &topicID, &c.TopicOther, &c.ApplyTime)

		// city和topic单独处理
		if cityID == nil {
			c.City = nil
		} else {
			c.City = common.GetDictInfoByID(*cityID)
		}
		c.Topic = *(common.GetDictInfoByID(topicID))

		// thanks letter
		letters, lCount := queryServiceLetter(c.ID)
		c.Letters = letters
		c.LettersCount = lCount
		// service count
		c.ServiceCount = queryServiceCount(c.ID)
		// rate
		rate := queryServiceRate(c.ID)
		c.GoodRate = &rate

		// post article
		c.ArticleList = common.GetArticleListByCID(c.ID)

		// details
		c.Details = queryCounselorDetail(c.ID)

		rows.Close()
		return &c
	}

	return nil
}

// 查看未读通知消息
func queryNotifications(uID int) []common.Notification {
	var notis []common.Notification
	var queryStr = fmt.Sprintf("select id, type, title, `desc`, create_time, payload from notification where u_id=%v and is_read=0 ORDER BY create_time DESC", uID)
	rows := utils.QueryDB(queryStr)
	for rows.Next() {
		var noti common.Notification
		rows.Scan(&noti.ID, &noti.Type, &noti.Title, &noti.Desc, &noti.Time, &noti.Payload)
		notis = append(notis, noti)
	}
	rows.Close()
	return notis
}

// 查询咨询记录
func queryCounselingRecords(userType int, id int) []common.RecordForm {
	var records []common.RecordForm
	var queryStr = "select id, c_id, method, times, name, age, gender, phone, contact_phone, contact_name, contact_rel, `desc`, status, create_time, start_time, location, cancel_reason1, cancel_reason2, rating_score, rating_text, letter, price from counseling_record where"
	if userType == 1 {
		queryStr = fmt.Sprintf("%v c_id=%v", queryStr, id)
	} else {
		queryStr = fmt.Sprintf("%v u_id=%v", queryStr, id)
	}

	queryStr += " ORDER BY create_time DESC"

	rows := utils.QueryDB(queryStr)
	for rows.Next() {
		var r common.RecordForm
		rows.Scan(&r.ID, &r.CID, &r.Method, &r.Times, &r.Name, &r.Age, &r.Gender, &r.Phone, &r.ContactPhone, &r.ContactName, &r.ContactRel, &r.Desc, &r.Status, &r.CreateTime, &r.StartTime, &r.Location, &r.CancelReason1, &r.CancelReason2, &r.RatingScore, &r.RatingText, &r.Letter, &r.Price)
		r.CounselorName = common.GetCounselorNameByCID(r.CID)
		r.Method = strings.Replace(r.Method, "\\", "", -1)
		records = append(records, r)
	}
	rows.Close()
	return records
}

// 查询未读留言记录
func queryMessage(uid int) []common.Message {
	var messages []common.Message
	var queryStr = fmt.Sprintf("select id, sender, detail, create_time from message where receiver=%v and is_read=0 ORDER BY create_time DESC", uid)

	rows := utils.QueryDB(queryStr)
	for rows.Next() {
		var m common.Message
		var senderID int
		rows.Scan(&m.ID, &senderID, &m.Detail, &m.Time)
		m.SenderName = common.GetNameByUID(senderID)
		m.SenderID = senderID
		messages = append(messages, m)
	}
	rows.Close()

	return messages
}

// 查询咨询记录，按ID查询
func queryCounselingRecordByID(userType int, id int, rID int) *(common.RecordForm) {
	var r common.RecordForm
	var queryStr = "select id, c_id, u_id, method, times, name, age, gender, phone, contact_phone, contact_name, contact_rel, `desc`, status, location, cancel_reason1, cancel_reason2, rating_score, rating_text, letter, create_time, start_time from counseling_record where"
	if userType == 1 {
		queryStr = fmt.Sprintf("%v c_id=%v and id=%v", queryStr, id, rID)
	} else {
		queryStr = fmt.Sprintf("%v u_id=%v and id=%v", queryStr, id, rID)
	}

	rows := utils.QueryDB(queryStr)
	if rows.Next() {
		rows.Scan(&r.ID, &r.CID, &r.UID, &r.Method, &r.Times, &r.Name, &r.Age, &r.Gender, &r.Phone, &r.ContactPhone, &r.ContactName, &r.ContactRel, &r.Desc, &r.Status, &r.Location, &r.CancelReason1, &r.CancelReason2, &r.RatingScore, &r.RatingText, &r.Letter, &r.CreateTime, &r.StartTime)
		r.CounselorName = common.GetCounselorNameByCID(r.CID)
		r.CID = common.GetUserIDByCID(r.CID)
		r.Method = strings.Replace(r.Method, "\\", "", -1)

		rows.Close()
		return &r
	}
	return nil
}

// 查询文章列表，按c_id或category查询，支持分页
func queryArticleList(args articleQueryArgs, p pagination) articleList {
	var queryStr = "select id, cover, title, excerpt, content, category, tags, c_id, update_time from article where is_draft=0"
	var queryCountStr = "select count(*) from article where is_draft=0"
	var al articleList
	var total = 0
	var list []common.Article

	// query args
	if args.cID != nil {
		queryStr += fmt.Sprintf(" and c_id=%v", *(args.cID))
		queryCountStr += fmt.Sprintf(" and c_id=%v", *(args.cID))
	}
	if args.category != nil {
		queryStr += fmt.Sprintf(" and category='%v'", *(args.category))
		queryCountStr += fmt.Sprintf(" and category='%v'", *(args.category))
	}

	// orderby
	queryStr += " ORDER BY create_time DESC"

	// pagination
	var firstRecordIndex = (p.PageNum - 1) * p.PageSize
	queryStr += fmt.Sprintf(" LIMIT %v,%v", firstRecordIndex, p.PageSize)

	// total count
	total = utils.QueryDBRow(queryCountStr)

	// query
	rows := utils.QueryDB(queryStr)
	for rows.Next() {
		var a common.Article
		rows.Scan(&a.ID, &a.Cover, &a.Title, &a.Excerpt, &a.Content, &a.Category, &a.Tags, &a.CID, &a.PostTime)
		// handle authorName
		a.AuthorName = common.GetCounselorNameByCID(a.CID)
		// handle readCount
		a.ReadCount = common.GetCountByID(*(a.ID), "read", "article")
		// handle like count
		a.LikeCount = common.GetCountByID(*(a.ID), "like", "article")
		list = append(list, a)
	}
	rows.Close()

	// result
	al.PageNum = p.PageNum
	al.PageSize = p.PageSize
	al.Total = total
	al.List = list
	return al
}

// 查询热门阅读文章前四篇
func queryPopularList() []common.Article {
	var queryStr = "select COUNT(*) AS count, ref_id from read_count WHERE type='article' GROUP BY `ref_id` ORDER BY count DESC LIMIT 4"
	rows := utils.QueryDB(queryStr)
	var list []common.Article

	for rows.Next() {
		var id int
		var count int
		rows.Scan(&count, &id)
		if p := common.QueryArticle(id, -1); p != nil {
			list = append(list, *p)
		}
	}
	return list
}

// 查询文章标签
func queryAskTagsByAskID(id int) []common.AskTag {
	var tagIDStrs string
	var tags []common.AskTag
	var queryStr = fmt.Sprintf("select tags from ask where id=%v", id)
	rows := utils.QueryDB(queryStr)
	if rows.Next() {
		rows.Scan(&tagIDStrs)
		tagIDArr := strings.Split(tagIDStrs, ",")

		for _, v := range tagIDArr {
			var tag common.AskTag
			tagID, _ := strconv.Atoi(v)
			tag = *(common.GetTagByID(tagID))
			tags = append(tags, tag)
		}
	}

	rows.Close()
	return tags
}

// 查询问答列表（是否精选，是否按最近回答优先）
func queryAskList(featured bool, answer bool) askList {
	var queryStr string
	var al askList

	// 最近回答优先
	if answer == true {
		queryCmtStr := "select id, text, author, create_time, ask_id from ask_comment where reply_to=0 ORDER BY create_time DESC"
		cmtRows := utils.QueryDB(queryCmtStr)
		for cmtRows.Next() {
			var a common.AskItem
			var askID int
			var ac common.AskComment
			cmtRows.Scan(&ac.ID, &ac.Text, &ac.AuthorID, &ac.Time, &askID)
			// 回答者姓名
			ac.AuthorName = common.GetUserNameByID(ac.AuthorID)
			a.RecentComment = &ac
			// a.ID = askID
			askID = ac.ID

			// tags
			a.Tags = queryAskTagsByAskID(askID)

			// ask item
			queryStr = fmt.Sprintf("select title, content, is_anonymous, create_time from ask where id=%v", askID)
			rows := utils.QueryDB(queryStr)
			if rows.Next() {
				rows.Scan(&a.Title, &a.Content, &a.IsAnony, &a.Time)
				// 回答数
				var queryAnswerCountStr = fmt.Sprintf("select count(*) from ask_comment where ask_id=%v and reply_to=0", askID)
				a.AnswerCount = utils.QueryDBRow(queryAnswerCountStr)
				a.StarCount = common.GetCountByID(askID, "star", "ask")
			}
			rows.Close()
			al = append(al, a)
		}
		cmtRows.Close()
	} else {
		queryStr = "select id, title, content, is_anonymous, create_time from ask ORDER BY create_time DESC"
		rows := utils.QueryDB(queryStr)
		for rows.Next() {
			var a common.AskItem
			rows.Scan(&a.ID, &a.Title, &a.Content, &a.IsAnony, &a.Time)
			// tags
			a.Tags = queryAskTagsByAskID(a.ID)
			// 回答数
			var askID = a.ID
			var queryAnswerCountStr = fmt.Sprintf("select count(*) from ask_comment where ask_id=%v and reply_to=0", askID)
			a.AnswerCount = utils.QueryDBRow(queryAnswerCountStr)
			a.StarCount = common.GetCountByID(askID, "star", "ask")
			al = append(al, a)
		}
		rows.Close()
	}

	// 是否精选
	if !answer && featured {
		sort.Sort(al)
	}

	return al
}

// QueryAskItem 查询问答项
func QueryAskItem(askID int, uID int) *(common.AskItem) {
	var queryStr string
	var a common.AskItem
	// 查询问答
	queryStr = fmt.Sprintf("select id, title, content, is_anonymous, create_time, user_id from ask where id=%v", askID)
	rows := utils.QueryDB(queryStr)
	if rows.Next() {
		var authorID int
		rows.Scan(&a.ID, &a.Title, &a.Content, &a.IsAnony, &a.Time, &authorID)
		// 作者
		if !a.IsAnony {
			a.AuthorID = authorID
			a.AuthorName = common.GetUserNameByID(authorID)
		}
		// 数目, isRead, isStar isLike
		if uID != -1 {
			isRead := common.CheckReadStarLike(uID, a.ID, "read", "ask")
			isLike := common.CheckReadStarLike(uID, a.ID, "like", "ask")
			isStar := common.CheckReadStarLike(uID, a.ID, "star", "ask")
			a.IsRead = &isRead
			a.IsLike = &isLike
			a.IsStar = &isStar
		} else {
			a.IsRead = nil
			a.IsLike = nil
			a.IsStar = nil
		}
		var queryAnswerCountStr = fmt.Sprintf("select count(*) from ask_comment where ask_id=%v and reply_to=0", a.ID)
		a.AnswerCount = utils.QueryDBRow(queryAnswerCountStr)
		a.StarCount = common.GetCountByID(askID, "star", "ask")
		a.LikeCount = common.GetCountByID(askID, "like", "ask")
		a.ReadCount = common.GetCountByID(askID, "read", "ask")

		// tags
		a.Tags = queryAskTagsByAskID(a.ID)
		// 回答
		var cmts []common.AskComment
		queryStr = fmt.Sprintf("select id, text, author, create_time from ask_comment where reply_to=0 and ask_id=%v", askID)
		cmtRows := utils.QueryDB(queryStr)
		for cmtRows.Next() {
			var cmt common.AskComment
			cmtRows.Scan(&cmt.ID, &cmt.Text, &cmt.AuthorID, &cmt.Time)
			cmt.AuthorName = common.GetUserNameByID(cmt.AuthorID)

			// 嵌套回答
			var subCmts []common.AskComment
			var querySubCmtsStr = fmt.Sprintf("select id, text, author, create_time, reply_to from ask_comment where ref=%v and ask_id=%v", cmt.ID, askID)
			subCmtsRows := utils.QueryDB(querySubCmtsStr)
			for subCmtsRows.Next() {
				var sc common.AskComment
				var replyID int
				subCmtsRows.Scan(&sc.ID, &sc.Text, &sc.AuthorID, &sc.Time, &replyID)
				sc.AuthorName = common.GetUserNameByID(sc.AuthorID)
				var replyName = common.GetUserNameByID(replyID)
				sc.ReplyTo = &replyName
				subCmts = append(subCmts, sc)
			}
			cmt.SubComments = subCmts
			subCmtsRows.Close()
			cmts = append(cmts, cmt)
		}
		a.Comment = cmts

		cmtRows.Close()
		rows.Close()
		return &a
	}
	rows.Close()
	return nil
}

func fuzzyQueryCounselor(keyword string) []common.Counselor {
	var queryStr = "select id from counselor where name like '%" + keyword + "%' || topic_other like '%" + keyword + "%'"
	var list []common.Counselor
	rows := utils.QueryDB(queryStr)
	for rows.Next() {
		var id int
		rows.Scan(&id)
		list = append(list, *(queryCounselor(id)))
	}
	return list
}

func fuzzyQueryAsk(keyword string, uID int) []common.AskItem {
	var queryStr = "select id from ask where title like '%" + keyword + "%'"
	var list []common.AskItem
	rows := utils.QueryDB(queryStr)
	for rows.Next() {
		var id int
		rows.Scan(&id)
		list = append(list, *(QueryAskItem(id, uID)))
	}
	return list
}

func fuzzyQueryArticle(keyword string, uID int) []common.Article {
	var queryStr = "select id from article where title like '%" + keyword + "%' || tags like '%" + keyword + "%'"
	var list []common.Article
	rows := utils.QueryDB(queryStr)
	for rows.Next() {
		var id int
		rows.Scan(&id)
		list = append(list, *(common.QueryArticle(id, uID)))
	}
	return list
}

// 关键字查询 (咨询师，文章，问答)
func fuzzyQuery(keyword string, ttype string, uID int) fuzzyList {
	if keyword == "" {
		return nil
	}

	var fuzzyList interface{}
	switch ttype {
	case "counselor":
		fuzzyList = fuzzyQueryCounselor(keyword)
		break
	case "ask":
		fuzzyList = fuzzyQueryAsk(keyword, uID)
		break
	case "article":
		fuzzyList = fuzzyQueryArticle(keyword, uID)
		break
	default:
		break
	}
	return fuzzyList
}

// 查询咨询师服务人数
func queryServiceCount(cID int) int {
	var count int
	var queryStr = fmt.Sprintf("select count(*) from counseling_record where c_id=%v", cID)
	count = utils.QueryDBRow(queryStr)

	return count
}

// 查询咨询师服务好评率
func queryServiceRate(cID int) float64 {
	var rating float64
	var count float64
	var sum float64
	var queryStr = fmt.Sprintf("select rating_score from counseling_record where c_id=%v and status='finish' and rating_score!=-1", cID)

	rows := utils.QueryDB(queryStr)
	for rows.Next() {
		count++
		var rate float64
		rows.Scan(&rate)
		sum += rate
	}
	rows.Close()

	if count == 0 {
		return 1.00
	}

	rating = sum / (5 * count)
	value, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", rating), 64)
	return value
}

// 查询咨询师感谢信
func queryServiceLetter(cID int) ([]common.ThankLetter, int) {
	var count int
	var queryStr = fmt.Sprintf("select id, letter, letter_time from counseling_record where c_id=%v and letter != ''", cID)
	var ls []common.ThankLetter

	rows := utils.QueryDB(queryStr)
	for rows.Next() {
		var l common.ThankLetter
		rows.Scan(&l.ID, &l.Text, &l.Time)
		ls = append(ls, l)
		count++
	}
	rows.Close()

	return ls, count
}

// 查询咨询师详细信息
func queryCounselorDetail(cID int) []common.CounselorDetail {
	var queryStr = fmt.Sprintf("select id, title, content from counselor_detail where c_id=%v", cID)
	var details []common.CounselorDetail

	rows := utils.QueryDB(queryStr)
	for rows.Next() {
		var d common.CounselorDetail
		rows.Scan(&d.ID, &d.Title, &d.Content)

		details = append(details, d)
	}
	rows.Close()

	return details
}
