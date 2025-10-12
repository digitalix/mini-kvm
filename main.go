package main

import (
	"fmt"
	"mini-kvm/pkg"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func main() {
	defer fmt.Println("exited")
	if err := pkg.Start(); err != nil {
		panic(err)
	}
}
