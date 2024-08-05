package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

const (
	STRING  = '+'
	ERROR   = '-'
	INTEGER = ':'
	BULK    = '$'
	ARRAY   = '*'
)

type Value struct {
	typ   string
	str   string
	bulk  string
	array []Value
	num   int
}

type Resp struct {
	reader *bufio.Reader
}

func NewResp(rd io.Reader) *Resp {
	return &Resp{reader: bufio.NewReader(rd)}
}

func (r *Resp) readLine() (line []byte, n int, err error) {
	for {
		b, err := r.reader.ReadByte()
		if err != nil {
			return nil, 0, err
		}

		n += 1
		line = append(line, b)

		// check if we have reached end of line
		if len(line) >= 2 && line[len(line)-2] == '\r' {
			break
		}

	}

	// return line without last 2 bytes /r/n
	return line[:len(line)-2], n, nil
}

func (r *Resp) readInteger() (x int, n int, err error) {
	line, n, err := r.readLine()
	if err != nil {
		return 0, 0, err
	}
	i64, err := strconv.ParseInt(string(line), 10, 64)
	if err != nil {
		return 0, 0, err
	}
	return int(i64), n, nil
}

func (r *Resp) Read() (Value, error) {
	_type, err := r.reader.ReadByte()
	if err != nil {
		return Value{}, err
	}

	switch _type {
	case STRING:
		return r.parseString()
	case INTEGER:
		return r.parseInteger()
	case ERROR:
		return r.parseError()
	case ARRAY:
		return r.parseArray()

	case BULK:
		return r.parseBulk()

	default:
		fmt.Printf("Unknown type %v: ", string(_type))
		return Value{}, nil
	}
}

func (r *Resp) parseString() (Value, error) {
	v := Value{}
	v.typ = "string"

	line, _, err := r.readLine()
	if err != nil {
		return v, err
	}

	v.str = string(line)

	return v, nil
}

func (r *Resp) parseInteger() (Value, error) {
	v := Value{}
	v.typ = "int"

	i, _, err := r.readInteger()
	if err != nil {
		return Value{}, err
	}

	v.num = i

	return v, nil
}

func (r *Resp) parseError() (Value, error) {
	v := Value{}
	v.typ = "error"

	line, _, err := r.readLine()
	if err != nil {
		return Value{}, err
	}
	v.str = string(line)

	return v, err
}

func (r *Resp) parseArray() (Value, error) {
	v := Value{}
	v.typ = "array"

	// read length of array
	len, _, err := r.readInteger()
	if err != nil {
		return v, err
	}

	v.array = make([]Value, 0)

	for i := 0; i < len; i++ {
		val, err := r.Read()
		if err != nil {
			return Value{}, err
		}

		// append read value
		v.array = append(v.array, val)
	}
	return v, nil
}

func (r *Resp) parseBulk() (Value, error) {
	v := Value{}
	v.typ = "bulk"

	// read length of string
	len, _, err := r.readInteger()
	if err != nil {
		return v, err
	}

	// check for null bulk string
	if len == -1 {
		v.typ = "null"
		return v, nil
	}

	bulk := make([]byte, len)
	r.reader.Read([]byte(bulk))
	v.bulk = string(bulk)

	// read trailing CRLF
	r.readLine()

	return v, nil
}

// Marshal value to bytes
func (v Value) Marshal() []byte {
	switch v.typ {
	case "array":
		return v.marshalArray()
	case "bulk":
		return v.marshalBulk()
	case "string":
		return v.marshalString()
	case "error":
		return v.marshalError()
	case "null":
		return v.marshalError()
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

func (v Value) marshalBulk() []byte {
	var bytes []byte
	bytes = append(bytes, BULK)
	bytes = append(bytes, strconv.Itoa(len(v.bulk))...)
	bytes = append(bytes, '\r', '\n')
	bytes = append(bytes, v.bulk...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

func (v Value) marshalError() []byte {
	var bytes []byte
	bytes = append(bytes, ERROR)
	bytes = append(bytes, v.str...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

func (v Value) marshalNull() []byte {
	return []byte("$-1\r\n")
}

// Writer
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
