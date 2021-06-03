package operation

// 流程处理参数
type processArgs struct {
	StartTime     *string  `json:"startTime"`
	Location      *string  `json:"location"`
	CancelReason1 *string  `json:"cancelReason1"`
	CancelReason2 *string  `json:"cancelReason2"`
	RatingScore   *float64 `json:"ratingScore"`
	RatingText    *string  `json:"ratingText"`
	Letter        *string  `json:"letter"`
}

// 新增问答Form
type askForm struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	IsAnony bool   `json:"isAnony"`
	Tags    string `json:"tags"`
}

// 新增问答评论Form
type askCmtForm struct {
	Text    string `json:"text"`
	Author  int    `json:"author"`
	ReplyTo *int   `json:"replyTo"`
	Ref     *int   `json:"ref"`
	AskID   int    `json:"askID"`
}
