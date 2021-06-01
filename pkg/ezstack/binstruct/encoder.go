package binstruct

import (
	"bytes"
	"io"
	"reflect"
	"strconv"
)

type encoder struct {
	buf *bytes.Buffer
}

//Encode struct to byte array
func Encode(request interface{}) []byte {
	value := reflect.ValueOf(request)
	buf := bytes.NewBuffer(make([]byte, 0, 200))
	encoder := &encoder{buf}
	encoder.encode(value)
	return buf.Bytes()
}

func (e *encoder) encode(value reflect.Value) {
	switch value.Kind() {
	case reflect.Ptr:
		e.pointer(value)
	case reflect.Struct:
		e.strukt(value)
	}
}

func (e *encoder) strukt(value reflect.Value) {
	var bitmaskBytes uint64
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		fieldType := value.Type().Field(i)
		tags := tags(fieldType.Tag)
		if !(tags.transient() == "true") && checkConditions(tags.cond(), value) {
			switch field.Kind() {
			case reflect.Ptr:
				e.pointer(field)
			case reflect.String:
				e.string(field, tags)
			case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				e.uint(field, tags, &bitmaskBytes)
			case reflect.Array:
				e.array(field, tags)
			case reflect.Slice:
				e.slice(field, tags)
			case reflect.Interface:
				e.pointer(field)
			}
		}
	}
}

func (e *encoder) slice(value reflect.Value, tags tags) {
	length := value.Len()
	e.dynamicLength(length, tags)
	for i := 0; i < length; i++ {
		sliceElement := value.Index(i)
		switch sliceElement.Kind() {
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			e.uint(sliceElement, tags, nil)
		case reflect.String:
			e.string(sliceElement, tags)
		case reflect.Ptr:
			e.pointer(sliceElement)
		case reflect.Struct:
			e.strukt(sliceElement)
		}
	}
}

func (e *encoder) array(value reflect.Value, tags tags) {
	for i := 0; i < value.Len(); i++ {
		e.write(tags.endianness(), value.Index(i))
	}
}

func (e *encoder) string(value reflect.Value, tags tags) {
	s := value.String()
	if tags.hex().nonEmpty() {
		size, _ := strconv.Atoi(string(tags.hex()))
		v, _ := strconv.ParseUint(s[2:], 16, size*8)
		e.writeUint(tags.endianness(), v, size)
	} else {
		e.dynamicLength(len(s), tags)
		e.writeUintSlice(tags.endianness(), []uint8(s))
	}
}

func (e *encoder) uint(value reflect.Value, tags tags, bitmaskBytes *uint64) {
	if tags.bits().nonEmpty() {
		bytes := *bitmaskBytes
		if tags.bitmask() == "start" {
			bytes = 0
		}
		bitmaskBits := bitmaskBits(tags.bits())
		pos := FirstBitPosition(bitmaskBits)
		bytes = bytes | ((value.Uint() << pos) & bitmaskBits)
		if tags.bitmask() == "end" {
			e.writeUint(tags.endianness(), bytes, int(value.Type().Size()))
		}
		*bitmaskBytes = bytes
	} else if tags.bound().nonEmpty() {
		size, _ := strconv.Atoi(string(tags.bound()))
		e.writeUint(tags.endianness(), value.Uint(), size)
	} else {
		e.write(tags.endianness(), value)
	}
}

func serialize(value reflect.Value, w io.Writer) {
	value.MethodByName("Serialize").Call([]reflect.Value{reflect.ValueOf(w)})
}

func (e *encoder) pointer(value reflect.Value) {
	if value.Type().Implements(serializable) {
		serialize(value, e.buf)
	} else {
		e.encode(value.Elem())
	}
}

func (e *encoder) dynamicLength(length int, tags tags) {
	if tags.size().nonEmpty() {
		size, _ := strconv.Atoi(string(tags.size()))
		e.writeUint(tags.endianness(), uint64(length), size)
	}
}

func (e *encoder) write(endianness tag, v reflect.Value) {
	switch v.Kind() {
	case reflect.Uint8:
		e.buf.WriteByte(uint8(v.Uint()))
	case reflect.Uint16:
		e.writeUint(endianness, v.Uint(), 2)
	case reflect.Uint32:
		e.writeUint(endianness, v.Uint(), 4)
	case reflect.Uint64:
		e.writeUint(endianness, v.Uint(), 8)
	}
}

func (e *encoder) writeUintSlice(endianness tag, v interface{}) {
	switch s := v.(type) {
	case []uint8:
		for i := 0; i < len(s); i++ {
			e.writeUint(endianness, uint64(s[i]), 1)
		}
	case []uint16:
		for i := 0; i < len(s); i++ {
			e.writeUint(endianness, uint64(s[i]), 2)
		}
	case []uint32:
		for i := 0; i < len(s); i++ {
			e.writeUint(endianness, uint64(s[i]), 4)
		}
	case []uint64:
		for i := 0; i < len(s); i++ {
			e.writeUint(endianness, uint64(s[i]), 8)
		}
	}
}

func (e *encoder) writeUint(endianness tag, t uint64, size int) {
	if endianness == "be" {
		for i := 0; i < size; i++ {
			e.buf.WriteByte(byte(t >> byte((size-i-1)*8)))
		}
	} else {
		for i := 0; i < size; i++ {
			e.buf.WriteByte(byte(t >> byte(i*8)))
		}
	}
}
