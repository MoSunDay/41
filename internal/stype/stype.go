package stype

import (
	"41/internal/utils"
	"bytes"
	"strconv"
	"time"

	"github.com/google/gopacket/layers"
)

var confLogger = utils.GetLogger("stype")

type HTTPRequestResponseRecord struct {
	IP           string
	SrcPort      layers.TCPPort
	DstPort      layers.TCPPort
	RequestTime  time.Time
	ResponseTime time.Time
	RequestBody  []byte
	ResponseBody []byte
}

func (r *HTTPRequestResponseRecord) EncodeToString() string {
	var buffer bytes.Buffer
	buffer.WriteString("#" + strconv.FormatInt(r.RequestTime.UnixMilli(), 10))
	buffer.WriteString("\n#" + r.SrcPort.String() + "->" + r.DstPort.String())
	buffer.WriteString("\n#" + r.IP)
	buffer.WriteString("\n#")
	for _, item := range r.RequestBody {
		buffer.WriteByte(item)
	}
	buffer.WriteString("\n#")
	for _, item := range r.ResponseBody {
		buffer.WriteByte(item)
	}
	buffer.WriteString("\n#")
	buffer.WriteString(strconv.FormatInt(r.ResponseTime.Sub(r.RequestTime).Milliseconds(), 10))
	result := buffer.String()

	confLogger.Println(result)
	confLogger.Println("\n")
	return result
}

func (r *HTTPRequestResponseRecord) EncodeToBytes() []byte {
	var buffer bytes.Buffer
	buffer.WriteString("#" + strconv.FormatInt(r.RequestTime.UnixMilli(), 10))
	buffer.WriteString("\n#" + r.SrcPort.String() + "->" + r.DstPort.String())
	buffer.WriteString("\n#" + r.IP)
	buffer.WriteByte('\n')
	buffer.WriteByte('#')
	for _, item := range r.RequestBody {
		buffer.WriteByte(item)
	}
	buffer.WriteByte('\n')
	buffer.WriteByte('#')
	for _, item := range r.ResponseBody {
		buffer.WriteByte(item)
	}
	buffer.WriteByte('\n')
	buffer.WriteByte('#')
	buffer.WriteString(strconv.FormatInt(r.ResponseTime.Sub(r.RequestTime).Milliseconds(), 10))

	result := buffer.Bytes()

	return result
}
