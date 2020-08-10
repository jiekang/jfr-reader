package main

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
)

const SHORT = 2
const INTEGER = 4
const LONG = 8

var compressed = true

func main() {
	if len(os.Args) < 2 {
		fmt.Println("reader <jfr-filename>")
		return
	}

	data, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		fmt.Println("Unable to read: ", os.Args[1])
		panic(err)
	}

	chunk(data)
}

func chunk(b []byte) {
	pos := 0
	magic := read(b, &pos, 4)
	for _, c := range magic {
		fmt.Printf("%c", c)
	}
	fmt.Println()

	major := read(b, &pos, SHORT)
	fmt.Println("major: ", asShort(major))
	minor := read(b, &pos, SHORT)
	fmt.Println("minor: ", asShort(minor))

	size := read(b, &pos, LONG)
	fmt.Println("size: ", asLong(size))

	cp := read(b, &pos, LONG)
	fmt.Println("constant pool offset: ", asLong(cp))

	me := read(b, &pos, LONG)
	fmt.Println("metadata offset: ", asLong(me))

	start := read(b, &pos, LONG)
	fmt.Println("start nanos: ", asLong(start))

	duration := read(b, &pos, LONG)
	fmt.Println("duration nanos: ", asLong(duration))

	cst := read(b, &pos, LONG)
	fmt.Println("chunk start ticks: ", asLong(cst))

	tps := read(b, &pos, LONG)
	fmt.Println("ticks per second: ", asLong(tps))

	fsfb := read(b, &pos, INTEGER)
	fmt.Println("flags: ", fsfb)
	if fsfb[3] == 0 {
		compressed = false
	}

	mpos := int(asLong(me))
	metadata(b, &mpos)
}

func event(b []byte, pos *int) {
	size := readInt(b, pos)
	fmt.Println("size: ", size)

	tid := readLong(b, pos)
	fmt.Println("type: ", tid)
}

func metadata(b []byte, pos *int) {
	event(b, pos)

	readLong(b, pos) // start time
	readLong(b, pos) // duration

	mid := readLong(b, pos)
	fmt.Println("metadata id: ", mid)

	spSize := readInt(b, pos)
	fmt.Println("string pool size: ", spSize)
}

func read(b []byte, pos *int, length int) []byte {
	slice := b[*pos : *pos+length]
	*pos += length
	return slice
}

func readLong(b []byte, pos *int) uint64 {
	if compressed {
		n, l := binary.Uvarint(b[*pos:])
		*pos += l
		return n
	}

	return asLong(read(b, pos, LONG))
}

func readInt(b []byte, pos *int) uint32 {
	if compressed {
		return uint32(readLong(b, pos))
	}

	return asInt(read(b, pos, INTEGER))
}

func asShort(b []byte) uint16 {
	return binary.BigEndian.Uint16(b)
}

func asInt(b []byte) uint32 {
	return binary.BigEndian.Uint32(b)
}

func asLong(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}
