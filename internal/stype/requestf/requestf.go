package requestf

import (
	"fmt"

	"41/internal/stype/jce"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = fmt.Errorf
var _ = jce.Marshal

// RequestPacket struct implement
type RequestPacket struct {
	IVersion     int16             `json:"iVersion"`
	CPacketType  int8              `json:"cPacketType"`
	IMessageType int32             `json:"iMessageType"`
	IRequestId   int32             `json:"iRequestId"`
	SServantName string            `json:"sServantName"`
	SFuncName    string            `json:"sFuncName"`
	SBuffer      []int8            `json:"sBuffer"`
	ITimeout     int32             `json:"iTimeout"`
	Context      map[string]string `json:"context"`
	Status       map[string]string `json:"status"`
}

func (st *RequestPacket) MakesureNotNil() {
}
func (st *RequestPacket) ResetDefault() {
	st.MakesureNotNil()
	st.CPacketType = 0
	st.IMessageType = 0
	st.SServantName = ""
	st.SFuncName = ""
	st.ITimeout = 0
}

// ReadFrom reads  from _is and put into struct.
func (st *RequestPacket) ReadFrom(_is *jce.Reader) error {
	var err error
	var length int32
	var have bool
	var ty byte
	st.ResetDefault()

	err = _is.Read_int16(&st.IVersion, 1, true)
	if err != nil {
		return err
	}

	err = _is.Read_int8(&st.CPacketType, 2, true)
	if err != nil {
		return err
	}

	err = _is.Read_int32(&st.IMessageType, 3, true)
	if err != nil {
		return err
	}

	err = _is.Read_int32(&st.IRequestId, 4, true)
	if err != nil {
		return err
	}

	err = _is.Read_string(&st.SServantName, 5, true)
	if err != nil {
		return err
	}

	err = _is.Read_string(&st.SFuncName, 6, true)
	if err != nil {
		return err
	}

	err, have, ty = _is.SkipToNoCheck(7, true)
	if err != nil {
		return err
	}

	if ty == jce.LIST {
		err = _is.Read_int32(&length, 0, true)
		if err != nil {
			return err
		}

		st.SBuffer = make([]int8, length)
		for i0, e0 := int32(0), length; i0 < e0; i0++ {

			err = _is.Read_int8(&st.SBuffer[i0], 0, false)
			if err != nil {
				return err
			}

		}
	} else if ty == jce.SIMPLE_LIST {

		err, _ = _is.SkipTo(jce.BYTE, 0, true)
		if err != nil {
			return err
		}

		err = _is.Read_int32(&length, 0, true)
		if err != nil {
			return err
		}

		err = _is.Read_slice_int8(&st.SBuffer, length, true)
		if err != nil {
			return err
		}

	} else {
		err = fmt.Errorf("require vector, but not")
		if err != nil {
			return err
		}

	}

	err = _is.Read_int32(&st.ITimeout, 8, true)
	if err != nil {
		return err
	}

	err, have = _is.SkipTo(jce.MAP, 9, true)
	if err != nil {
		return err
	}

	err = _is.Read_int32(&length, 0, true)
	if err != nil {
		return err
	}

	st.Context = make(map[string]string)
	for i1, e1 := int32(0), length; i1 < e1; i1++ {
		var k1 string
		var v1 string

		err = _is.Read_string(&k1, 0, false)
		if err != nil {
			return err
		}

		err = _is.Read_string(&v1, 1, false)
		if err != nil {
			return err
		}

		st.Context[k1] = v1
	}

	err, have = _is.SkipTo(jce.MAP, 10, true)
	if err != nil {
		return err
	}

	err = _is.Read_int32(&length, 0, true)
	if err != nil {
		return err
	}

	st.Status = make(map[string]string)
	for i2, e2 := int32(0), length; i2 < e2; i2++ {
		var k2 string
		var v2 string

		err = _is.Read_string(&k2, 0, false)
		if err != nil {
			return err
		}

		err = _is.Read_string(&v2, 1, false)
		if err != nil {
			return err
		}

		st.Status[k2] = v2
	}

	_ = err
	_ = length
	_ = have
	_ = ty
	return nil
}

// ReadBlock reads struct from the given tag , require or optional.
func (st *RequestPacket) ReadBlock(_is *jce.Reader, tag byte, require bool) error {
	var err error
	var have bool
	st.ResetDefault()

	err, have = _is.SkipTo(jce.STRUCT_BEGIN, tag, require)
	if err != nil {
		return err
	}
	if !have {
		if require {
			return fmt.Errorf("require RequestPacket, but not exist. tag %d", tag)
		}
		return nil
	}

	err = st.ReadFrom(_is)
	if err != nil {
		return err
	}

	err = _is.SkipToStructEnd()
	if err != nil {
		return err
	}
	_ = have
	return nil
}

// WriteTo encode struct to buffer
func (st *RequestPacket) WriteTo(_os *jce.Buffer) error {
	var err error
	st.MakesureNotNil()

	err = _os.Write_int16(st.IVersion, 1)
	if err != nil {
		return err
	}

	err = _os.Write_int8(st.CPacketType, 2)
	if err != nil {
		return err
	}

	err = _os.Write_int32(st.IMessageType, 3)
	if err != nil {
		return err
	}

	err = _os.Write_int32(st.IRequestId, 4)
	if err != nil {
		return err
	}

	err = _os.Write_string(st.SServantName, 5)
	if err != nil {
		return err
	}

	err = _os.Write_string(st.SFuncName, 6)
	if err != nil {
		return err
	}

	err = _os.WriteHead(jce.SIMPLE_LIST, 7)
	if err != nil {
		return err
	}

	err = _os.WriteHead(jce.BYTE, 0)
	if err != nil {
		return err
	}

	err = _os.Write_int32(int32(len(st.SBuffer)), 0)
	if err != nil {
		return err
	}

	err = _os.Write_slice_int8(st.SBuffer)
	if err != nil {
		return err
	}

	err = _os.Write_int32(st.ITimeout, 8)
	if err != nil {
		return err
	}

	err = _os.WriteHead(jce.MAP, 9)
	if err != nil {
		return err
	}

	err = _os.Write_int32(int32(len(st.Context)), 0)
	if err != nil {
		return err
	}

	for k3, v3 := range st.Context {

		err = _os.Write_string(k3, 0)
		if err != nil {
			return err
		}

		err = _os.Write_string(v3, 1)
		if err != nil {
			return err
		}

	}

	err = _os.WriteHead(jce.MAP, 10)
	if err != nil {
		return err
	}

	err = _os.Write_int32(int32(len(st.Status)), 0)
	if err != nil {
		return err
	}

	for k4, v4 := range st.Status {

		err = _os.Write_string(k4, 0)
		if err != nil {
			return err
		}

		err = _os.Write_string(v4, 1)
		if err != nil {
			return err
		}

	}

	_ = err

	return nil
}

// WriteBlock encode struct
func (st *RequestPacket) WriteBlock(_os *jce.Buffer, tag byte) error {
	var err error
	err = _os.WriteHead(jce.STRUCT_BEGIN, tag)
	if err != nil {
		return err
	}

	err = st.WriteTo(_os)
	if err != nil {
		return err
	}

	err = _os.WriteHead(jce.STRUCT_END, 0)
	if err != nil {
		return err
	}
	return nil
}

func (st *RequestPacket) WupDecode(raw []byte) error {
	return st.ReadBlock(jce.NewReader(raw), 0, true)
}

func (st *RequestPacket) WupEncode(_os *jce.Buffer) error {
	return st.WriteBlock(_os, 0)
}

func (st *RequestPacket) WupTypeName() string {
	return "requestf.RequestPacket"
}

// ResponsePacket struct implement
type ResponsePacket struct {
	IVersion     int16             `json:"iVersion"`
	CPacketType  int8              `json:"cPacketType"`
	IRequestId   int32             `json:"iRequestId"`
	IMessageType int32             `json:"iMessageType"`
	IRet         int32             `json:"iRet"`
	SBuffer      []int8            `json:"sBuffer"`
	Status       map[string]string `json:"status"`
	SResultDesc  string            `json:"sResultDesc"`
	Context      map[string]string `json:"context"`
}

func (st *ResponsePacket) MakesureNotNil() {
}
func (st *ResponsePacket) ResetDefault() {
	st.MakesureNotNil()
	st.CPacketType = 0
	st.IMessageType = 0
	st.IRet = 0
}

// ReadFrom reads  from _is and put into struct.
func (st *ResponsePacket) ReadFrom(_is *jce.Reader) error {
	var err error
	var length int32
	var have bool
	var ty byte
	st.ResetDefault()

	err = _is.Read_int16(&st.IVersion, 1, true)
	if err != nil {
		return err
	}

	err = _is.Read_int8(&st.CPacketType, 2, true)
	if err != nil {
		return err
	}

	err = _is.Read_int32(&st.IRequestId, 3, true)
	if err != nil {
		return err
	}

	err = _is.Read_int32(&st.IMessageType, 4, true)
	if err != nil {
		return err
	}

	err = _is.Read_int32(&st.IRet, 5, true)
	if err != nil {
		return err
	}

	err, have, ty = _is.SkipToNoCheck(6, true)
	if err != nil {
		return err
	}

	if ty == jce.LIST {
		err = _is.Read_int32(&length, 0, true)
		if err != nil {
			return err
		}

		st.SBuffer = make([]int8, length)
		for i0, e0 := int32(0), length; i0 < e0; i0++ {

			err = _is.Read_int8(&st.SBuffer[i0], 0, false)
			if err != nil {
				return err
			}

		}
	} else if ty == jce.SIMPLE_LIST {

		err, _ = _is.SkipTo(jce.BYTE, 0, true)
		if err != nil {
			return err
		}

		err = _is.Read_int32(&length, 0, true)
		if err != nil {
			return err
		}

		err = _is.Read_slice_int8(&st.SBuffer, length, true)
		if err != nil {
			return err
		}

	} else {
		err = fmt.Errorf("require vector, but not")
		if err != nil {
			return err
		}

	}

	err, have = _is.SkipTo(jce.MAP, 7, true)
	if err != nil {
		return err
	}

	err = _is.Read_int32(&length, 0, true)
	if err != nil {
		return err
	}

	st.Status = make(map[string]string)
	for i1, e1 := int32(0), length; i1 < e1; i1++ {
		var k1 string
		var v1 string

		err = _is.Read_string(&k1, 0, false)
		if err != nil {
			return err
		}

		err = _is.Read_string(&v1, 1, false)
		if err != nil {
			return err
		}

		st.Status[k1] = v1
	}

	err = _is.Read_string(&st.SResultDesc, 8, false)
	if err != nil {
		return err
	}

	err, have = _is.SkipTo(jce.MAP, 9, false)
	if err != nil {
		return err
	}

	if have {
		err = _is.Read_int32(&length, 0, true)
		if err != nil {
			return err
		}

		st.Context = make(map[string]string)
		for i2, e2 := int32(0), length; i2 < e2; i2++ {
			var k2 string
			var v2 string

			err = _is.Read_string(&k2, 0, false)
			if err != nil {
				return err
			}

			err = _is.Read_string(&v2, 1, false)
			if err != nil {
				return err
			}

			st.Context[k2] = v2
		}
	}

	_ = err
	_ = length
	_ = have
	_ = ty
	return nil
}

// ReadBlock reads struct from the given tag , require or optional.
func (st *ResponsePacket) ReadBlock(_is *jce.Reader, tag byte, require bool) error {
	var err error
	var have bool
	st.ResetDefault()

	err, have = _is.SkipTo(jce.STRUCT_BEGIN, tag, require)
	if err != nil {
		return err
	}
	if !have {
		if require {
			return fmt.Errorf("require ResponsePacket, but not exist. tag %d", tag)
		}
		return nil
	}

	err = st.ReadFrom(_is)
	if err != nil {
		return err
	}

	err = _is.SkipToStructEnd()
	if err != nil {
		return err
	}
	_ = have
	return nil
}

// WriteTo encode struct to buffer
func (st *ResponsePacket) WriteTo(_os *jce.Buffer) error {
	var err error
	st.MakesureNotNil()

	err = _os.Write_int16(st.IVersion, 1)
	if err != nil {
		return err
	}

	err = _os.Write_int8(st.CPacketType, 2)
	if err != nil {
		return err
	}

	err = _os.Write_int32(st.IRequestId, 3)
	if err != nil {
		return err
	}

	err = _os.Write_int32(st.IMessageType, 4)
	if err != nil {
		return err
	}

	err = _os.Write_int32(st.IRet, 5)
	if err != nil {
		return err
	}

	err = _os.WriteHead(jce.SIMPLE_LIST, 6)
	if err != nil {
		return err
	}

	err = _os.WriteHead(jce.BYTE, 0)
	if err != nil {
		return err
	}

	err = _os.Write_int32(int32(len(st.SBuffer)), 0)
	if err != nil {
		return err
	}

	err = _os.Write_slice_int8(st.SBuffer)
	if err != nil {
		return err
	}

	err = _os.WriteHead(jce.MAP, 7)
	if err != nil {
		return err
	}

	err = _os.Write_int32(int32(len(st.Status)), 0)
	if err != nil {
		return err
	}

	for k3, v3 := range st.Status {

		err = _os.Write_string(k3, 0)
		if err != nil {
			return err
		}

		err = _os.Write_string(v3, 1)
		if err != nil {
			return err
		}

	}

	err = _os.Write_string(st.SResultDesc, 8)
	if err != nil {
		return err
	}

	err = _os.WriteHead(jce.MAP, 9)
	if err != nil {
		return err
	}

	err = _os.Write_int32(int32(len(st.Context)), 0)
	if err != nil {
		return err
	}

	for k4, v4 := range st.Context {

		err = _os.Write_string(k4, 0)
		if err != nil {
			return err
		}

		err = _os.Write_string(v4, 1)
		if err != nil {
			return err
		}

	}

	_ = err

	return nil
}

// WriteBlock encode struct
func (st *ResponsePacket) WriteBlock(_os *jce.Buffer, tag byte) error {
	var err error
	err = _os.WriteHead(jce.STRUCT_BEGIN, tag)
	if err != nil {
		return err
	}

	err = st.WriteTo(_os)
	if err != nil {
		return err
	}

	err = _os.WriteHead(jce.STRUCT_END, 0)
	if err != nil {
		return err
	}
	return nil
}

func (st *ResponsePacket) WupDecode(raw []byte) error {
	return st.ReadBlock(jce.NewReader(raw), 0, true)
}

func (st *ResponsePacket) WupEncode(_os *jce.Buffer) error {
	return st.WriteBlock(_os, 0)
}

func (st *ResponsePacket) WupTypeName() string {
	return "requestf.ResponsePacket"
}
