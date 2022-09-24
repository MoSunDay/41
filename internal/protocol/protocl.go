package protocol

import (
	"41/internal/sender"
	"41/internal/utils"
	"strconv"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"github.com/urfave/cli/v2"
)

func ProtocolHander(ctx *cli.Context) (err error) {
	var confLogger = utils.GetLogger("ProtocolHander")
	einterface, snapshotLength, port := ctx.String("interface"), ctx.Int("snapshot-length"), ctx.Int("port")

	inactive, err := pcap.NewInactiveHandle(einterface)

	defer inactive.CleanUp()
	if err != nil {
		confLogger.Fatalf("inactive handle error: %q, interface: %q", err, einterface)
		return
	}

	err = inactive.SetSnapLen(snapshotLength)
	if err != nil {
		confLogger.Fatalf("snapshot length error: %q, interface: %q", err, einterface)
		return
	}

	err = inactive.SetBufferSize(5242880)
	if err != nil {
		confLogger.Fatalf("handle buffer size error: %q, interface: %q", err, einterface)
		return
	}

	err = inactive.SetTimeout(0)
	if err != nil {
		confLogger.Fatalf("handle buffer timeout error: %q, interface: %q", err, einterface)
		return
	}

	handle, err := inactive.Activate()
	if err != nil {
		confLogger.Fatalf("PCAP Activate device error: %q, interface: %q", err, einterface)
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
		http1Hander(packetSource, ctx, sender)
	default:
		confLogger.Fatal("unkown protocol")
		return
	}
	return
}
