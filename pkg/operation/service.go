package operation

import (
	"counseling-system/pkg/common"
	"counseling-system/pkg/utils"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func addCounselingRecord(formData common.RecordForm, uid int) (string, bool) {
	var insertStr = "insert counseling_record set c_id=?, u_id=?, method=?, times=?, name=?, age=?, gender=?, phone=?, contact_phone=?, contact_name=?, contact_rel=?, `desc`=?, price=?, status=?"
	var resp common.Response

	// method含双引号，插入数据库前进行转义处理
	methodStr := strings.Replace(formData.Method, "\"", "\\\"", -1)
	// method name处理
	methodReg := regexp.MustCompile(`name":"(.*?)"`)
	params := methodReg.FindStringSubmatch(formData.Method)
	var methodName = params[1]

	if rID, success := utils.InsertDB(insertStr, formData.CID, uid, methodStr, formData.Times, formData.Name, formData.Age, formData.Gender, formData.Phone, formData.ContactPhone, formData.ContactName, formData.ContactRel, formData.Desc, formData.Price, "wait_contact"); success {
		// 增加通知
		var title = fmt.Sprintf("%v向您发起了咨询预约(%v)，请及时确认", formData.Name, methodName)
		var no = common.Notification{Title: title, Desc: "", Type: "counseling", Payload: int(rID)}
		common.AddNotification(common.GetUserIDByCID(formData.CID), no)

		resp.Code = 1
		resp.Message = "ok"
		resJSON, _ := json.Marshal(resp)
		return string(resJSON), true
	}
	fmt.Println("新增咨询记录失败，数据库插入错误")
	return "", false
}

// 将通知或私信标记为已读
func markRead(ids []string, t string) bool {
	var updateStr string
	if t == "notification" {
		updateStr = "update notification set is_read=1 where "
	} else if t == "message" {
		updateStr = "update message set is_read=1 where "
	} else {
		fmt.Println("标注通知已读，非法类型")
		return false
	}

	for _, id := range ids {
		iid, _ := strconv.Atoi(id)
		updateStr += fmt.Sprintf("id=%v || ", iid)
	}

	updateStr = strings.TrimSuffix(updateStr, " || ")
	if ok := utils.UpdateDB(updateStr); !ok {
		fmt.Println("标注通知或私信已读，数据库更新失败")
		return false
	}
	return true
}

// 增加私信
func addMessage(uid int, m common.Message) bool {
	var insertStr = "insert into message set sender=?, receiver=?, detail=?"
	if _, ok := utils.InsertDB(insertStr, uid, m.Receiver, m.Detail); !ok {
		fmt.Println("添加私信，插入数据库错误！")
		return false
	}
	return true
}

// 咨询流程进度处理, -2表示权限错误, -1表示数据库错误, 0表示处理失败(参数不足,非法操作等), 1表示成功
func appointProcess(uID int, userType int, recordID int, operation int, args processArgs) (int, string) {
	var prevStatus string
	var uuID int
	var cID int
	var visitorName string
	var updateStr = "update counseling_record set status=?"

	var queryStr = fmt.Sprintf("select c_id, u_id, status, name from counseling_record where id=%v", recordID)
	rows := utils.QueryDB(queryStr)
	if rows.Next() {
		rows.Scan(&cID, &uuID, &prevStatus, &visitorName)
		if userType == 1 && cID != common.GetCounselorIDByUID(uID) || (userType == 2 && uID != uuID) {
			fmt.Println("咨询流程进度处理，非有效UID")
			return -2, "非有效UID"
		}
	} else {
		fmt.Println("查找咨询记录失败")
		return -1, "查找咨询记录失败"
	}
	rows.Close()

	var counselorName = common.GetCounselorNameByCID(cID)
	var counselorUID = common.GetUserIDByCID(cID)
	var no = common.Notification{Type: "counseling", Title: "", Payload: recordID}
	var notifUID int

	// 逻辑判断处理
	switch prevStatus {
	case "wait_contact":
		// 咨询师操作
		if userType == 1 {
			notifUID = uuID
			if operation == 0 {
				updateStr += fmt.Sprintf(", cancel_reason2=? where id=%v", recordID)
				if args.CancelReason2 != nil {
					utils.UpdateDB(updateStr, "cancel", *(args.CancelReason2))
					no.Title = fmt.Sprintf("咨询师 %v 取消了您的咨询请求，点击查看详情", counselorName)
				} else {
					return 0, "取消理由不能为空"
				}
			} else if operation == 1 {
				updateStr += ", start_time=?"
				if args.StartTime != nil {
					if args.Location != nil && *(args.Location) != "" {
						updateStr += fmt.Sprintf(", location=? where id=%v", recordID)
						utils.UpdateDB(updateStr, "wait_confirm", *(args.StartTime), *(args.Location))
					} else {
						updateStr += fmt.Sprintf(" where id=%v", recordID)
						utils.UpdateDB(updateStr, "wait_confirm", *(args.StartTime))
					}
					no.Title = fmt.Sprintf("咨询师 %v 与您协商了咨询事宜，点击查看详情", counselorName)
				} else {
					return 0, "确认时间不能为空"
				}
			} else {
				return 0, "非法操作，必须为同意(1)或拒绝(0)"
			}
		}

		// 咨询者操作
		if userType == 2 {
			notifUID = counselorUID
			if operation == 1 {
				return 0, "非法操作，咨询者无法主动协商咨询时间"
			} else if operation == 0 {
				if args.CancelReason1 != nil {
					updateStr += fmt.Sprintf(", cancel_reason1=? where id=%v", recordID)
					utils.UpdateDB(updateStr, "cancel", *(args.CancelReason1))
					no.Title = fmt.Sprintf("%v 取消了咨询订单，点击查看详情", visitorName)
				} else {
					return 0, "取消理由不能为空" // 咨询者理由为选择项
				}
			} else {
				return 0, "非法操作，必须为同意(1)或拒绝(0)"
			}
		}

	case "wait_confirm":
		if userType == 2 {
			notifUID = counselorUID
			if operation == 1 {
				if args.StartTime != nil {
					updateStr += ", start_time=?"
					if args.Location != nil {
						updateStr += fmt.Sprintf(", location=? where id=%v", recordID)
						utils.UpdateDB(updateStr, "wait_counseling", *(args.StartTime), *(args.Location))
					} else {
						updateStr += fmt.Sprintf(" where id=%v", recordID)
						utils.UpdateDB(updateStr, "wait_counseling", *(args.StartTime))
					}
				} else {
					updateStr += fmt.Sprintf(" where id=%v", recordID)
					utils.UpdateDB(updateStr, "wait_counseling")
				}
				no.Title = fmt.Sprintf("%v 确认了咨询协商，请按时完成咨询，点击查看详情", visitorName)
			} else if operation == 0 {
				if args.CancelReason1 != nil {
					updateStr += fmt.Sprintf(", cancel_reason1=? where id=%v", recordID)
					utils.UpdateDB(updateStr, "cancel", *(args.CancelReason1))
					no.Title = fmt.Sprintf("%v 取消了咨询，点击查看详情", visitorName)
				} else {
					return 0, "取消理由不能为空" // 咨询者理由为选择项
				}
			} else {
				return 0, "非法操作，必须为同意(1)或拒绝(0)"
			}
		} else {
			return -2, "咨询师无法确认协商结果"
		}

	case "wait_counseling":
		if userType == 2 {
			notifUID = counselorUID
			if operation == 1 {
				updateStr += fmt.Sprintf(" where id=%v", recordID)
				utils.UpdateDB(updateStr, "wait_comment")
			} else if operation == 0 {
				if args.CancelReason1 != nil {
					updateStr += fmt.Sprintf(", cancel_reason1=? where id=%v", recordID)
					utils.UpdateDB(updateStr, "cancel", args.CancelReason1)
					no.Title = fmt.Sprintf("%v 取消了您的咨询，点击查看详情", visitorName)
				} else {
					return 0, "取消理由不能为空" // 此时取消不退款
				}
			} else {
				return 0, "非法操作，必须为同意(1)或拒绝(0)"
			}
		} else {
			return -2, "咨询师无法操作"
		}

	case "wait_comment":
		if userType == 2 {
			notifUID = counselorUID
			if operation == 1 {
				updateStr += fmt.Sprintf(", rating_score=?, rating_text=?, letter=? where id=%v", recordID)
				utils.UpdateDB(updateStr, "finish", *(args.RatingScore), *(args.RatingText), *(args.Letter))
				if args.Letter != nil && *(args.Letter) != "" {
					no.Type = "letter"
					no.Title = fmt.Sprintf("您收到来自 %v 的一封感谢信，点击查看", visitorName)
					updateLetterTime(recordID)
				}
			} else {
				return 0, "非法操作，只能确认评价"
			}
		} else {
			return -2, "咨询师无法操作"
		}

	case "finish":
		// 更新感谢信
		if userType == 2 {
			notifUID = counselorUID
			if args.Letter != nil {
				updateStr += fmt.Sprintf(", letter=? where id=%v", recordID)
				utils.UpdateDB(updateStr, "finish", *(args.Letter))
				updateLetterTime(recordID)
				no.Type = "letter"
				no.Title = fmt.Sprintf("您收到来自 %v 的一封感谢信，点击查看", visitorName)
			} else {
				return 0, "感谢信不能为空"
			}
		} else {
			return -2, "咨询师无法操作"
		}

	default:
		return -1, "咨询状态错误"
	}

	// 更新通知
	if no.Title != "" {
		common.AddNotification(notifUID, no)
	}

	return 1, "ok"
}

// update user info
func updateUserInfo(uID int, userInfo common.User) bool {
	var updateStr = fmt.Sprintf("update user set phone=?, email=? where id=%v", uID)

	if sucess := utils.UpdateDB(updateStr, userInfo.Phone, userInfo.Email); !sucess {
		fmt.Println("更新用户信息失败")
		return false
	}
	return true
}

// update counselor info
func updateCounselorInfo(uID int, info common.CounselorForm) bool {
	var updateStr = fmt.Sprintf("update counselor set motto=?, audio_price=?, video_price=?, ftf_price=? where u_id=%v", uID)

	if success := utils.UpdateDB(updateStr, info.Motto, info.AudioPrice, info.VideoPrice, info.FtfPrice); success {
		common.HandleApplyCity(info.City, uID)
	} else {
		fmt.Println("更新咨询师信息失败")
		return false
	}
	return true
}

// 新增文章或保存草稿
func articleProcess(cID int, args common.Article) bool {
	var dbStr string
	if args.ID != nil {
		dbStr = fmt.Sprintf("update article set cover=?, title=?, content=?, excerpt=?, is_draft=?, category=?, tags=? where id=%v", *(args.ID))
		if success := utils.UpdateDB(dbStr, args.Cover, args.Title, args.Content, args.Excerpt, args.IsDraft, args.Category, args.Tags); success {
			return true
		}
		fmt.Println("更新文章失败")
		return false
	}

	// 首次提交
	dbStr = "insert into article set cover=?, title=?, content=?, excerpt=?, is_draft=?, category=?, tags=?, c_id=?"
	if _, success := utils.InsertDB(dbStr, args.Cover, args.Title, args.Content, args.Excerpt, args.IsDraft, args.Category, args.Tags, cID); success {
		return true
	}
	fmt.Println("插入文章失败")
	return false
}

// 新增文章留言
func addArticleComment(uID int, cmt common.ArticleComment) (bool, int) {
	var insertStr = "insert into article_comment set text=?, a_id=?, ref=?, author=?"
	var id int64
	var success bool

	if id, success = utils.InsertDB(insertStr, cmt.Text, cmt.AID, cmt.RefID, cmt.AuthorID); !success {
		fmt.Println("新增文章留言失败")
		return false, -1
	}

	// 新增通知
	var no common.Notification
	var noUserName = common.GetUserNameByID(uID)
	no.Type = "article"
	no.Payload = cmt.AID
	no.Title = noUserName + "评论了您的文章，点击查看详情"
	if authorID := common.GetAuthorID("article", cmt.AID); authorID != -1 {
		common.AddNotification(authorID, no)
	}

	return true, int(id)
}

// 添加或取消收藏、点赞
func toggleStarLike(uID int, refID int, type1 string, type2 string) bool {
	// validate
	if validate := validateRefByType(refID, type2); !validate {
		return false
	}

	var queryStr = fmt.Sprintf("select id, is_cancel from star_like where u_id=%v and ref_id=%v and type1='%v'and type2='%v'", uID, refID, type1, type2)
	queryRows := utils.QueryDB(queryStr)
	if queryRows.Next() {
		// toggle is_cancel is ok
		var iid int
		var isCancel int
		queryRows.Scan(&iid, &isCancel)
		var updateStr = fmt.Sprintf("update star_like set is_cancel=? where id=%v", iid)
		var iisCancel int
		if isCancel == 0 {
			iisCancel = 1
		} else {
			iisCancel = 0
		}
		queryRows.Close()
		if success := utils.UpdateDB(updateStr, iisCancel); !success {
			fmt.Println("更新收藏点赞失败")
			return false
		}
		return true
	}

	// insert
	var insertStr = "insert into star_like set u_id=?, ref_id=?, type1=?, type2=?"
	if _, success := utils.InsertDB(insertStr, uID, refID, type1, type2); !success {
		fmt.Println("更新收藏点赞失败")
		return false
	}
	return true
}

// 新增阅读量(文章或帖子)
func markReadCounter(uID int, refID int, ttype string) bool {
	// validate
	if validate := validateRefByType(refID, ttype); !validate {
		return false
	}

	if isRead := common.CheckReadStarLike(uID, refID, "read", ttype); isRead {
		return true
	}

	// insert
	var insertStr = "insert into read_count set u_id=?, ref_id=?, type=?"
	if _, success := utils.InsertDB(insertStr, uID, refID, ttype); !success {
		fmt.Println("新增阅读量失败")
		return false
	}
	return true
}

// 校验文章，评论，问答帖子等是否存在
func validateRefByType(refID int, t string) bool {
	var dbTable string
	switch t {
	case "article":
		dbTable = "article"
	case "article_comment":
		dbTable = "article_comment"
	case "ask":
		dbTable = "ask"
	default:
		dbTable = ""
	}
	if dbTable == "" {
		return false
	}

	var queryStr = fmt.Sprintf("select * from %v where id=%v", dbTable, refID)
	rows := utils.QueryDB(queryStr)
	if rows.Next() {
		rows.Close()
		return true
	}
	rows.Close()
	return false
}

// 新增问答帖子
func addAsk(uID int, f askForm) bool {
	var insertStr = "insert into ask set title=?, content=?, is_anonymous=?, tags=?, user_id=?"
	if _, success := utils.InsertDB(insertStr, f.Title, f.Content, f.IsAnony, f.Tags, uID); !success {
		fmt.Println("新增问答帖子，数据库操作失败")
		return false
	}
	return true
}

// 新增问答评论
func addAskComment(uID int, f askCmtForm) (bool, int) {
	var insertStr = "insert into ask_comment set text=?, author=?, reply_to=?, ask_id=?"
	var addCmtID int64
	var success bool
	if f.Ref != nil {
		insertStr += ", ref=?"
		if addCmtID, success = utils.InsertDB(insertStr, f.Text, f.Author, f.ReplyTo, f.AskID, f.Ref); !success {
			fmt.Println("新增问答评论出错")
			return false, -1
		}
	} else {
		if addCmtID, success = utils.InsertDB(insertStr, f.Text, f.Author, 0, f.AskID); !success {
			fmt.Println("新增问答评论出错")
			return false, -1
		}
	}

	// 新增通知
	var no common.Notification
	var noUserName = common.GetUserNameByID(uID)
	no.Type = "ask"
	no.Payload = f.AskID
	no.Title = noUserName + "评论了您的问答，点击查看详情"
	if authorID := common.GetAuthorID("ask", f.AskID); authorID != -1 {
		common.AddNotification(authorID, no)
	}

	return true, int(addCmtID)
}

// 更新感谢信时间
func updateLetterTime(id int) {
	var updateStr = fmt.Sprintf("update counseling_record set letter_time=? where id=%v", id)
	t := time.Now()
	var timeStr = t.Format("2006-01-02 15:04:05")

	if success := utils.UpdateDB(updateStr, timeStr); !success {
		fmt.Println("更新感谢信时间，数据库操作失败")
	}
}
