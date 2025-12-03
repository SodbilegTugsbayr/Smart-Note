package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/SodbilegTugsbayr/Smart-Note/backend/cmd/web/app"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/cmd/web/validators"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/common/oapi"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/userman"
	"github.com/google/uuid"
)

type signupPayload struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func signup(w http.ResponseWriter, r *http.Request) {
	var payload signupPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		oapi.CustomError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user := &userman.User{
		Name:     strings.TrimSpace(payload.Name),
		Email:    strings.TrimSpace(payload.Email),
		AuthType: userman.AUTH_TYPE_BASIC,
		Role:     userman.ROLE_BASIC,
	}

	if err := validators.ValidateUser(user); err != nil {
		oapi.CustomError(w, http.StatusBadRequest, err.Error())
		return
	}

	if strings.TrimSpace(payload.Password) == "" || len(payload.Password) < 8 {
		oapi.CustomError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	if _, err := app.Users.Get(&userman.User{Email: user.Email}); err == nil {
		oapi.CustomError(w, http.StatusConflict, "user already exists")
		return
	} else if !errors.Is(err, userman.ErrNotFound) {
		oapi.ServerError(w, err)
		return
	}

	passwordHash, err := userman.HashPassword(payload.Password)
	if err != nil {
		oapi.ServerError(w, err)
		return
	}

	user.PasswordHash = passwordHash
	user.UUID = uuid.NewString()
	user.LastLogin = time.Now()
	user.IsVerified = true

	savedUser, err := app.Users.Save(user)
	if err != nil {
		oapi.ServerError(w, err)
		return
	}

	app.Session.Put(r, "auth_user_id", savedUser.ID)
	app.Session.Remove(r, "oauth2_provider_name")

	oapi.SendResp(w, savedUser)
}

func login(w http.ResponseWriter, r *http.Request) {
	var payload loginPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		oapi.CustomError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	email := strings.TrimSpace(payload.Email)
	if email == "" || payload.Password == "" {
		oapi.CustomError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	user, err := app.Users.GetWithAuthTypes(&userman.User{Email: email}, []string{userman.AUTH_TYPE_BASIC})
	if err != nil {
		if errors.Is(err, userman.ErrNotFound) {
			oapi.CustomError(w, http.StatusUnauthorized, "invalid email or password")
		} else {
			oapi.ServerError(w, err)
		}
		return
	}

	if user.PasswordHash == "" || !userman.VerifyPassword(payload.Password, user.PasswordHash) {
		oapi.CustomError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	if !user.IsVerified {
		user.IsVerified = true
	}
	user.LastLogin = time.Now()

	if _, err := app.Users.Save(user); err != nil {
		oapi.ServerError(w, err)
		return
	}

	app.Session.Put(r, "auth_user_id", user.ID)
	app.Session.Remove(r, "oauth2_provider_name")

	oapi.SendResp(w, user)
}
