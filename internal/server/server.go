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
			Aliases: []string{"s"},
			Value:   262144,
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
			Name:    "protocol",
			Aliases: []string{"p"},
			Usage:   "the target protocol that the package body needs to interpret",
		},
		&cli.StringFlag{
			Name:    "kafka-topic",
			Aliases: []string{"t"},
			Usage:   "kafka's topic",
		},
		&cli.StringFlag{
			Name:    "kafka-addr",
			Aliases: []string{"a"},
			Usage:   "kafka's addr",
		},
		&cli.StringFlag{
			Name:    "kafka-topis-partitions",
			Aliases: []string{"a"},
			Usage:   "the number of kafka topic partitions",
		},
		&cli.IntFlag{
			Name:    "packet-queue-length",
			Aliases: []string{"q"},
			Value:   262144,
		},
	}
	server.Action = protocol.ProtocolHander
	return server
}
