package stype

import (
	"41/internal/utils"
	"bytes"
	"time"

	"github.com/google/gopacket/layers"
)

var confLogger = utils.GetLogger("stype")

type HTTPRequestResponseRecord struct {
	SrcPort      layers.TCPPort
	DstPort      layers.TCPPort
	RequestTime  time.Time
	ResponseTime time.Time
	RequestBody  []byte
	ResponseBody []byte
}

func (r *HTTPRequestResponseRecord) EncodeToString() string {
	var buffer bytes.Buffer
	buffer.WriteString(r.SrcPort.String() + "->" + r.DstPort.String())
	buffer.WriteString("\n")
	for _, item := range r.RequestBody {
		buffer.WriteByte(item)
	}
	buffer.WriteString("\n")
	for _, item := range r.ResponseBody {
		buffer.WriteByte(item)
	}
	buffer.WriteString("\n")
	buffer.WriteString("cost: " + r.ResponseTime.Sub(r.RequestTime).String())
	buffer.WriteString("\n")
	result := buffer.String()

	confLogger.Println(result)
	confLogger.Println("\n")
	return result
}

func (r *HTTPRequestResponseRecord) EncodeToBytes() []byte {
	var buffer bytes.Buffer
	buffer.WriteString(r.SrcPort.String() + "->" + r.DstPort.String())
	buffer.WriteByte('\n')
	for _, item := range r.RequestBody {
		buffer.WriteByte(item)
	}
	buffer.WriteByte('\n')
	for _, item := range r.ResponseBody {
		buffer.WriteByte(item)
	}
	buffer.WriteByte('\n')
	buffer.WriteString("cost: " + r.ResponseTime.Sub(r.RequestTime).String())
	buffer.WriteString("\n")
	result := buffer.Bytes()

	confLogger.Println(result)
	confLogger.Println("\n")
	return result
}
