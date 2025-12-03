package easyOAuth2

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

// Source: https://github.com/douglasmakey/oauth2-example 🙏

type EasyOAuthClient struct {
	*oauth2.Config
	Name             string
	UserInfoEndpoint string
}

func (oauth2Client *EasyOAuthClient) RedirectToLogin(w http.ResponseWriter, r *http.Request) error {
	// Create oauthState cookie
	oauthState := oauth2Client.generateStateOauthCookie(w)
	u := oauth2Client.AuthCodeURL(oauthState)
	http.Redirect(w, r, u, http.StatusTemporaryRedirect)
	return nil
}

func (oauth2Client *EasyOAuthClient) HandleCallback(w http.ResponseWriter, r *http.Request) (*oauth2.Token, error) {
	// Read oauthState from Cookie
	oauthState, err := r.Cookie("oauthstate")
	if err != nil {
		return nil, err
	}

	if r.FormValue("state") != oauthState.Value {
		return nil, errors.New("invalid oauth google state")
	}

	token, err := oauth2Client.getAccessToken(r.FormValue("code"))
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (oauth2Client *EasyOAuthClient) generateStateOauthCookie(w http.ResponseWriter) string {
	var expiration = time.Now().Add(365 * 24 * time.Hour)

	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}

	state := base64.URLEncoding.EncodeToString(b)
	cookie := http.Cookie{Name: "oauthstate", Value: state, Expires: expiration}
	http.SetCookie(w, &cookie)

	return state
}

func (oauth2Client *EasyOAuthClient) getAccessToken(code string) (*oauth2.Token, error) {
	// Use code to get token and get user info from Google.
	token, err := oauth2Client.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("code exchange wrong: %s", err.Error())
	}
	return token, nil
}

func (oauth2Client *EasyOAuthClient) GetUserInfo(accessToken string) ([]byte, error) {
	response, err := http.Get(oauth2Client.UserInfoEndpoint + accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	}
	defer response.Body.Close()

	contents, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed read response: %s", err.Error())
	}

	return contents, nil
}
