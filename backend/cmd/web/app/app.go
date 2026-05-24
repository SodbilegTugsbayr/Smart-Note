package app

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/common/apputils"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/common/websocket"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/courseman"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/eguneapi"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/noteman"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/ocrapi"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/quizman"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/userman"
	"github.com/golangcollege/sessions"
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
	Users   *userman.Service
	Notes   *noteman.Service
	Quizzes *quizman.Service
	Courses *courseman.Service

	//AIService
	EguneService *eguneapi.EguneService
	OCRService   *ocrapi.OcrService
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

	DB = apputils.OpenDB(Config.DSN)

	Users = userman.NewService(DB, InfoLog, ErrorLog)
	Courses = courseman.NewService(DB, InfoLog, ErrorLog)
	Notes = noteman.NewService(DB, InfoLog, ErrorLog)
	Quizzes = quizman.NewService(DB, InfoLog, ErrorLog)

	EguneService = eguneapi.NewService(Config.EguneAPI.BaseURL, Config.EguneAPI.APIKey, Config.EguneAPI.Model)
	OCRService = ocrapi.NewService(Config.OCRAPI.BaseURL, Config.OCRAPI.APIKey)

	Session = sessions.New([]byte(Config.SessionSecret))
	Session.Lifetime = 7 * 24 * time.Hour
	Session.Secure = true
	Session.HttpOnly = false
	Session.Path = "/"
}

func Close() {
}
