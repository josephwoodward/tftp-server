package tftp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"strings"
)

const (
	DatagramSize = 516 // Maximum supported datagram size
	BlockSize    = DatagramSize
)

type OpCode uint16

//opcode  operation
//1     Read request (RRQ)
//2     Write request (WRQ)
//3     Data (DATA)
//4     Acknowledgment (ACK)
//5     Error (ERROR)
const (
	OpRRQ = iota + 1
	_     // This will be read only for the moment
	OpData
	OpAck
	OpErr
)

type ErrCode uint16

const (
	ErrUnknown ErrCode = iota
	ErrNotFound
	ErrAccessViolation
	ErrDiskFull
	ErrIllegalOp
	ErrUnknownID
	ErrFileExists
	ErrNoUser
)

type ReadReq struct {
	Filename string
	Mode     string
}

func (q *ReadReq) MarshalBinary([]byte, error) ([]byte, error) {
	mode := "octet"
	if q.Mode != "" {
		mode = q.Mode
	}

	//2 bytes     string    1 byte     string   1 byte
	//------------------------------------------------
	//| Opcode |  Filename  |   0  |    Mode    |   0  |
	//------------------------------------------------
	// capacity: operation code + filename + 0 byte + mode + 0 byte
	// https://datatracker.ietf.org/doc/html/rfc1350#section-5
	cap := 2 + 2 + len(q.Filename) + 1 + len(q.Mode) + 1

	b := new(bytes.Buffer)
	b.Grow(cap)

	// Write Opcode
	if err := binary.Write(b, binary.BigEndian, OpRRQ); err != nil {
		return nil, err
	}

	// Write Filename
	if _, err := b.WriteString(q.Filename); err != nil {
		return nil, err
	}

	// Write null byte
	if err := b.WriteByte(0); err != nil {
		return nil, err
	}

	// Write Mode
	if _, err := b.WriteString(mode); err != nil {
		return nil, err
	}

	// Write another null byte
	if err := b.WriteByte(0); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (q *ReadReq) UnmarshalBinary(p []byte) error {
	r := bytes.NewBuffer(p)

	var code OpCode
	var err error

	// Read the OpCode
	if err = binary.Read(r, binary.BigEndian, &code); err != nil {
		return err
	}

	if code != OpRRQ {
		return errors.New("invalid RRQ")
	}

	// Read the filename
	q.Filename, err = r.ReadString(0)
	if err != nil {
		return errors.New("invalid RRQ")
	}

	// Remove the null byte
	q.Filename = strings.TrimRight(q.Filename, "\x00")
	if len(q.Filename) == 0 {
		return errors.New("invalid RRQ")
	}

	// Get the mode
	q.Mode, err = r.ReadString(0)
	if err != nil {
		return errors.New("invalid RRQ")
	}

	actual := strings.ToLower(q.Mode)
	if actual != "octet" {
		return errors.New("only binary transfers supported at the moment")
	}

	return nil
}
