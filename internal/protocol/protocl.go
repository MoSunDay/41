package protocol

import (
	"41/internal/sender"
	"41/internal/stype"
	"41/internal/utils"
	"os"
	"strconv"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"github.com/urfave/cli/v2"
)

func ProtocolHander(ctx *cli.Context) (err error) {
	var confLogger = utils.GetLogger("ProtocolHander")

	einterface, snapshotLength, port := ctx.String("interface"), ctx.Int("snapshot-length"), ctx.Int("port")
	bufferSize := ctx.Int("buffer")
	filter := "tcp and port " + strconv.Itoa(port)

	szFrame, szBlock, numBlocks, err := stype.AfPacketComputeSize(bufferSize, snapshotLength, os.Getpagesize())
	if err != nil {
		confLogger.Fatal(err)
	}

	AfPacketHandle, err := stype.NewAfPacketHandle(einterface, szFrame, szBlock, numBlocks, false, pcap.BlockForever)
	if err != nil {
		confLogger.Fatal(err)
	}

	err = AfPacketHandle.SetBPFFilter(filter, snapshotLength)
	if err != nil {
		confLogger.Fatal(err)
	}
	defer AfPacketHandle.Close()

	sender := sender.NewKafkaSender(ctx)

	source := gopacket.ZeroCopyPacketDataSource(AfPacketHandle)
	confLogger.Println("start...")
	switch ctx.String("protocol") {
	case "http1":
		http1Hander(source, ctx, sender)
	default:
		confLogger.Fatal("unkown protocol")
		return
	}
	return
}
