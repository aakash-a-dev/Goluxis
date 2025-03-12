package resp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

const (
	// RESP type bytes
	SimpleString = '+'
	Error        = '-'
	Integer      = ':'
	BulkString   = '$'
	Array        = '*'
)

var (
	ErrInvalidFormat = errors.New("invalid RESP format")
	CRLF             = "\r\n"
)

// Reader implements RESP protocol reading
type Reader struct {
	*bufio.Reader
}

// NewReader creates a new RESP reader
func NewReader(rd io.Reader) *Reader {
	return &Reader{bufio.NewReader(rd)}
}

// ReadObject reads a RESP object from the reader
func (r *Reader) ReadObject() (interface{}, error) {
	typ, err := r.ReadByte()
	if err != nil {
		return nil, err
	}

	switch typ {
	case SimpleString:
		return r.readLine()
	case Error:
		msg, err := r.readLine()
		if err != nil {
			return nil, err
		}
		return errors.New(msg), nil
	case Integer:
		return r.readInteger()
	case BulkString:
		return r.readBulkString()
	case Array:
		return r.readArray()
	default:
		return nil, fmt.Errorf("unknown RESP type byte: %c", typ)
	}
}

// readLine reads a line terminated by CRLF
func (r *Reader) readLine() (string, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	if len(line) < 2 || line[len(line)-2] != '\r' {
		return "", ErrInvalidFormat
	}
	return line[:len(line)-2], nil
}

// readInteger reads a RESP integer
func (r *Reader) readInteger() (int64, error) {
	line, err := r.readLine()
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(line, 10, 64)
}

// readBulkString reads a RESP bulk string
func (r *Reader) readBulkString() (string, error) {
	length, err := r.readInteger()
	if err != nil {
		return "", err
	}

	if length == -1 {
		return "", nil // null bulk string
	}

	buf := make([]byte, length+2) // +2 for CRLF
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return "", err
	}

	if !strings.HasSuffix(string(buf), CRLF) {
		return "", ErrInvalidFormat
	}

	return string(buf[:length]), nil
}

// readArray reads a RESP array
func (r *Reader) readArray() ([]interface{}, error) {
	length, err := r.readInteger()
	if err != nil {
		return nil, err
	}

	if length == -1 {
		return nil, nil // null array
	}

	array := make([]interface{}, length)
	for i := range array {
		array[i], err = r.ReadObject()
		if err != nil {
			return nil, err
		}
	}

	return array, nil
}

// Writer implements RESP protocol writing
type Writer struct {
	*bufio.Writer
}

// NewWriter creates a new RESP writer
func NewWriter(w io.Writer) *Writer {
	return &Writer{bufio.NewWriter(w)}
}

// WriteSimpleString writes a RESP simple string
func (w *Writer) WriteSimpleString(s string) error {
	return w.writeString(fmt.Sprintf("%c%s%s", SimpleString, s, CRLF))
}

// WriteError writes a RESP error
func (w *Writer) WriteError(err error) error {
	return w.writeString(fmt.Sprintf("%c%s%s", Error, err.Error(), CRLF))
}

// WriteInteger writes a RESP integer
func (w *Writer) WriteInteger(i int64) error {
	return w.writeString(fmt.Sprintf("%c%d%s", Integer, i, CRLF))
}

// WriteBulkString writes a RESP bulk string
func (w *Writer) WriteBulkString(s string) error {
	if s == "" {
		return w.writeString(fmt.Sprintf("%c-1%s", BulkString, CRLF))
	}
	return w.writeString(fmt.Sprintf("%c%d%s%s%s", BulkString, len(s), CRLF, s, CRLF))
}

// WriteArray writes a RESP array header
func (w *Writer) WriteArray(length int) error {
	if length < 0 {
		return w.writeString(fmt.Sprintf("%c-1%s", Array, CRLF))
	}
	return w.writeString(fmt.Sprintf("%c%d%s", Array, length, CRLF))
}

// writeString writes a string and flushes the writer
func (w *Writer) writeString(s string) error {
	_, err := w.WriteString(s)
	if err != nil {
		return err
	}
	return w.Flush()
}
