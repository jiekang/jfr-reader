package main

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
)

const BYTE = 1
const SHORT = 2
const INTEGER = 4
const LONG = 8

var compressed = true

var stringPool = []string{}

type Attribute struct {
	Name  string
	Value string
}

type Element struct {
	Name       string
	Attributes []Attribute
	Elements   []Element
}

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

	if len(os.Args) == 2 {
		chunk(data)
	} else {
		checkMetadata(data)
	}

}

func checkMetadata(b []byte) {
	fmt.Println(b[0:16])

	i := 0

	fmt.Println(readInt(b, &i))
	fmt.Println(b[i])
}

func chunk(b []byte) {
	fmt.Println("chunk")
	fmt.Println()
	pos := 0
	magic := read(b, &pos, 4)
	for _, c := range magic {
		fmt.Printf("%c", c)
	}
	fmt.Println()

	major := read(b, &pos, SHORT)
	fmt.Println("major:", asShort(major))
	minor := read(b, &pos, SHORT)
	fmt.Println("minor:", asShort(minor))

	size := read(b, &pos, LONG)
	fmt.Println("size:", asLong(size))

	cp := read(b, &pos, LONG)
	fmt.Println("constant pool offset:", asLong(cp))

	me := read(b, &pos, LONG)
	fmt.Println("metadata offset:", asLong(me))

	start := read(b, &pos, LONG)
	fmt.Println("start nanos:", asLong(start))

	duration := read(b, &pos, LONG)
	fmt.Println("duration nanos:", asLong(duration))

	cst := read(b, &pos, LONG)
	fmt.Println("chunk start ticks:", asLong(cst))

	tps := read(b, &pos, LONG)
	fmt.Println("ticks per second:", asLong(tps))

	fsfb := read(b, &pos, INTEGER)
	fmt.Println("flags:", fsfb)
	if fsfb[3] == 0 {
		compressed = false
	}

	mpos := int(asLong(me))
	metadata(b, &mpos)

	end := int(asLong(size))
	events(b, &pos, &end)
}

func events(b []byte, pos *int, end *int) {
	fmt.Println()
	fmt.Println("events")
	fmt.Println()

	count := 0

	for *pos < *end {
		event(b, pos)
		count++
	}

	fmt.Println("event count:", count)
}

func event(b []byte, pos *int) {
	size, sr := readInt(b, pos)
	fmt.Println("size:", size)

	tid, tr := readLong(b, pos)
	fmt.Println("type:", tid)

	// if tid == 9265 {
	// 	e := *pos + int(size) - sr - tr
	// 	fmt.Println("BasicEvent: ", *pos, " ", e)
	// 	l1, _ := readLong(b, pos)
	// 	fmt.Println("start time: ", l1)
	// 	l2, _ := readLong(b, pos)
	// 	fmt.Println("duration: ", l2)

	// 	l3, _ := readLong(b, pos)
	// 	fmt.Println("eventThread: ", l3)

	// 	l4, _ := readLong(b, pos)
	// 	fmt.Println("stackTrace: ", l4)

	// 	l5, _ := readLong(b, pos)
	// 	fmt.Println("string type: ", l5)

	// 	length, _ := readLong(b, pos)
	// 	fmt.Println("string length: ", length)

	// 	for i := 0; i < int(length); i++ {
	// 		c, _ := readLong(b, pos)
	// 		fmt.Printf("%c", c)
	// 	}

	// 	fmt.Println()

	// 	*pos = e
	// }
	*pos += int(size) - sr - tr

}

func metadata(b []byte, pos *int) {
	fmt.Println()
	fmt.Println("chunk metadata")
	fmt.Println()

	size, _ := readInt(b, pos)
	fmt.Println("size:", size)

	tid, _ := readLong(b, pos)
	fmt.Println("type:", tid)

	st, _ := readLong(b, pos) // start time
	fmt.Println("start time:", st)
	d, _ := readLong(b, pos) // duration
	fmt.Println("duration:", d)

	mid, _ := readLong(b, pos)
	fmt.Println("metadata id:", mid)

	spSize, _ := readInt(b, pos)
	fmt.Println("string pool size:", spSize)

	for i := 0; i < int(spSize); i++ {
		stringPool = append(stringPool, readMetadataStringPool(b, pos))
	}

	createElement(b, pos, 0)
}

func createElement(b []byte, pos *int, spaces int) Element {
	name := readMetadataString(b, pos)
	element := Element{name, []Attribute{}, []Element{}}

	attrCount, _ := readInt(b, pos)
	printWithSpaces(spaces, "E:", element.Name)
	for i := 0; i < int(attrCount); i++ {
		n := readMetadataString(b, pos)
		v := readMetadataString(b, pos)
		element.Attributes = append(element.Attributes, Attribute{n, v})
		printWithSpaces(spaces+1, "A:", n, v)
	}

	childCount, _ := readInt(b, pos)
	for i := 0; i < int(childCount); i++ {
		c := createElement(b, pos, spaces+4)
		element.Elements = append(element.Elements, c)
	}

	return element
}

func printWithSpaces(spaces int, a ...interface{}) {
	for i := 0; i < spaces; i++ {
		fmt.Print(" ")
	}
	fmt.Println(a...)
}

func readMetadataStringPool(b []byte, pos *int) string {
	encoding, _ := readByte(b, pos)
	if encoding == 4 {
		return parseCharArray(b, pos)
	}
	return ""
}

func readMetadataString(b []byte, pos *int) string {
	index, _ := readInt(b, pos)
	return stringPool[int(index)]
}

func parseCharArray(b []byte, pos *int) string {
	size, _ := readInt(b, pos)
	str := read(b, pos, int(size))
	return string(str)
}

func read(b []byte, pos *int, length int) []byte {
	slice := b[*pos : *pos+length]
	*pos += length
	return slice
}

func readByte(b []byte, pos *int) (byte, int) {
	return read(b, pos, BYTE)[0], 1
}

func readLong(b []byte, pos *int) (uint64, int) {
	if compressed {
		n, l := binary.Uvarint(b[*pos:])
		*pos += l
		return n, l
	}

	return asLong(read(b, pos, LONG)), 8
}

func readInt(b []byte, pos *int) (uint32, int) {
	if compressed {
		n, l := readLong(b, pos)
		return uint32(n), l
	}

	return asInt(read(b, pos, INTEGER)), 4
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
