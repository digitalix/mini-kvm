package main

import (
	"context"
	"fmt"
	"mini-kvm/pkg"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func main() {
	defer fmt.Println("exited")
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{}, 1)
	go func() {
		defer func() {
			done <- struct{}{}
		}()
		if err := pkg.Run(ctx); err != nil {
			panic(err)
		}
	}()

	<-c
	cancel()
	<-done
}
