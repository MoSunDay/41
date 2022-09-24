package server

import (
	"41/internal/protocol"

	"github.com/urfave/cli/v2"
)

func NewServer() *cli.App {
	server := cli.NewApp()
	server.Version = "v1.0.0"

	server.Flags = []cli.Flag{
		&cli.IntFlag{
			Name:    "snapshot-length",
			Aliases: []string{"l"},
			Value:   1000,
		},
		&cli.StringFlag{
			Name:     "interface",
			Aliases:  []string{"i"},
			Required: true,
		},
		&cli.StringFlag{
			Name:    "filter",
			Aliases: []string{"f"},
			Usage:   "filter packet",
		},
		&cli.StringFlag{
			Name:  "protocol",
			Usage: "the target protocol that the package body needs to interpret",
			Value: "http1",
		},
		&cli.IntFlag{
			Name:    "port",
			Usage:   "the target port that the package body needs to interpret",
			Aliases: []string{"p"},
			Value:   80,
		},
		&cli.StringFlag{
			Name:    "kafka-topic",
			Aliases: []string{"t"},
			Usage:   "kafka's topic",
			Value:   "test-41",
		},
		&cli.StringFlag{
			Name:    "kafka-host",
			Aliases: []string{"d"},
			Usage:   "kafka's host",
			Value:   "127.0.0.1:9092",
		},
		&cli.IntFlag{
			Name:    "kafka-worker",
			Aliases: []string{"w"},
			Value:   10,
		},
		&cli.IntFlag{
			Name:    "kafka-send-interval",
			Aliases: []string{"s"},
			Value:   5,
		},
		&cli.IntFlag{
			Name:    "kafka-send-queue",
			Aliases: []string{"q"},
			Value:   500000,
		},
	}
	server.Action = protocol.ProtocolHander
	return server
}
