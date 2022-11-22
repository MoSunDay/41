package protocol

import (
	"41/internal/sender"
	"41/internal/stype/jce"
	"41/internal/stype/requestf"
	"41/internal/utils"
	"encoding/json"
	"log"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/urfave/cli/v2"
)

func requestfHandler(packetSource *gopacket.PacketSource, ctx *cli.Context, sender sender.Sender) {
	var confLogger = utils.GetLogger("http1Handler")
	var ethLayer layers.Ethernet
	var ipLayer layers.IPv4
	var tcpLayer layers.TCP
	port := layers.TCPPort(ctx.Int("port"))
	// cmap := cmap.New(64)
	// localIP := utils.GetLocalIpV4(ctx.String("interface"))

	// go func() {
	// 	ticker := time.NewTicker(10 * time.Second)
	// 	defer ticker.Stop()
	// 	for range ticker.C {
	// 		timestamp := time.Now()
	// 		for _, fd := range cmap.Keys() {
	// 			if tmp, ok := cmap.Get(fd); ok {
	// 				item := tmp.(stype.HTTPRequestResponseRecord)
	// 				if len(item.ResponseBody) == 0 && time.Since(item.RequestTime).Seconds() > 30 {
	// 					item.ResponseBody = []byte("Request timeout")
	// 					item.ResponseTime = timestamp
	// 					item.DstPort = 0
	// 					cmap.Pop(fd)
	// 					// item.EncodeToString()
	// 					sender.Send(&item)
	// 				}
	// 			}
	// 		}
	// 	}
	// }()

	for {
		select {
		case packet := <-packetSource.Packets():
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
					log.Println(string((tcpLayer.BaseLayer.Payload)))
					// timestamp := packet.Metadata().Timestamp
					if tcpLayer.SrcPort == port {
						response := &requestf.ResponsePacket{}
						jceReader := jce.NewReader(tcpLayer.BaseLayer.Payload)
						err := response.ReadFrom(jceReader)
						if err != nil {
							log.Println("&requestf.ResponsePacket{} err:", err)
							continue
						}
						res, _ := json.Marshal(response)
						log.Println(string(res))
					} else {
						request := &requestf.RequestPacket{}
						err := request.ReadFrom(jce.NewReader(tcpLayer.BaseLayer.Payload))
						if err != nil {
							log.Println("&requestf.RequestPacket{} err:", err)
							continue
						}
						res, _ := json.Marshal(request)
						log.Println(string(res))
					}
				}
			}
		case <-ctx.Done():
			confLogger.Println("41 requestf cap exiting")
			return
		}
	}
}
