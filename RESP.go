package main

import (
	"bufio"
	"io"
	"strconv"
)

// type for the RESP string we get
const (
	STRING  = '+'
	ERROR   = '-'
	INTEGER = ':'
	BULK    = '$'
	ARRAY   = '*'
)

// type for the RESP string we get
type Value struct {
	typ   string
	str   string
	num   int
	bulk  string
	array []Value
}

// reader variable of type "pointer to bufio reader"
type Resp struct {
	reader *bufio.Reader
}

// function that takes in a buffer data holder and returns a pointer to Resp struct since the function returns the address of the reader and to use it, we dereference it (mutating orignal reader without copy)
func NewResp(rd io.Reader) *Resp {
	return &Resp{reader: bufio.NewReader(rd)}
}

// readLine takes a reader, reads through it one byte at a time and increment n. if the length of the word is greater than or equal to 2 and the first element of it is \r, then the line is empty or has come to an end ("\r\n" or "<example>\r\n") hence return line, n and nil or else err if any
func (r *Resp) readLine() (line []byte, n int, err error) {
	for {
		b, err := r.reader.ReadByte()
		if err != nil {
			return nil, 0, err
		}
		n += 1
		//consturct line
		line = append(line, b)

		//check if line is empty or has come to an end
		if len(line) >= 2 && line[len(line)-2] == '\r' {
			break
		}
	}
	//return the final constructed line
	return line[:len(line)-2], n, nil
}

// reading integer from the line and returning it as a number
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
		fmt.printf("Unknown type: %v", string(_type))
		return Value{}, nil
	}
}
