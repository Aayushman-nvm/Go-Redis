package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

// type for the RESP string we get (prefixes as defined by the Redis Serialization Protocol)
const (
	STRING  = '+'
	ERROR   = '-'
	INTEGER = ':'
	BULK    = '$'
	ARRAY   = '*'
)

// Value represents a parsed RESP value.
type Value struct {
	typ   string
	str   string
	num   int
	bulk  string
	array []Value
}

// Resp wraps a buffered reader and provides methods to parse RESP messages.
type Resp struct {
	reader *bufio.Reader
}

// NewResp creates a RESP parser that reads from the provided io.Reader aka our "conn" from main.go.
// The reader is wrapped in a bufio.Reader for efficient buffered reading.
func NewResp(rd io.Reader) *Resp {
	return &Resp{reader: bufio.NewReader(rd)}
}

// readLine reads bytes until it encounters the RESP line terminator (\r\n) aka CRLF or "registered nurse"-shoutout to ThePrimeagen :)
// It returns the line without the trailing \r\n, the number of bytes read,
// and any error encountered.
func (r *Resp) readLine() (line []byte, n int, err error) {
	for {
		b, err := r.reader.ReadByte()
		if err != nil {
			return nil, 0, err
		}
		n += 1
		//consturct line
		line = append(line, b)

		// Stop reading once the line ends with "\r\n".
		if len(line) >= 2 &&
			line[len(line)-2] == '\r' &&
			line[len(line)-1] == '\n' {
			break
		}
	}
	//return the final constructed line
	return line[:len(line)-2], n, nil
}

// readInteger reads a RESP integer (terminated by \r\n)
// and converts it to a Go int.
func (r *Resp) readInteger() (x int, n int, err error) {
	line, n, err := r.readLine()
	if err != nil {
		return 0, 0, err
	}

	i64, err := strconv.ParseInt(string(line), 10, 64)
	if err != nil {
		return 0, n, err
	}
	return int(i64), n, nil
}

// since first byte determines the RESP type, we store it and switch based on it to use the right parser according to the type
func (r *Resp) Read() (Value, error) {
	_type, err := r.reader.ReadByte()
	if err != nil {
		return Value{}, err
	}

	switch _type {
	case ARRAY:
		return r.readArray() //eg: "*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n"
	case BULK:
		return r.readBulk()
	default:
		fmt.Printf("Unknown type: %v", string(_type))
		return Value{}, nil
	}
}

// Read the array length, then recursively parse each element.
// Each element is itself a complete RESP value.
func (r *Resp) readArray() (Value, error) {
	v := Value{}
	v.typ = "array"
	length, _, err := r.readInteger()
	if err != nil {
		return v, err
	}

	v.array = make([]Value, length)
	for i := 0; i < length; i++ {
		val, err := r.Read()
		if err != nil {
			return v, err
		}

		v.array[i] = val
	}
	return v, nil
}

func (r *Resp) readBulk() (Value, error) {
	v := Value{}
	v.typ = "bulk"
	length, _, err := r.readInteger()
	if err != nil {
		return v, err
	}

	bulk := make([]byte, length)

	r.reader.Read(bulk)
	v.bulk = string(bulk)

	// Consume the trailing "\r\n" after the bulk string data.
	r.readLine()
	return v, nil
}
