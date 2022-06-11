package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/arbinish/go-bookmarks/bookmarks"
)

func main() {
	infoLog := log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime)
	errLog := log.New(os.Stderr, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile)

	db := bookmarks.NewDB()
	app := bookmarks.NewApp(infoLog, errLog, &db)
	app.Load()
	infoLog.Println("db size", db.Size())
	srv := &http.Server{
		Addr:     ":4912",
		ErrorLog: errLog,
		Handler:  app.Routes(),
	}
	ticker := time.NewTicker(59 * time.Second)
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			<-ticker.C
			app.Save()
		}
	}()

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errLog.Fatalln(err)
		}
	}()

	infoLog.Println("starting server on :4912")

	<-done
	infoLog.Println("server stopped.")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		errLog.Fatalf("server shutdown failed: %v\n", err)
		return
	}
	infoLog.Println("graceful shutdown completed.")
}
