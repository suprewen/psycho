package oauth

import (
	"counseling-system/pkg/common"
	"counseling-system/pkg/utils"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// SignupHandler to handle the req for signup
func SignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		res, _ := ioutil.ReadAll(r.Body)

		var formData SignupForm
		err := json.Unmarshal(res, &formData)
		utils.CheckErr(err)

		// handle signup
		result, uid := signup(formData)

		// init user session
		if result.Code == 1 {
			utils.InitUserSession(w, r, uid)
		}

		resJSON, _ := json.Marshal(result)
		fmt.Fprintf(w, string(resJSON))
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

// SigninHandler to handle the request form signin
func SigninHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		res, _ := ioutil.ReadAll(r.Body)

		var formData SigninForm
		err := json.Unmarshal(res, &formData)
		utils.CheckErr(err)

		// handle user login
		result, uid := signin(formData)

		// init user session
		if result.Code == 1 {
			utils.InitUserSession(w, r, uid)
			// cookie := http.Cookie{Name: "x-token", Value: string(uid)}
			// http.SetCookie(w, &cookie)
		}

		resJSON, _ := json.Marshal(result)
		fmt.Fprintf(w, string(resJSON))
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

// AuthHandler to check if a user is logged in
func AuthHandler(w http.ResponseWriter, r *http.Request) {
	var res common.Response

	// Check if user is authenticated
	uid, userType := common.IsUserLogin(r)
	if uid == -1 {
		res.Code = 0
		res.Message = "用户未登录"
	} else {
		res.Code = 1
		res.Message = "用户已登录"
		res.Data = authData{UserType: userType}
	}

	resJSON, _ := json.Marshal(res)

	// Print secret message
	fmt.Fprintln(w, string(resJSON))
}

// SignoutHandler to handle the signout
func SignoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := utils.Store.Get(r, "user_session")

	// Revoke user authentication
	session.Values["auth"] = false
	session.Save(r, w)
}

// ApplyHandler to handle the apply for counselor
func ApplyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		res, _ := ioutil.ReadAll(r.Body)

		var formData common.CounselorForm
		err := json.Unmarshal(res, &formData)
		utils.CheckErr(err)

		var uid int
		if uid, _ = common.IsUserLogin(r); uid == -1 {
			var result = common.Response{Code: -1, Message: "用户未登录，无法执行操作！"}
			resJSON, _ := json.Marshal(result)
			fmt.Fprintf(w, string(resJSON))
			return
		}

		resString, success := applyCounselor(formData, uid)
		if success {
			fmt.Fprintln(w, resString)
		}
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}
