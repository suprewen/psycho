package main

import (
	"counseling-system/pkg/info"
	"counseling-system/pkg/oauth"
	"counseling-system/pkg/operation"
	"counseling-system/pkg/query"

	"github.com/gorilla/mux"

	"log"
	"net/http"
)

func main() {
	mux := mux.NewRouter()

	// API handler
	oauthHandlers(mux)
	infoHandlers(mux)
	queryHandlers(mux)
	operationHandlers(mux)

	log.Println("Listening on port 8081 ...")
	err := http.ListenAndServe(":8081", mux)
	log.Fatal(err)
}

func oauthHandlers(mux *(mux.Router)) {
	mux.HandleFunc("/api/oauth/signup", oauth.SignupHandler)
	mux.HandleFunc("/api/oauth/signin", oauth.SigninHandler)
	mux.HandleFunc("/api/oauth/auth", oauth.AuthHandler)
	mux.HandleFunc("/api/oauth/signout", oauth.SignoutHandler)
	mux.HandleFunc("/api/oauth/apply", oauth.ApplyHandler)
}

func infoHandlers(mux *(mux.Router)) {
	mux.HandleFunc("/api/info/counselingFilters", info.CounselorFilterHandler)
	mux.HandleFunc("/api/info/pre", info.PreHandler)
	mux.HandleFunc("/api/info/preCounselor", info.PreCounselorHandler)
	mux.HandleFunc("/api/info/articleDraft", info.ArticleDraftHandler)
	mux.HandleFunc("/api/info/askTags", info.AskTagsHandler)
	mux.HandleFunc("/api/info/myArticleList", info.MyArticleListHandler)
	mux.HandleFunc("/api/info/myAskList", info.MyAskListHandler)
}

func queryHandlers(mux *(mux.Router)) {
	mux.HandleFunc("/api/query/counselorList", query.CounselorListHandler)
	mux.HandleFunc("/api/query/newlyCounselors", query.NewlyCounselorsHandler)
	mux.HandleFunc("/api/query/counselor", query.CounselorInfoHandler)
	mux.HandleFunc("/api/query/notifications", query.NotificationHandler)
	mux.HandleFunc("/api/query/messages", query.MessageHandler)
	mux.HandleFunc("/api/query/counselingRecords", query.CounselingRecordHandler)
	mux.HandleFunc("/api/query/articleList", query.ArticleHandler)
	mux.HandleFunc("/api/query/article", query.ArticleHandler)
	mux.HandleFunc("/api/query/askList", query.AskHandler)
	mux.HandleFunc("/api/query/ask", query.AskItemHandler)
	mux.HandleFunc("/api/query/search", query.FuzzyQueryHandler)
	mux.HandleFunc("/api/query/popularList", query.PopularListHandler)
}

func operationHandlers(mux *(mux.Router)) {
	mux.HandleFunc("/api/operation/appoint", operation.AppointHandler)
	mux.HandleFunc("/api/operation/markRead", operation.MarkReadHandler)
	mux.HandleFunc("/api/operation/addMessage", operation.AddMessageHandler)
	mux.HandleFunc("/api/operation/appointProcess/{recordID}/{type}", operation.AppointProcessHandler)
	mux.HandleFunc("/api/operation/updateInfo", operation.UpdateInfoHandler)
	mux.HandleFunc("/api/operation/article", operation.AddArticleHandler)
	mux.HandleFunc("/api/operation/articleComment", operation.AddArticleCommentHandler)
	mux.HandleFunc("/api/operation/starLike", operation.UpdateStarLikeHandler)
	mux.HandleFunc("/api/operation/countRead", operation.ReadCountHandler)
	mux.HandleFunc("/api/operation/addAsk", operation.AddAskHandler)
	mux.HandleFunc("/api/operation/addAskComment", operation.AddAskCommentHandler)
}
