package testdb

import (
	"io"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/courseman"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/noteman"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/quizman"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/userman"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func Open(t testing.TB) *gorm.DB {
	t.Helper()

	dsn := strings.TrimSpace(os.Getenv("TEST_DATABASE_URL"))
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to run database-backed tests")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlogger.New(log.New(io.Discard, "", 0), gormlogger.Config{}),
	})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		t.Fatalf("ping test database: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)

	if err := db.Exec("SELECT pg_advisory_lock(20260524)").Error; err != nil {
		t.Fatalf("lock test database: %v", err)
	}

	if err := db.AutoMigrate(
		new(userman.User),
		new(courseman.Course),
		new(noteman.Note),
		new(quizman.Quiz),
		new(quizman.QuizResult),
	); err != nil {
		t.Fatalf("auto migrate test database: %v", err)
	}

	Reset(t, db)
	t.Cleanup(func() {
		Reset(t, db)
		if err := db.Exec("SELECT pg_advisory_unlock(20260524)").Error; err != nil {
			t.Fatalf("unlock test database: %v", err)
		}
		_ = sqlDB.Close()
	})

	return db
}

func Reset(t testing.TB, db *gorm.DB) {
	t.Helper()

	if err := db.Exec("TRUNCATE TABLE quiz_results, quizzes, notes, courses, users RESTART IDENTITY CASCADE").Error; err != nil {
		t.Fatalf("reset test database: %v", err)
	}
}

func DiscardLogger() *log.Logger {
	return log.New(io.Discard, "", 0)
}
