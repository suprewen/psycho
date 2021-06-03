package utils

import (
	"net/http"
	"os"

	"github.com/gorilla/sessions"
)

var (
	key = []byte(os.Getenv("SESSION_KEY"))
	// Store for session management
	Store = sessions.NewCookieStore(key)
)

// InitUserSession to initialize the user session when signin or signup
func InitUserSession(w http.ResponseWriter, r *http.Request, uid int) {
	// Get a session. Get() always returns a session, even if empty.
	session, err := Store.Get(r, "user_session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set expires after one week
	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
	}

	// Set some session values.
	session.Values["auth"] = true
	session.Values["uid"] = uid
	// Save it before we write to the response/return from the handler.
	session.Save(r, w)
}
