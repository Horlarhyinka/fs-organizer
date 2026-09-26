package main

import (

	"github.com/Horlarhyinka/fs-organizer/internal/cli"
	"github.com/Horlarhyinka/fs-organizer/internal/config"
)

func main() {
	cfg, err := config.LoadConfig(); if err != nil {
		panic(err)
	}
	cli.InitCli(*cfg)
	cli.Execute()
}