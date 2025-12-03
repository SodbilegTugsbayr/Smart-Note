package main

import (
	"net/http"

	"github.com/SodbilegTugsbayr/Smart-Note/backend/cmd/web/app"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/common/oapi"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/common/websocket"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/userman"
)

const (
	GB = 1 << 30
	MB = 1 << 20
	KB = 1 << 10
)

func ping(w http.ResponseWriter, r *http.Request) {
	oapi.SendResp(w, "OK")
}

func logout(w http.ResponseWriter, r *http.Request) {
	app.InfoLog.Println("HEre")
	app.Session.Remove(r, "auth_user_id")
	app.Session.Remove(r, "oauth2_provider_name")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func clearSession(w http.ResponseWriter, r *http.Request) {
	app.Session.Remove(r, "auth_user_id")
	app.Session.Remove(r, "oauth2_provider_name")
	oapi.SendResp(w, "OK")
}

func onSocketConnect(r *http.Request, conn *websocket.Connection) error {
	isAuth, _ := r.Context().Value(app.ContextKeyIsAuthenticated).(bool)
	if !isAuth {
		return nil
	}

	loggedUser := r.Context().Value(app.ContextKeyAuthUser).(*userman.User)

	app.CustomerWSCsMutex.Lock()
	app.CustomerWSCs[int(loggedUser.ID)] = append(app.CustomerWSCs[int(loggedUser.ID)], conn)
	app.CustomerWSCsMutex.Unlock()

	conn.OnClose = func() {
		app.CustomerWSCsMutex.Lock()
		defer app.CustomerWSCsMutex.Unlock()

		for i, c := range app.CustomerWSCs[int(loggedUser.ID)] {
			if c.Key == conn.Key {
				app.CustomerWSCs[int(loggedUser.ID)] = append(app.CustomerWSCs[int(loggedUser.ID)][:i], app.CustomerWSCs[int(loggedUser.ID)][i+1:]...)
				break
			}
		}
	}

	return nil
}

func me(w http.ResponseWriter, r *http.Request) {
	loggedUser := r.Context().Value(app.ContextKeyAuthUser).(*userman.User)
	oapi.SendResp(w, loggedUser)
}
