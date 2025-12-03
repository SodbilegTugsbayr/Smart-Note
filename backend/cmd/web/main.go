package main

import (
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SodbilegTugsbayr/Smart-Note/backend/cmd/web/app"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/userman"
)

func main() {
	configPath := flag.String("conf", "./confs/dev.yaml", "Configuration file path")
	flag.Parse()

	app.Init(*configPath)
	defer app.Close()
	closeOnSignalInterrupt(app.Close)

	app.CustomerWSConnections.OnConnect = onSocketConnect

	panicOnError(app.DB.AutoMigrate(
		new(userman.User),
	))

	// addDefaultRecordsIfNotExist()
	srv := &http.Server{
		Addr:         app.Config.Port,
		ErrorLog:     app.ErrorLog,
		Handler:      routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Minute,
		WriteTimeout: 10 * time.Minute,
	}

	app.InfoLog.Printf("Starting server on %s", app.Config.Port)
	app.ErrorLog.Fatal(srv.ListenAndServe())
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

// function to run a cleanup function on signal interruptions such as SIGINT (Ctl+C).
func closeOnSignalInterrupt(cleanupFunc func()) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cleanupFunc()
		os.Exit(0)
	}()
}
