package app

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/golangcollege/sessions"
	"github.com/peakdot/go-nuxt-example/backend/pkg/common/apputils"
	"github.com/peakdot/go-nuxt-example/backend/pkg/common/mailer"
	"github.com/peakdot/go-nuxt-example/backend/pkg/common/websocket"
	"github.com/peakdot/go-nuxt-example/backend/pkg/easyOAuth2"
	"github.com/peakdot/go-nuxt-example/backend/pkg/userman"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
)

var (
	// Defaults
	ErrorLog *log.Logger
	InfoLog  *log.Logger
	DB       *gorm.DB
	Config   = conf{}
	Location *time.Location
	Session  *sessions.Session

	// Websocket
	CustomerWSConnections *websocket.Websocket
	CustomerWSCs          map[int][]*websocket.Connection
	CustomerWSCsMutex     *sync.RWMutex

	// Services
	Mailer         *mailer.Mailer
	Users          *userman.Service
	GoogleOAuth2   *easyOAuth2.EasyOAuthClient
	FacebookOAuth2 *easyOAuth2.EasyOAuthClient
)

func Init(path string) {
	InfoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	ErrorLog = log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	CustomerWSConnections = websocket.New()
	CustomerWSCs = make(map[int][]*websocket.Connection)
	CustomerWSCsMutex = &sync.RWMutex{}

	loc, err := time.LoadLocation("Asia/Ulaanbaatar")
	if err != nil {
		panic(err)
	}
	Location = loc

	apputils.LoadConfig(&Config, path)

	GoogleOAuth2 = &easyOAuth2.EasyOAuthClient{
		Name: "google",
		Config: &oauth2.Config{
			RedirectURL:  Config.OAuth2.Google.RedirectURL,
			ClientID:     Config.OAuth2.Google.ClientID,
			ClientSecret: Config.OAuth2.Google.ClientSecret,
			Scopes:       Config.OAuth2.Google.Scopes,
			Endpoint:     google.Endpoint,
		},
		UserInfoEndpoint: Config.OAuth2.Google.UserInfoEndpoint,
	}
	FacebookOAuth2 = &easyOAuth2.EasyOAuthClient{
		Name: "facebook",
		Config: &oauth2.Config{
			RedirectURL:  Config.OAuth2.Facebook.RedirectURL,
			ClientID:     Config.OAuth2.Facebook.ClientID,
			ClientSecret: Config.OAuth2.Facebook.ClientSecret,
			Scopes:       Config.OAuth2.Facebook.Scopes,
			Endpoint:     facebook.Endpoint,
		},
		UserInfoEndpoint: Config.OAuth2.Facebook.UserInfoEndpoint,
	}

	DB = apputils.OpenDB(Config.DSN)

	Users = userman.NewService(DB, InfoLog, ErrorLog)

	Session = sessions.New([]byte(Config.SessionSecret))
	Session.Lifetime = 7 * 24 * time.Hour
	Session.Secure = true
	Session.HttpOnly = false
	Session.Path = "/"
}

func Close() {
}
