package binstruct

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"io"
	"testing"

	"github.com/davecgh/go-spew/spew"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) TestEncode1(c *C) {
	type Bitmask struct {
		F0 uint8
		F1 uint16 `bits:"0x0001" bitmask:"start" `
		F2 uint16 `bits:"0x0002"`
		F3 uint16 `bits:"0x0004"`
		F4 uint16 `bits:"0x0008" bitmask:"end"`
		F5 uint8
		F6 uint8 `bits:"0x01" bitmask:"start" `
		F7 uint8 `bits:"0x02" bitmask:"end"`
	}

	type Struct2 struct {
		V uint8
	}

	type Struct struct {
		V1          uint8
		V2          uint8
		BMask       *Bitmask
		Hex         string `hex:"2"`
		Str         string `size:"1"`
		Arr         [2]uint8
		Slice       []uint8
		HexStrings  []string   `size:"1" hex:"2"`
		StructSlice []*Struct2 `size:"1"`
		UintSize    uint32     `bound:"3"`
		UintSizeBe  uint32     `bound:"3" endianness:"be"`
	}

	bitmask := &Bitmask{6, 1, 0, 0, 1, 7, 1, 1}
	str := &Struct{1, 2, bitmask, "0x0A0B", "hello world", [2]uint8{1, 2}, []uint8{3, 4},
		[]string{"0xffaa", "0xaaff"}, []*Struct2{&Struct2{5}, &Struct2{6}}, 1315, 1315}

	payload := Encode(str)
	spew.Dump(payload)
	c.Assert(payload, DeepEquals, []uint8{1, 2, 6, 9, 0, 7, 3, 0x0B, 0x0A, 0x0b, 0x68, 0x65, 0x6c, 0x6c,
		0x6f, 0x20, 0x77, 0x6f, 0x72, 0x6c, 0x64, 1, 2, 3, 4, 2, 0xaa, 0xff, 0xff, 0xaa, 2, 5, 6, 0x23, 0x05, 0, 0, 0x05, 0x23})
}

func (s *MySuite) TestEncode2(c *C) {
	type AfInterPanCtlData interface{}

	type AfInterPanChkData struct {
		PanID    uint16
		Endpoint uint8
	}

	type Capabilities struct {
		Sys   uint16 `bitmask:"start" bits:"0x0001"`
		Mac   uint16 `bits:"0x0002"`
		Nwk   uint16 `bits:"0x0004"`
		Af    uint16 `bits:"0x0008"`
		Zdo   uint16 `bits:"0x0010"`
		Sapi  uint16 `bits:"0x0020"`
		Util  uint16 `bits:"0x0040"`
		Debug uint16 `bits:"0x0080"`
		App   uint16 `bits:"0x0100"`
		Zoad  uint16 `bitmask:"end" bits:"0x1000"`
	}

	type Network struct {
		NeighborPanID   uint16
		LogicalChannel  uint8
		StackProfile    uint8 `bitmask:"start" bits:"0b00001111"`
		ZigbeeVersion   uint8 `bitmask:"end" bits:"0b11110000"`
		BeaconOrder     uint8 `bitmask:"start" bits:"0b00001111"`
		SuperFrameOrder uint8 `bitmask:"end" bits:"0b11110000"`
		PermitJoin      uint8
	}

	type Test struct {
		F0  uint8
		F1  uint16 `endianness:"be"`
		F2  uint16
		F3  uint32
		F4  [8]byte
		F5  [16]byte
		F6  [18]byte
		F7  [32]byte
		F8  [42]byte
		F9  [100]byte
		F11 string `hex:"8"` // string '0x00124b00019c2ee9'
		F12 [2]uint16
		F13 []*Network `size:"1"`
		F14 *Capabilities
		F15 []string `size:"1" hex:"2"`
		F16 AfInterPanCtlData
	}
	test := &Test{F0: 1, F1: 2, F2: 2, F3: 3, F11: "0x00124b00019c2ee9", F12: [2]uint16{4, 5},
		F13: []*Network{&Network{
			NeighborPanID:   500,
			LogicalChannel:  2,
			StackProfile:    3,
			ZigbeeVersion:   4,
			BeaconOrder:     5,
			SuperFrameOrder: 6,
			PermitJoin:      100,
		}},
		F14: &Capabilities{1, 0, 0, 1, 1, 1, 1, 0, 1, 0},
		F15: []string{"0xffaa", "0xaaff"},
		F16: &AfInterPanChkData{1, 2},
	}
	payload := Encode(test)
	c.Assert(payload, DeepEquals, []byte{0x1, 0x0, 0x2, 0x2, 0x0, 0x3, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xe9, 0x2e, 0x9c,
		0x1, 0x0, 0x4b, 0x12, 0x0, 0x4, 0x0, 0x5, 0x0, 0x1, 0xf4, 0x1, 0x2, 0x43, 0x65, 0x64, 0x79, 0x1,
		0x02, 0xaa, 0xff, 0xff, 0xaa, 0x1, 0x0, 0x2})
}

func (s *MySuite) TestEnumEncodeDecode(c *C) {
	type LatencyReq uint8

	const (
		NoLatency LatencyReq = iota
		FastBeacons
		SlowBeacons
	)

	type AfRegister struct {
		EndPoint          uint8
		AppProfID         uint16
		AppDeviceID       uint16
		AddDevVer         uint8
		LatencyReq        LatencyReq
		AppInClusterList  []uint16 `size:"1"`
		AppOutClusterList []uint16 `size:"1"`
	}

	request := &AfRegister{EndPoint: 1, AppProfID: 2, AppDeviceID: 3, AddDevVer: 4,
		LatencyReq: NoLatency, AppInClusterList: []uint16{5, 6}, AppOutClusterList: []uint16{7, 8}}
	payload := Encode(request)
	res := &AfRegister{}
	Decode(payload, res)
	c.Assert(res, DeepEquals, request)
}

func (s *MySuite) TestDecode(c *C) {
	type Status uint8
	const (
		NO Status = iota
		YES
	)

	type Statuses struct {
		Status1 Status `bitmask:"start" bits:"0x01"`
		Status2 Status `bitmask:"end" bits:"0x02"`
	}

	type Network struct {
		NeighborPanID   uint16
		LogicalChannel  uint8
		StackProfile    uint8 `bitmask:"start" bits:"0b00001111"`
		ZigbeeVersion   uint8 `bitmask:"end" bits:"0b11110000"`
		BeaconOrder     uint8 `bitmask:"start" bits:"0b00001111"`
		SuperFrameOrder uint8 `bitmask:"end" bits:"0b11110000"`
		PermitJoin      uint8
	}

	type Test struct {
		F0  uint8
		F1  uint16 `endianness:"be"`
		F2  uint16
		F3  uint32
		F4  [8]byte
		F5  [16]byte
		F6  [18]byte
		F7  [32]byte
		F8  [42]byte
		F9  [100]byte
		F11 string `hex:"8"` // string '0x00124b00019c2ee9'
		F12 [2]uint16
		F13 []*Network `size:"1"`
		F14 []string   `size:"1" hex:"2"`
		F15 string     `size:"1"`
		F16 *Statuses
	}
	test := &Test{F0: 1, F1: 2, F2: 2, F3: 3, F11: "0x00124b00019c2ee9", F12: [2]uint16{4, 5},
		F13: []*Network{&Network{
			NeighborPanID:   500,
			LogicalChannel:  2,
			StackProfile:    3,
			ZigbeeVersion:   4,
			BeaconOrder:     5,
			SuperFrameOrder: 6,
			PermitJoin:      100,
		}},
		F14: []string{"0xffaa", "0xaaff"},
		F15: "hello world",
		F16: &Statuses{YES, NO},
	}
	res := &Test{}
	payload := []byte{0x1, 0x0, 0x2, 0x2, 0x0, 0x3, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xe9, 0x2e, 0x9c,
		0x1, 0x0, 0x4b, 0x12, 0x0, 0x4, 0x0, 0x5, 0x0, 0x1, 0xf4, 0x1, 0x2, 0x43, 0x65, 0x64, 0x02,
		0xaa, 0xff, 0xff, 0xaa, 0x0b, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x20, 0x77, 0x6f, 0x72, 0x6c, 0x64, 0x01}
	Decode(payload, res)

	c.Assert(res, DeepEquals, test)
}

func (s *MySuite) TestEncodeDecode(c *C) {
	type Capabilities struct {
		Sys   uint16 `bitmask:"start" bits:"0x0001"`
		Mac   uint16 `bits:"0x0002"`
		Nwk   uint16 `bits:"0x0004"`
		Af    uint16 `bits:"0x0008"`
		Zdo   uint16 `bits:"0x0010"`
		Sapi  uint16 `bits:"0x0020"`
		Util  uint16 `bits:"0x0040"`
		Debug uint16 `bits:"0x0080"`
		App   uint16 `bits:"0x0100"`
		Zoad  uint16 `bitmask:"end" bits:"0x1000"`
	}

	type Network struct {
		NeighborPanID   uint16
		LogicalChannel  uint8
		StackProfile    uint8 `bitmask:"start" bits:"0b00001111"`
		ZigbeeVersion   uint8 `bitmask:"end" bits:"0b11110000"`
		BeaconOrder     uint8 `bitmask:"start" bits:"0b00001111"`
		SuperFrameOrder uint8 `bitmask:"end" bits:"0b11110000"`
		PermitJoin      uint8
	}

	type Test struct {
		F0  uint8
		F1  uint16 `endianness:"be"`
		F2  uint16
		F3  uint32
		F4  [8]byte
		F5  [16]byte
		F6  [18]byte
		F7  [32]byte
		F8  [42]byte
		F9  [100]byte
		F10 string     `hex:"8"` // string '0x00124b00019c2ee9'
		F11 []uint16   `size:"1"`
		F12 []byte     `size:"1"`
		F13 []*Network `size:"1"`
		F14 *Capabilities
		F15 string `hex:"4"` // string '0x00124b00'
		F16 string `size:"1"`
	}
	networks := []*Network{
		&Network{
			NeighborPanID:   400,
			LogicalChannel:  5,
			StackProfile:    4,
			ZigbeeVersion:   5,
			BeaconOrder:     6,
			SuperFrameOrder: 7,
			PermitJoin:      200,
		},
		&Network{
			NeighborPanID:   500,
			LogicalChannel:  2,
			StackProfile:    3,
			ZigbeeVersion:   4,
			BeaconOrder:     5,
			SuperFrameOrder: 6,
			PermitJoin:      100,
		},
	}
	test := &Test{F0: 1, F1: 2, F2: 2, F3: 3, F4: [8]byte{0, 1, 2, 3, 4, 5, 6, 7}, F10: "0x00124b00019c2ee9",
		F11: []uint16{4, 5}, F12: []byte{1, 2, 3},
		F13: networks, F14: &Capabilities{1, 0, 0, 1, 1, 1, 1, 0, 1, 0}, F15: "0x00124b00",
		F16: "hello world"}
	payload := Encode(test)
	res := &Test{}
	Decode(payload, res)
	c.Assert(res, DeepEquals, test)
}

func (s *MySuite) TestDecodeBitmask(c *C) {
	type Bitmask struct {
		F0 uint8
		F1 uint16 `bitmask:"start" bits:"0x0001"`
		F2 uint16 `bits:"0x0002"`
		F3 uint16 `bits:"0x0004"`
		F4 uint16 `bitmask:"end" bits:"0x0008"`
		F5 uint8
		F6 uint8 `bitmask:"start" bits:"0x01"`
		F7 uint8 `bitmask:"end" bits:"0x02"`
	}
	bitmask := &Bitmask{6, 1, 0, 0, 1, 7, 1, 1}
	res := &Bitmask{}
	payload := []byte{6, 9, 0, 7, 3}
	Decode(payload, res)

	c.Assert(res, DeepEquals, bitmask)
}

func (s *MySuite) TestDecodeUnsizedArray(c *C) {
	type DataStruct struct {
		Data []uint8
	}

	str := &DataStruct{Data: []uint8{6, 1, 0, 0, 1, 7, 1, 1}}
	res := &DataStruct{}
	payload := Encode(str)
	Decode(payload, res)

	c.Assert(res, DeepEquals, str)
}

func (s *MySuite) TestDecodeConditional(c *C) {
	type Struct struct {
		V0 uint8
		V1 uint8
		V2 uint16 `cond:"uint:V1==2;uint:V0==7"`
		V3 uint8  `cond:"uint:V1==3"`
		V4 string `transient:"true"`
	}

	expected := &Struct{V0: 7, V1: 2, V2: 4, V3: 0}
	res := &Struct{}
	payload := []uint8{7, 2, 4, 0}
	Decode(payload, res)

	c.Assert(res, DeepEquals, expected)

	expected = &Struct{V0: 7, V1: 3, V2: 0, V3: 4}
	res = &Struct{}
	payload = []uint8{7, 3, 4}
	Decode(payload, res)

	c.Assert(res, DeepEquals, expected)
}

func (s *MySuite) TestDecodeConditionalDeep(c *C) {
	type Inner struct {
		V uint8
	}

	type Struct struct {
		V1 *Inner
		V2 uint16 `cond:"uint:V1.V==2"`
		V3 uint8  `cond:"uint:V1.V==3"`
	}

	expected := &Struct{V1: &Inner{2}, V2: 4, V3: 0}
	res := &Struct{}
	payload := []uint8{2, 4, 0}
	Decode(payload, res)

	c.Assert(res, DeepEquals, expected)

	expected = &Struct{V1: &Inner{3}, V2: 0, V3: 4}
	res = &Struct{}
	payload = []uint8{3, 4}
	Decode(payload, res)

	c.Assert(res, DeepEquals, expected)
}

func (s *MySuite) TestEncodeConditional(c *C) {
	type Struct struct {
		V1 uint8
		V2 uint16 `cond:"uint:V1==2"`
		V3 uint8  `cond:"uint:V1==3"`
	}

	st := &Struct{V1: 2, V2: 4, V3: 0}
	expected := []uint8{2, 4, 0}
	res := Encode(st)

	c.Assert(res, DeepEquals, expected)

	st = &Struct{V1: 3, V2: 0, V3: 4}
	expected = []uint8{3, 4}
	res = Encode(st)

	c.Assert(res, DeepEquals, expected)
}

func (s *MySuite) TestEncodeConditionalDeep(c *C) {
	type Inner struct {
		V uint8
	}

	type Struct struct {
		V1 *Inner
		V2 uint16 `cond:"uint:V1.V==2"`
		V3 uint8  `cond:"uint:V1.V==3"`
	}

	st := &Struct{V1: &Inner{2}, V2: 4, V3: 0}
	expected := []uint8{2, 4, 0}
	res := Encode(st)

	c.Assert(res, DeepEquals, expected)

	st = &Struct{V1: &Inner{3}, V2: 0, V3: 4}
	expected = []uint8{3, 4}
	res = Encode(st)

	c.Assert(res, DeepEquals, expected)
}

func (s *MySuite) TestDecodeStruct(c *C) {
	type Bitmask struct {
		F0 uint8
		F1 uint16 `bits:"0x0001" bitmask:"start" `
		F2 uint16 `bits:"0x0002"`
		F3 uint16 `bits:"0x0004"`
		F4 uint16 `bits:"0x0008" bitmask:"end"`
		F5 uint8
		F6 uint8 `bits:"0x01" bitmask:"start" `
		F7 uint8 `bits:"0x02" bitmask:"end"`
	}
	type Struct3 struct {
		V1 uint8
		V2 uint8
	}
	type Struct2 struct {
		V           string `hex:"2"`
		Ui16        uint16
		BMask       *Bitmask
		Arr         [2]uint8
		Slice       []uint8    `size:"1"`
		StructSlice []*Struct3 `size:"1"`
	}
	type Struct struct {
		V          string `hex:"2"`
		S          *Struct2
		Ui16       uint16
		UintSize   uint32 `bound:"3"`
		UintSizeBe uint32 `bound:"3" endianness:"be"`
	}

	bitmask := &Bitmask{6, 1, 0, 0, 1, 7, 1, 1}

	payload := []uint8{1, 2, 1, 2, 1, 2, 6, 9, 0, 7, 3, 5, 6, 2, 7, 8, 2, 1, 2, 3, 4, 1, 2, 0x23, 0x05, 0, 0, 0x05, 0x23}
	str := &Struct{"0x0201", &Struct2{"0x0201", 0x201, bitmask, [2]uint8{5, 6}, []uint8{7, 8},
		[]*Struct3{&Struct3{1, 2}, &Struct3{3, 4}}}, 0x201, 1315, 1315}
	res := &Struct{}
	Decode(payload, res)

	c.Assert(res, DeepEquals, str)
}

type Struct4 struct {
	V1 uint8
	V2 uint8
}

func Fn() (res *Struct4) {
	payload := []uint8{1, 2}
	Decode(payload, &res)
	return res
}

func (s *MySuite) TestDecodeFunctionReturnValue(c *C) {
	str := &Struct4{1, 2}
	res := Fn()

	c.Assert(res, DeepEquals, str)
}

type BenchInner struct {
	F1 uint8
}

type Bench struct {
	F0  uint8
	F1  uint16 `endianness:"be"`
	F2  uint16
	F3  uint32
	F4  [8]byte
	F10 string        `hex:"8"` // string '0x00124b00019c2ee9'
	F11 []uint16      `size:"1"`
	F12 []byte        `size:"1"`
	F13 []*BenchInner `size:"1"`
	F14 *BenchInner
	F15 string `hex:"4"` // string '0x00124b00'
	F16 string `size:"1"`
}

func newBenchStruct() *Bench {
	slice := []*BenchInner{
		&BenchInner{200},
		&BenchInner{100}}

	return &Bench{F0: 1, F1: 2, F2: 2, F3: 3, F10: "0x00124b00019c2ee9",
		F11: []uint16{4, 5}, F12: []byte{1, 2, 3},
		F13: slice, F14: &BenchInner{1}, F15: "0x00124b00",
		F16: "hello world"}
}

func Benchmark_Gob(b *testing.B) {
	v := newBenchStruct()

	buffer := &bytes.Buffer{}
	gob.NewEncoder(buffer).Encode(v)
	payload := buffer.Bytes()

	b.Run("encode", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			gob.NewEncoder(new(bytes.Buffer)).Encode(v)
		}
	})

	b.Run("decode", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		out := &Bench{}
		for n := 0; n < b.N; n++ {
			gob.NewDecoder(bytes.NewBuffer(payload)).Decode(out)
		}
	})
}

func Benchmark_Payload(b *testing.B) {
	v := newBenchStruct()

	payload := Encode(v)

	b.Run("encode", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			Encode(v)
		}
	})

	b.Run("decode", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		out := &Bench{}
		for n := 0; n < b.N; n++ {
			Decode(payload, out)
		}
	})
}

func Benchmark_JSON(b *testing.B) {
	v := newBenchStruct()
	enc, _ := json.Marshal(v)

	b.Run("encode", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			json.Marshal(v)
		}
	})

	b.Run("decode", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		out := &Bench{}
		for n := 0; n < b.N; n++ {
			json.Unmarshal(enc, out)
		}
	})
}

type MySerializable struct {
	AttributeValue interface{}
}

func (m *MySerializable) Serialize(w io.Writer) {
	var b [1]byte
	switch t := m.AttributeValue.(type) {
	case string:
		b[0] = 1
		w.Write(b[:])
		b[0] = uint8(len(t))
		w.Write(b[:])
		w.Write([]byte(t))
	case uint8:
		b[0] = 2
		w.Write(b[:])
		b[0] = t
		w.Write(b[:])
	}
}

func (m *MySerializable) Deserialize(r io.Reader) {
	var b [1]byte
	r.Read(b[:])
	switch b[0] {
	case 1:
		r.Read(b[:])
		buf := make([]byte, b[0])
		r.Read(buf)
		m.AttributeValue = string(buf)
	case 2:
		r.Read(b[:])
		m.AttributeValue = b[0]
	}
}

func (s *MySuite) TestEncodeSerializable(c *C) {
	type Struct struct {
		M *MySerializable
	}
	v := &Struct{&MySerializable{uint8(5)}}
	res := Encode(v)
	c.Assert(res, DeepEquals, []byte{2, 5})
}

func (s *MySuite) TestDecodeSerializable(c *C) {
	type Struct struct {
		M *MySerializable
	}
	res := &Struct{}
	Decode([]byte{2, 5}, res)
	c.Assert(res, DeepEquals, &Struct{&MySerializable{uint8(5)}})
}

func (s *MySuite) TestDecodeSliceOfStructs(c *C) {
	type Struct struct {
		M []*MySerializable
	}

	res := &Struct{}
	Decode([]byte{2, 5, 2, 6}, res)
	c.Assert(res, DeepEquals, &Struct{[]*MySerializable{&MySerializable{uint8(5)}, &MySerializable{uint8(6)}}})
}
