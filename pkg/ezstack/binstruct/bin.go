package binstruct

import (
	"encoding/binary"
	"io"
	"reflect"
	"strconv"
	"strings"
)

type tag string

func (t tag) nonEmpty() bool {
	return len(t) > 0
}

type tags reflect.StructTag

func (t tags) hex() tag {
	return tag(reflect.StructTag(t).Get("hex"))
}

func (t tags) cond() tag {
	return tag(reflect.StructTag(t).Get("cond"))
}

func (t tags) endianness() tag {
	return tag(reflect.StructTag(t).Get("endianness"))
}

func (t tags) size() tag {
	return tag(reflect.StructTag(t).Get("size"))
}

func (t tags) bitmask() tag {
	return tag(reflect.StructTag(t).Get("bitmask"))
}

func (t tags) bits() tag {
	return tag(reflect.StructTag(t).Get("bits"))
}

func (t tags) bound() tag {
	return tag(reflect.StructTag(t).Get("bound"))
}

func (t tags) transient() tag {
	return tag(reflect.StructTag(t).Get("transient"))
}

func valueConvertTo(value reflect.Value, typ reflect.Type) reflect.Value {
	return value.Convert(typ)
}

func bitmaskBits(value tag) (bitmaskBits uint64) {
	prefix := string(value[:2])
	bitmask := string(value[2:])
	if prefix == "0x" {
		bitmaskBits, _ = strconv.ParseUint(bitmask, 16, len(bitmask)*4)
		return
	} else if prefix == "0b" {
		bitmaskBits, _ = strconv.ParseUint(bitmask, 2, len(bitmask))
		return
	}
	panic("Unsupported prefix: " + prefix)
}

func order(endianness tag) binary.ByteOrder {
	if endianness == "be" {
		return binary.BigEndian
	}
	return binary.LittleEndian
}

func checkConditions(cond tag, parent reflect.Value) bool {
	if cond.nonEmpty() {
		conditions := strings.Split(string(cond), ";")
		for _, c := range conditions {
			if !checkCondition(c, parent) {
				return false
			}
		}
	}
	return true
}

func checkCondition(cond string, parent reflect.Value) bool {
	v := strings.Split(cond, ":")
	t := v[0]
	c := v[1]
	var op string
	switch {
	case strings.Contains(c, "=="):
		op = "=="
	case strings.Contains(c, "!="):
		op = "!="
	}
	v = strings.Split(c, op)
	l := v[0]
	r := v[1]
	getField := func() reflect.Value {
		v := parent
		pathElements := strings.Split(l, ".")
		for _, e := range pathElements {
			v = unpoint(v.FieldByName(e))
		}
		return v
	}
	switch t {
	case "uint":
		lv := uint64(getField().Uint())
		n, _ := strconv.Atoi(r)
		rv := uint64(n)
		switch op {
		case "==":
			return lv == rv
		case "!=":
			return lv != rv
		}
	default:
		panic("Unknown condition type: " + t)
	}
	return true
}

func unpoint(p reflect.Value) reflect.Value {
	if p.Kind() == reflect.Ptr {
		return unpoint(p.Elem())
	} else {
		return p
	}
}

type Serializable interface {
	Serialize(w io.Writer)
	Deserialize(r io.Reader)
}

var serializable = reflect.TypeOf((*Serializable)(nil)).Elem()
