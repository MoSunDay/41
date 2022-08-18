package main

import (
	"41/internal/server"
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/gookit/color"
)

func main() {
	server := server.NewServer()
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		<-c
		cancel()
	}()

	if err := server.RunContext(ctx, os.Args); err != nil {
		color.Errorln(err)
	}
}
