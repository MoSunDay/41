package protocol

import (
	"41/internal/sender"
	"41/internal/utils"
	"log"

	"github.com/gogo/protobuf/proto"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/urfave/cli/v2"
)

func grpcHandler(packetSource *gopacket.PacketSource, ctx *cli.Context, sender sender.Sender) {
	var confLogger = utils.GetLogger("http1Handler")
	var ethLayer layers.Ethernet
	var ipLayer layers.IPv4
	var tcpLayer layers.TCP
	// port := layers.TCPPort(ctx.Int("port"))
	// localIP := utils.GetLocalIpV4(ctx.String("interface"))

	for {
		select {
		case packet := <-packetSource.Packets():
			{
				parser := gopacket.NewDecodingLayerParser(
					layers.LayerTypeEthernet,
					&ethLayer,
					&ipLayer,
					&tcpLayer,
				)

				foundLayerTypes := []gopacket.LayerType{}
				parser.DecodeLayers(packet.Data(), &foundLayerTypes)
				for _, layerType := range foundLayerTypes {
					if layerType == layers.LayerTypeTCP {
						if len(tcpLayer.BaseLayer.Payload) == 0 {
							continue
						}
						resp, err := proto.BytesToExtensionsMap(tcpLayer.BaseLayer.Payload)
						if err == nil {
							log.Println(resp)
						} else {
							log.Println(err)
						}
					}
				}
			}
		case <-ctx.Done():
			confLogger.Println("41 grpc cap exiting")
			return
		}
	}
}
