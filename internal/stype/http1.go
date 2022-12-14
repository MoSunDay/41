package stype

import (
	"41/internal/utils"
	"bufio"
	"bytes"
	"net/http"
	"net/textproto"
	"strconv"
	"strings"
	"time"

	"github.com/google/gopacket/layers"
)

var confLogger = utils.GetLogger("stype/http1")
var CRLF = []byte("\r\n")
var EmptyLine = []byte("\r\n\r\n")
var HeaderDelim = []byte(": ")

// MIMEHeadersEndPos finds end of the Headers section, which should end with empty line.
func MIMEHeadersEndPos(payload []byte) int {
	pos := bytes.Index(payload, EmptyLine)
	if pos < 0 {
		return -1
	}
	return pos + 4
}

// MIMEHeadersStartPos finds start of Headers section
// It just finds position of second line (first contains location and method).
func MIMEHeadersStartPos(payload []byte) int {
	pos := bytes.Index(payload, CRLF)
	if pos < 0 {
		return -1
	}
	return pos + 2 // Find first line end
}

// header return value and positions of header/value start/end.
// If not found, value will be blank, and headerStart will be -1
// Do not support multi-line headers.
func header(payload []byte, name []byte) (value []byte, headerStart, headerEnd, valueStart, valueEnd int) {
	if HasTitle(payload) {
		headerStart = MIMEHeadersStartPos(payload)
		if headerStart < 0 {
			return
		}
	} else {
		headerStart = 0
	}

	var colonIndex int
	for headerStart < len(payload) {
		headerEnd = bytes.IndexByte(payload[headerStart:], '\n')
		if headerEnd == -1 {
			break
		}
		headerEnd += headerStart
		colonIndex = bytes.IndexByte(payload[headerStart:headerEnd], ':')
		if colonIndex == -1 {
			// Malformed header, skip, most likely packet with partial headers
			headerStart = headerEnd + 1
			continue
		}
		colonIndex += headerStart

		if bytes.EqualFold(payload[headerStart:colonIndex], name) {
			valueStart = colonIndex + 1
			valueEnd = headerEnd - 2
			break
		}
		headerStart = headerEnd + 1 // move to the next header
	}
	if valueStart == 0 {
		headerStart = -1
		headerEnd = -1
		valueEnd = -1
		valueStart = -1
		return
	}

	// ignore empty space after ':'
	for valueStart < valueEnd {
		if payload[valueStart] < 0x21 {
			valueStart++
		} else {
			break
		}
	}

	// ignore empty space at end of header value
	for valueEnd > valueStart {
		if payload[valueEnd] < 0x21 {
			valueEnd--
		} else {
			break
		}
	}
	value = payload[valueStart : valueEnd+1]

	return
}

// ParseHeaders Parsing headers from the payload
func ParseHeaders(p []byte) textproto.MIMEHeader {
	// trimming off the title of the request
	if HasTitle(p) {
		headerStart := MIMEHeadersStartPos(p)
		if headerStart > len(p)-1 {
			return nil
		}
		p = p[headerStart:]
	}
	headerEnd := MIMEHeadersEndPos(p)
	if headerEnd > 1 {
		p = p[:headerEnd]
	}
	return GetHeaders(p)
}

// GetHeaders returns mime headers from the payload
func GetHeaders(p []byte) textproto.MIMEHeader {
	reader := textproto.NewReader(bufio.NewReader(bytes.NewReader(p)))
	mime, err := reader.ReadMIMEHeader()
	if err != nil {
		return nil
	}
	return mime
}

// Header returns header value, if header not found, value will be blank
func Header(payload, name []byte) []byte {
	val, _, _, _, _ := header(payload, name)

	return val
}

// SetHeader sets header value. If header not found it creates new one.
// Returns modified request payload
func SetHeader(payload, name, value []byte) []byte {
	_, hs, _, vs, ve := header(payload, name)

	if hs != -1 {
		// If header found we just replace its value
		return utils.Replace(payload, vs, ve+1, value)
	}

	return AddHeader(payload, name, value)
}

// AddHeader takes http payload and appends new header to the start of headers section
// Returns modified request payload
func AddHeader(payload, name, value []byte) []byte {
	mimeStart := MIMEHeadersStartPos(payload)
	if mimeStart < 1 {
		return payload
	}
	header := make([]byte, len(name)+2+len(value)+2)
	copy(header[0:], name)
	copy(header[len(name):], HeaderDelim)
	copy(header[len(name)+2:], value)
	copy(header[len(header)-2:], CRLF)

	return utils.Insert(payload, mimeStart, header)
}

// DeleteHeader takes http payload and removes header name from headers section
// Returns modified request payload
func DeleteHeader(payload, name []byte) []byte {
	_, hs, he, _, _ := header(payload, name)
	if hs != -1 {
		return utils.Cut(payload, hs, he+1)
	}
	return payload
}

// Body returns request/response body
func Body(payload []byte) []byte {
	pos := MIMEHeadersEndPos(payload)
	if pos == -1 || len(payload) <= pos {
		return nil
	}
	return payload[pos:]
}

// Path takes payload and returns request path: Split(firstLine, ' ')[1]
func Path(payload []byte) []byte {
	if !HasRequestTitle(payload) {
		return nil
	}
	start := bytes.IndexByte(payload, ' ') + 1
	end := bytes.IndexByte(payload[start:], ' ')

	return payload[start : start+end]
}

// SetPath takes payload, sets new path and returns modified payload
func SetPath(payload, path []byte) []byte {
	if !HasTitle(payload) {
		return nil
	}
	start := bytes.IndexByte(payload, ' ') + 1
	end := bytes.IndexByte(payload[start:], ' ')

	return utils.Replace(payload, start, start+end, path)
}

// PathParam returns URL query attribute by given name, if no found: valueStart will be -1
func PathParam(payload, name []byte) (value []byte, valueStart, valueEnd int) {
	path := Path(payload)

	paramStart := -1
	if paramStart = bytes.Index(path, append([]byte{'&'}, append(name, '=')...)); paramStart == -1 {
		if paramStart = bytes.Index(path, append([]byte{'?'}, append(name, '=')...)); paramStart == -1 {
			return []byte(""), -1, -1
		}
	}

	valueStart = paramStart + len(name) + 2
	paramEnd := bytes.IndexByte(path[valueStart:], '&')

	// Param can end with '&' (another param), or end of line
	if paramEnd == -1 { // It is final param
		paramEnd = len(path)
	} else {
		paramEnd += valueStart
	}
	return path[valueStart:paramEnd], valueStart, paramEnd
}

// SetPathParam takes payload and updates path Query attribute
// If query param not found, it will append new
// Returns modified payload
func SetPathParam(payload, name, value []byte) []byte {
	path := Path(payload)
	_, vs, ve := PathParam(payload, name)

	if vs != -1 { // If param found, replace its value and set new Path
		newPath := make([]byte, len(path))
		copy(newPath, path)
		newPath = utils.Replace(newPath, vs, ve, value)

		return SetPath(payload, newPath)
	}

	// if param not found append to end of url
	// Adding 2 because of '?' or '&' at start, and '=' in middle
	newParam := make([]byte, len(name)+len(value)+2)

	if bytes.IndexByte(path, '?') == -1 {
		newParam[0] = '?'
	} else {
		newParam[0] = '&'
	}

	// Copy "param=value" into buffer, after it looks like "?param=value"
	copy(newParam[1:], name)
	newParam[1+len(name)] = '='
	copy(newParam[2+len(name):], value)

	// Append param to the end of path
	newPath := make([]byte, len(path)+len(newParam))
	copy(newPath, path)
	copy(newPath[len(path):], newParam)

	return SetPath(payload, newPath)
}

// Method returns HTTP method
func Method(payload []byte) []byte {
	end := bytes.IndexByte(payload, ' ')
	if end == -1 {
		return nil
	}

	return payload[:end]
}

// Methods holds the http methods ordered in ascending order
var Methods = [...]string{
	http.MethodConnect, http.MethodDelete, http.MethodGet,
	http.MethodHead, http.MethodOptions, http.MethodPatch,
	http.MethodPost, http.MethodPut, http.MethodTrace,
}

const (
	//MinRequestCount GET / HTTP/1.1\r\n
	MinRequestCount = 16
	// MinResponseCount HTTP/1.1 200\r\n
	MinResponseCount = 14
	// VersionLen HTTP/1.1
	VersionLen = 8
)

// HasResponseTitle reports whether this payload has an HTTP/1 response title
func HasResponseTitle(payload []byte) bool {
	s := utils.SliceToString(payload)
	if len(s) < MinResponseCount {
		return false
	}
	titleLen := bytes.Index(payload, CRLF)
	if titleLen == -1 {
		return false
	}
	major, minor, ok := http.ParseHTTPVersion(s[0:VersionLen])
	if !(ok && major == 1 && (minor == 0 || minor == 1)) {
		return false
	}
	if s[VersionLen] != ' ' {
		return false
	}
	status, ok := utils.AtoI(payload[VersionLen+1:VersionLen+4], 10)
	if !ok {
		return false
	}
	// only validate status codes mentioned in rfc2616.
	if http.StatusText(status) == "" {
		return false
	}
	// handle cases from #875
	return payload[VersionLen+4] == ' ' || payload[VersionLen+4] == '\r'
}

// HasRequestTitle reports whether this payload has an HTTP/1 request title
func HasRequestTitle(payload []byte) bool {
	s := utils.SliceToString(payload)
	if len(s) < MinRequestCount {
		return false
	}
	titleLen := bytes.Index(payload, CRLF)
	if titleLen == -1 {
		return false
	}
	if strings.Count(s[:titleLen], " ") != 2 {
		return false
	}
	method := string(Method(payload))
	var methodFound bool
	for _, m := range Methods {
		if methodFound = method == m; methodFound {
			break
		}
	}
	if !methodFound {
		return false
	}
	path := strings.Index(s[len(method)+1:], " ")
	if path == -1 {
		return false
	}
	major, minor, ok := http.ParseHTTPVersion(s[path+len(method)+2 : titleLen])
	return ok && major == 1 && (minor == 0 || minor == 1)
}

// HasTitle reports if this payload has an http/1 title
func HasTitle(payload []byte) bool {
	return HasRequestTitle(payload) || HasResponseTitle(payload)
}

// CheckChunked checks HTTP/1 chunked data integrity(https://tools.ietf.org/html/rfc7230#section-4.1)
// and returns the length of total valid scanned chunks(including chunk size, extensions and CRLFs) and
// full is true if all chunks was scanned.
func CheckChunked(bufs ...[]byte) (chunkEnd int, full bool) {
	var buf []byte
	if len(bufs) > 0 {
		buf = bufs[0]
	}
	for chunkEnd < len(buf) {
		sz := bytes.IndexByte(buf[chunkEnd:], '\r')
		if sz < 1 {
			break
		}
		// don't parse chunk extensions https://github.com/golang/go/issues/13135.
		// chunks extensions are no longer a thing, but we do check if the byte
		// following the parsed hex number is ';'
		sz += chunkEnd
		chkLen, ok := utils.AtoI(buf[chunkEnd:sz], 16)
		if !ok && bytes.IndexByte(buf[chunkEnd:sz], ';') < 1 {
			break
		}
		sz++ // + '\n'
		// total length = SIZE + CRLF + OCTETS + CRLF
		allChunk := sz + chkLen + 2
		if allChunk >= len(buf) ||
			buf[sz]&buf[allChunk] != '\n' ||
			buf[allChunk-1] != '\r' {
			break
		}
		chunkEnd = allChunk + 1
		if chkLen == 0 {
			full = true
			break
		}
	}
	return
}

// HasFullPayload checks if this message has full or valid payloads and returns true.
// Message param is optional but recommended on cases where 'data' is storing
// partial-to-full stream of bytes(packets).

type HTTPRequestResponseRecord struct {
	IP           string
	SrcPort      layers.TCPPort
	DstPort      layers.TCPPort
	RequestTime  time.Time
	ResponseTime time.Time
	RequestBody  []byte
	// ResponseBody []HTTPResponse
	ResponseBody [][]byte
	HTTPState    *HTTPState
	ChunkACK     uint32
	IsChunked    bool
}

type HTTPState struct {
	Body           int // body index
	HeaderStart    int
	HeaderEnd      int
	HeaderParsed   bool // we checked necessary headers
	HasFullPayload bool // all chunks has been parsed
	IsChunked      bool // Transfer-Encoding: chunked
	BodyLen        int  // Content-Length's value
	HasTrailer     bool // Trailer header?
	Continue100    bool
}

func partition(array []int, begin, end int) int {
	i := begin + 1
	j := end

	for i < j {
		if array[i] > array[begin] {
			array[i], array[j] = array[j], array[i]
			j--
		} else {
			i++
		}
	}
	if array[i] >= array[begin] {
		i--
	}

	array[begin], array[i] = array[i], array[begin]
	return i
}

// func (r *HTTPRequestResponseRecord) InertSort() {
// 	if len(r.ResponseBody) <= 1 {
// 		return
// 	}
// 	for i := 1; i < len(r.ResponseBody); i++ {
// 		back := r.ResponseBody[i]
// 		j := i - 1
// 		for j >= 0 && back.Seq < r.ResponseBody[j].Seq {
// 			r.ResponseBody[j+1] = r.ResponseBody[j]
// 			j--
// 		}
// 		r.ResponseBody[j+1] = back
// 	}
// }

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
	for _, items := range r.ResponseBody {
		for _, item := range items {
			buffer.WriteByte(item)
		}
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
	// for _, item := range r.ResponseBody {
	// 	for _, it := range item.PlayLoad {
	// 		buffer.WriteByte(it)
	// 	}
	// }
	for _, items := range r.ResponseBody {
		for _, item := range items {
			buffer.WriteByte(item)
		}
	}

	buffer.WriteByte('\n')
	buffer.WriteByte('#')
	buffer.WriteString(strconv.FormatInt(r.ResponseTime.Sub(r.RequestTime).Milliseconds(), 10))

	result := buffer.Bytes()

	return result
}

func (r *HTTPRequestResponseRecord) HasFullPayload() bool {
	// r.InertSort()
	if r.HTTPState == nil {
		r.HTTPState = new(HTTPState)
	}

	if len(r.ResponseBody) == 0 {
		return false
	}
	if !HasResponseTitle(r.ResponseBody[0]) {
		return false
	}

	if r.HTTPState.HeaderStart < 1 {
		for _, data := range r.ResponseBody {
			r.HTTPState.HeaderStart = MIMEHeadersStartPos(data)
			if r.HTTPState.HeaderStart < 0 {
				return false
			} else {
				break
			}
		}
	}

	if r.HTTPState.Body < 1 || r.HTTPState.HeaderEnd < 1 {
		var pos int
		for _, data := range r.ResponseBody {
			endPos := MIMEHeadersEndPos(data)
			if endPos < 0 {
				pos += len(data)
			} else {
				pos += endPos
				r.HTTPState.HeaderEnd = pos
			}

			if endPos > 0 {
				r.HTTPState.Body = pos
				break
			}
		}
	}

	if r.HTTPState.HeaderEnd < 1 {
		return false
	}

	if !r.HTTPState.HeaderParsed {
		var pos int
		for _, data := range r.ResponseBody {
			chunked := Header(data, []byte("Transfer-Encoding"))

			if len(chunked) > 0 && bytes.Index(data, []byte("chunked")) > 0 {
				r.HTTPState.IsChunked = true
				// trailers are generally not allowed in non-chunks body
				r.HTTPState.HasTrailer = len(Header(data, []byte("Trailer"))) > 0
			} else {
				contentLen := Header(data, []byte("Content-Length"))
				r.HTTPState.BodyLen, _ = utils.AtoI(contentLen, 10)
			}

			pos += len(data)

			if string(Header(data, []byte("Expect"))) == "100-continue" {
				r.HTTPState.Continue100 = true
			}

			if r.HTTPState.BodyLen > 0 || pos >= r.HTTPState.Body {
				r.HTTPState.HeaderParsed = true
				break
			}
		}
	}

	bodyLen := 0
	for _, data := range r.ResponseBody {
		bodyLen += len(data)
	}
	bodyLen -= r.HTTPState.Body

	if r.HTTPState.IsChunked {
		// check chunks
		if bodyLen < 1 {
			confLogger.Println("bodyLen < 1")
			return false
		}

		// check trailer headers
		if r.HTTPState.HasTrailer {
			if bytes.HasSuffix(r.ResponseBody[len(r.ResponseBody)-1], []byte("\r\n\r\n")) {
				return true
			}
		} else {
			if bytes.HasSuffix(r.ResponseBody[len(r.ResponseBody)-1], []byte("0\r\n\r\n")) {
				r.HTTPState.HasFullPayload = true
				return true
			}
		}
		// if len(r.ResponseBody) == 2 {
		// 	confLogger.Println(string(r.ResponseBody[0]))
		// 	confLogger.Println(string(r.ResponseBody[1]))
		// }

		return false
	}
	// check for content-length header
	return r.HTTPState.BodyLen == bodyLen
}
