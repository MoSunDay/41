package protocol

import (
	"41/internal/sender"
	"41/internal/stype"
	"41/internal/utils"
	"encoding/binary"
	"encoding/hex"
	"net"
	"strconv"
	"time"

	cmap "github.com/MoSunDay/concurrent-map"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/urfave/cli/v2"

	_ "fmt"
)

func UUID(SrcIP, DstIP net.IP, SrcPort, DstPort uint16, Ack, Seq uint32, Incoming bool) string {
	var streamID uint64

	if Incoming {
		streamID = uint64(SrcPort)<<48 | uint64(DstPort)<<32 |
			uint64(utils.IP2int(SrcIP))
	} else {
		streamID = uint64(DstPort)<<48 | uint64(SrcPort)<<32 |
			uint64(utils.IP2int(DstIP))
	}

	id := make([]byte, 12)
	binary.BigEndian.PutUint64(id, streamID)

	if Incoming {
		binary.BigEndian.PutUint32(id[8:], Ack)
	} else {
		binary.BigEndian.PutUint32(id[8:], Seq)
	}

	uuidHex := make([]byte, 24)
	hex.Encode(uuidHex[:], id[:])
	return utils.SliceToString(uuidHex)
}

func http1Hander(packetSource *gopacket.PacketSource, ctx *cli.Context, sender sender.Sender) {
	var confLogger = utils.GetLogger("http1Hander")
	var ethLayer layers.Ethernet
	var ipLayer layers.IPv4
	var tcpLayer layers.TCP
	port := layers.TCPPort(ctx.Int("port"))
	cmap := cmap.New(64)
	localIP := utils.GetLocalIpV4(ctx.String("interface"))

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			timestamp := time.Now()
			for _, uuid := range cmap.Keys() {
				if tmp, ok := cmap.Get(uuid); ok {
					item := tmp.(stype.HTTPRequestResponseRecord)
					if len(item.ResponseBody) == 0 && time.Since(item.RequestTime).Seconds() > 5 {
						var content [][]byte
						content = append(content, []byte("Request timeout"))
						item.ResponseBody = content
						item.ResponseTime = timestamp
						item.DstPort = 0
						cmap.Pop(uuid)
						sender.Send(item.EncodeToBytes())
					}
				}
			}
		}
	}()

	workerNum := 1
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

						timestamp := packet.Metadata().Timestamp
						if len(tcpLayer.BaseLayer.Payload) < 100 {
							confLogger.Println(string(tcpLayer.BaseLayer.Payload))
						}

						if tcpLayer.SrcPort == port {
							uuid := strconv.Itoa(int(tcpLayer.DstPort)) + strconv.FormatUint(uint64(tcpLayer.Seq), 10)
							if tmp, ok := cmap.Pop(uuid); ok {
								item := tmp.(stype.HTTPRequestResponseRecord)
								item.ResponseBody = append(item.ResponseBody, tcpLayer.BaseLayer.Payload)
								if item.HasFullPayload() {
									item.ResponseTime = timestamp
									item.DstPort = tcpLayer.SrcPort
									item.EncodeToString()
									sender.Send(item.EncodeToBytes())
								} else {
									item.ChunkACK = tcpLayer.Ack
									uuid := strconv.Itoa(int(tcpLayer.DstPort)) + strconv.FormatUint(uint64(tcpLayer.Seq)+uint64(len(tcpLayer.BaseLayer.Payload)), 10)
									cmap.Set(uuid, item)
								}
							}
						} else {
							uuid := strconv.Itoa(int(tcpLayer.SrcPort)) + strconv.FormatUint(uint64(tcpLayer.Ack), 10)
							if stype.HasRequestTitle(tcpLayer.BaseLayer.Payload) {
								rrrecord := stype.HTTPRequestResponseRecord{
									RequestBody: tcpLayer.BaseLayer.Payload,
									RequestTime: timestamp,
									SrcPort:     tcpLayer.SrcPort,
									IP:          localIP,
								}
								cmap.Set(uuid, rrrecord)
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
