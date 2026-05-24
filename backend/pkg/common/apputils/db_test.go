package apputils

import (
	"os"
	"strings"
	"testing"
)

func TestOpenDBConnectsWithValidDSN(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("TEST_DATABASE_URL"))
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to run database-backed tests")
	}

	db := OpenDB(dsn)
	if db == nil {
		t.Fatal("OpenDB() = nil, want non-nil *gorm.DB")
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB() error = %v", err)
	}
	defer sqlDB.Close()

	if err := sqlDB.Ping(); err != nil {
		t.Fatalf("Ping() error = %v", err)
	}
}
