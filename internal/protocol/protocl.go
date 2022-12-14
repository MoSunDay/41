package protocol

import (
	"41/internal/sender"
	"41/internal/utils"
	"strconv"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"github.com/urfave/cli/v2"
)

func ProtocolHandler(ctx *cli.Context) (err error) {
	var confLogger = utils.GetLogger("ProtocolHandler")
	einterface, snapshotLength, port := ctx.String("interface"), ctx.Int("snapshot-length"), ctx.Int("port")

	handle, err := pcap.OpenLive(einterface, int32(snapshotLength), false, pcap.BlockForever)
	if err != nil {
		confLogger.Fatal(err)
		return
	}
	defer handle.Close()

	filter := "tcp and port " + strconv.Itoa(port)
	if err = handle.SetBPFFilter(filter); err != nil {
		confLogger.Fatal(err)
		return
	} else {
		confLogger.Printf("only capturing tcp port %d packets\n", port)
	}

	sender := sender.NewKafkaSender(ctx)
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	confLogger.Println("start...")

	switch ctx.String("protocol") {
	case "http1":
		http1Handler(packetSource, ctx, sender)
	case "requestf":
		requestfHandler(packetSource, ctx, sender)
	default:
		confLogger.Fatal("unkown protocol")
		return
	}
	return
}
