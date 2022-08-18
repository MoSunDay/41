package protocol

import (
	"41/internal/utils"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"github.com/urfave/cli/v2"
)

func ProtocolHander(ctx *cli.Context) (err error) {
	var confLogger = utils.GetLogger("ProtocolHander")
	einterface, snapshotLength, filter := ctx.String("interface"), ctx.Int("snapshot-length"), ctx.String("filter")

	handle, err := pcap.OpenLive(einterface, int32(snapshotLength), false, pcap.BlockForever)
	if err != nil {
		confLogger.Fatal(err)
		return
	}
	defer handle.Close()

	if ctx.Args().Present() {
		if err = handle.SetBPFFilter(einterface + " " + string(snapshotLength) + " " + filter); err != nil {
			confLogger.Fatal(err)
			return
		}
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	switch ctx.String("protocol") {
	case "http1":
		http1Hander(packetSource, ctx)
	default:
		confLogger.Fatal("unkown protocol")
		return
	}
	return
}
