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

func (v Value) Marshal() []byte {
	switch v.typ {
	case "array":
		return v.marshalArray()
	case "bulk":
		return v.marshalBulk()
	case "string":
		return v.marshalString()
	case "null":
		return v.marshallNull()
	case "error":
		return v.marshallError()
	default:
		return []byte{}
	}
}

func (v Value) marshalString() []byte {
	var bytes []byte
	bytes = append(bytes, STRING)
	bytes = append(bytes, v.str...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

func (v Value) marshalBulk() []byte {
	var bytes []byte
	bytes = append(bytes, BULK)
	bytes = append(bytes, strconv.Itoa(len(v.bulk))...)
	bytes = append(bytes, '\r', '\n')
	bytes = append(bytes, v.bulk...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

// Redis clients (e.g. redis-cli) communicate using RESP. Since commands are
// received in this format, our responses must also be encoded in RESP so the
// client can understand them.

func (v Value) marshalArray() []byte {
	len := len(v.array)
	var bytes []byte
	bytes = append(bytes, ARRAY)
	bytes = append(bytes, strconv.Itoa(len)...)
	bytes = append(bytes, '\r', '\n')

	for i := 0; i < len; i++ {
		bytes = append(bytes, v.array[i].Marshal()...)
	}

	return bytes
}

func (v Value) marshallError() []byte {
	var bytes []byte
	bytes = append(bytes, ERROR)
	bytes = append(bytes, v.str...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

func (v Value) marshallNull() []byte {
	return []byte("$-1\r\n")
}

type Writer struct {
	writer io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{writer: w}
}

func (w *Writer) Write(v Value) error {
	bytes := v.Marshal()

	_, err := w.writer.Write(bytes)
	if err != nil {
		return err
	}

	return nil
}
