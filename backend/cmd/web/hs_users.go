package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/SodbilegTugsbayr/Smart-Note/backend/cmd/web/app"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/cmd/web/validators"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/common/oapi"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/userman"
)

func getUsers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	size, _ := strconv.Atoi(q.Get("size"))

	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 25
	}

	filter := new(userman.Filter)
	filter.Role = q.Get("role")
	filter.Keyword = q.Get("keyword")

	users, total, err := app.Users.GetAll(filter, page, size)
	if err != nil {
		oapi.ServerError(w, err)
		return
	}

	oapi.SendResp(w, map[string]interface{}{
		"items": users,
		"total": total,
	})
}

func getUser(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(app.ContextKeyChosenUser).(*userman.User)
	oapi.SendResp(w, user)
}

func editUser(w http.ResponseWriter, r *http.Request) {
	var data *userman.User
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		oapi.CustomError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := validators.ValidateUser(data); err != nil {
		oapi.CustomError(w, http.StatusBadRequest, err.Error())
		return
	}

	user, _ := r.Context().Value(app.ContextKeyChosenUser).(*userman.User)

	user.Name = data.Name
	user.PhoneNumber = data.PhoneNumber

	savedUser, err := app.Users.Save(user)
	if err != nil {
		oapi.ServerError(w, err)
		return
	}

	oapi.SendResp(w, savedUser)
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(app.ContextKeyChosenUser).(*userman.User)

	if err := app.Users.Delete(user.ID); err != nil {
		oapi.ServerError(w, err)
		return
	}

	oapi.SendResp(w, user)
}

func updateUserInfo(w http.ResponseWriter, r *http.Request) {
	loggedUser := r.Context().Value(app.ContextKeyAuthUser).(*userman.User)

	var data *userman.User
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		oapi.CustomError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := validators.ValidateUser(data); err != nil {
		oapi.CustomError(w, http.StatusBadRequest, err.Error())
		return
	}

	loggedUser.Name = data.Name
	loggedUser.FacebookURL = data.FacebookURL
	loggedUser.PhoneNumber = data.PhoneNumber

	savedUser, err := app.Users.Save(loggedUser)
	if err != nil {
		oapi.ServerError(w, err)
		return
	}

	oapi.SendResp(w, savedUser)
}
