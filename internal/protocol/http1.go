package protocol

import (
	"41/internal/sender"
	"41/internal/stype"
	"41/internal/utils"
	"fmt"
	"strconv"
	"time"

	cmap "github.com/MoSunDay/concurrent-map"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/urfave/cli/v2"

	_ "fmt"
)

func http1Hander(packetSource *gopacket.PacketSource, ctx *cli.Context, sender sender.Sender) {
	var confLogger = utils.GetLogger("http1Hander")
	var ethLayer layers.Ethernet
	var ipLayer layers.IPv4
	var tcpLayer layers.TCP
	port := layers.TCPPort(ctx.Int("port"))
	cmap := cmap.New(64)
	localIP := utils.GetLocalIpV4(ctx.String("interface"))

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			timestamp := time.Now()
			for _, fd := range cmap.Keys() {
				if tmp, ok := cmap.Get(fd); ok {
					item := tmp.(stype.HTTPRequestResponseRecord)
					if len(item.ResponseBody) == 0 && time.Since(item.RequestTime).Seconds() > 30 {
						var content [][]byte
						content = append(content, []byte("Request timeout"))
						item.ResponseBody = content
						item.ResponseTime = timestamp
						item.DstPort = 0
						cmap.Pop(fd)
						sender.Send(item.EncodeToBytes())
					}
				}
			}
		}
	}()

	workerNum := 25
	packetChans := make([]chan gopacket.Packet, workerNum)
	for i := 0; i < workerNum; i++ {
		ch := make(chan gopacket.Packet, 2500)
		packetChans[i] = ch
		go func(ch chan gopacket.Packet) {
			for packet := range ch {
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
						fmt.Println(string(tcpLayer.BaseLayer.Payload))
						fmt.Println("size:", len(tcpLayer.BaseLayer.Payload))
						timestamp := packet.Metadata().Timestamp
						if tcpLayer.SrcPort == port {
							fd := strconv.Itoa(int(tcpLayer.DstPort)) + strconv.FormatUint(uint64(tcpLayer.Seq), 10)
							if tmp, ok := cmap.Get(fd); ok {
								cmap.Pop(fd)
								item := tmp.(stype.HTTPRequestResponseRecord)
								item.ResponseBody = append(item.ResponseBody, tcpLayer.BaseLayer.Payload)
								if item.HasFullPayload() {
									item.ResponseTime = timestamp
									item.DstPort = tcpLayer.SrcPort
									sender.Send(item.EncodeToBytes())
								} else {
									cmap.Set(strconv.Itoa(int(tcpLayer.SrcPort))+strconv.FormatUint(uint64(tcpLayer.Ack), 10), item)
								}
							}
						} else {
							if stype.HasRequestTitle(tcpLayer.BaseLayer.Payload) {
								rrrecord := stype.HTTPRequestResponseRecord{
									RequestBody: tcpLayer.BaseLayer.Payload,
									RequestTime: timestamp,
									SrcPort:     tcpLayer.SrcPort,
									IP:          localIP,
								}
								cmap.Set(strconv.Itoa(int(tcpLayer.SrcPort))+strconv.FormatUint(uint64(tcpLayer.Ack), 10), rrrecord)
							}
						}
					}
				}
			}
		}(ch)
	}

	for {
		select {
		case packet := <-packetSource.Packets():
			packetChans[packet.Metadata().Timestamp.Nanosecond()%workerNum] <- packet
		case <-ctx.Done():
			confLogger.Println("41 http1 cap exiting")
			return
		}
	}
}
