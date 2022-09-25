package main

import (
	"41/internal/server"
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gookit/color"
)

func GetRandomString2(n int) string {
	randBytes := make([]byte, n/2)
	rand.Read(randBytes)
	return fmt.Sprintf("%x", randBytes)
}
func main() {
	server := server.NewServer()
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()
	go func() {
		<-c
		cancel()
		os.Exit(0)
	}()

	go func() {
		http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(10 * time.Millisecond)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"status":"OK", "a": "a", "b": {"c": "d"}}`))
		})

		http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(10 * time.Millisecond)
			w.Write([]byte(GetRandomString2(32000)))
		})
		http.ListenAndServe("127.0.0.1:8001", nil)
	}()

	if err := server.RunContext(ctx, os.Args); err != nil {
		color.Errorln(err)
	}
}
