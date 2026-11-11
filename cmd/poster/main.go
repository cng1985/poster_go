package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/cng1985/poster_go/internal/httpapi"
	"github.com/cng1985/poster_go/internal/library"
	"github.com/cng1985/poster_go/internal/store"
)

func main() {
	addr := flag.String("addr", ":8080", "监听地址")
	dataDir := flag.String("data", "data", "数据目录")
	webDir := flag.String("web", "web", "前端目录")
	flag.Parse()

	if _, err := os.Stat(*webDir + "/index.html"); err != nil {
		log.Fatalf("找不到前端文件 %s/index.html，请在仓库根目录运行", *webDir)
	}
	cat, err := library.Load(*webDir)
	if err != nil {
		log.Fatal(err)
	}
	st, err := store.Open(*dataDir)
	if err != nil {
		log.Fatal(err)
	}
	srv := &http.Server{
		Addr:              *addr,
		Handler:           httpapi.New(st, cat, *webDir).Handler(),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdown); err != nil {
			log.Printf("shutdown: %v", err)
		}
	}()
	shown := *addr
	if strings.HasPrefix(shown, ":") {
		shown = "127.0.0.1" + shown
	}
	log.Printf("PosterGo 已启动 http://%s", shown)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
