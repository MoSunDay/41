package protocol

import (
	"41/internal/stype"
	"time"

	cmap "github.com/MoSunDay/concurrent-map"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/urfave/cli/v2"
)

func http1Hander(packetSource *gopacket.PacketSource, ctx *cli.Context) {
	// var confLogger = utils.GetLogger("http1Hander")
	var ethLayer layers.Ethernet
	var ipLayer layers.IPv4
	var tcpLayer layers.TCP
	cmap := cmap.New(64)

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			timestamp := time.Now()
			for _, fd := range cmap.Keys() {
				if tmp, ok := cmap.Get(fd); ok {
					item := tmp.(stype.HTTPRequestResponseRecord)
					if len(item.ResponseBody) == 0 && time.Since(item.RequestTime).Seconds() > 30 {
						item.ResponseBody = []byte("Request timeout")
						item.ResponseTime = timestamp
						item.DstPort = 0
						cmap.Pop(fd)
						item.EncodeToString()
					}
				}
			}
		}
	}()
	for packet := range packetSource.Packets() {
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
				timestamp := packet.Metadata().Timestamp
				if tcpLayer.SrcPort == 80 {
					fd := string(tcpLayer.DstPort)
					if tmp, ok := cmap.Get(fd); ok {
						item := tmp.(stype.HTTPRequestResponseRecord)
						item.ResponseBody = tcpLayer.BaseLayer.Payload
						item.ResponseTime = timestamp
						item.DstPort = tcpLayer.SrcPort

						item.EncodeToString()
						cmap.Pop(fd)
					}
				} else {
					rrrecord := stype.HTTPRequestResponseRecord{
						RequestBody: tcpLayer.BaseLayer.Payload,
						RequestTime: timestamp,
						SrcPort:     tcpLayer.SrcPort,
					}
					cmap.Set(string(tcpLayer.SrcPort), rrrecord)
				}
			}
		}
	}
}
