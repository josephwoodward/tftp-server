package tftp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
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

// ReadReq acts as the initial read request packet (RRQ) informing the server which file it would like to read
//2 bytes     string    1 byte     string   1 byte
//------------------------------------------------
//| Opcode |  Filename  |   0  |    Mode    |   0  |
//------------------------------------------------
type ReadReq struct {
	Filename string
	Mode     string
}

// MarshalBinary won't work yet as we're only focusing on downloading
func (q *ReadReq) MarshalBinary() ([]byte, error) {
	mode := "octet"
	if q.Mode != "" {
		mode = q.Mode
	}

	// capacity: operation code + filename + 0 byte + mode + 0 byte
	// https://datatracker.ietf.org/doc/html/rfc1350#section-5
	capacity := 2 + 2 + len(q.Filename) + 1 + len(q.Mode) + 1

	b := new(bytes.Buffer)
	b.Grow(capacity)

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

	// Read the filename including the packet null byte delimiter
	q.Filename, err = r.ReadString(0)
	if err != nil {
		return errors.New("invalid RRQ")
	}

	// Remove the null byte from the end of the filename
	q.Filename = strings.TrimRight(q.Filename, "\x00")
	if len(q.Filename) == 0 {
		return errors.New("invalid RRQ")
	}

	// Get the mode including null byte delimiter again
	q.Mode, err = r.ReadString(0)
	if err != nil {
		return errors.New("invalid RRQ")
	}

	// Remove null byte delimiter again
	q.Mode = strings.TrimRight(q.Mode, "\x00")
	if len(q.Mode) == 0 {
		return errors.New("invalid RRQ")
	}

	actual := strings.ToLower(q.Mode)
	if actual != "octet" {
		return errors.New("only binary transfers supported at the moment")
	}

	return nil
}

// Data acts as the data packet that will transfer the files payload
// 2 bytes     2 bytes      n bytes
// ----------------------------------
// | Opcode |   Block #  |   Data     |
// ----------------------------------
type Data struct {
	// Block enables UDP reliability by incrementing on each packet sent,
	// the client discriminate between new packets and duplicates, sending an ack including the block number to
	// confirm delivery
	Block   uint16
	Payload io.Reader
}

func (d *Data) MarshalBinary() ([]byte, error) {
	b := new(bytes.Buffer)
	b.Grow(DatagramSize)

	d.Block++

	err := binary.Write(b, binary.BigEndian, d.Block) // write block number to packet
	if err != nil {
		return nil, err
	}

	// Every packet will be BlockSize (516 bytes) expect for the last one, which is how the client knows
	// it's reached the end of the stream
	_, err = io.CopyN(b, d.Payload, BlockSize)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return b.Bytes(), nil
}

func (d *Data) UnmarshalBinary(p []byte) error {
	// Sanity check the payload data
	if l := len(p); l < 4 || l > DatagramSize {
		return errors.New("invalid DATA")
	}

	var opcode any
	// Read opcode from packet
	err := binary.Read(bytes.NewReader(p[:2]), binary.BigEndian, &opcode)
	if err != nil || opcode != OpData {
		return errors.New("invalid DATA")
	}

	// Read block number
	err = binary.Read(bytes.NewReader(p[2:4]), binary.BigEndian, &d.Block)
	if err != nil {
		return errors.New("invalid DATA")
	}

	// Read byte slice to get the end to get data
	d.Payload = bytes.NewBuffer(p[4:])

	return nil
}

// Ack responds to the server with a block number to inform the server
// which packet it just received
// 2 bytes     2 bytes
// ---------------------
// | Opcode |   Block #  |
// ---------------------
type Ack uint16

func (a *Ack) MarshalBinary() ([]byte, error) {
	capacity := 2 + 2 // operation code + block number

	b := new(bytes.Buffer)
	b.Grow(capacity)

	err := binary.Write(b, binary.BigEndian, OpAck) // Write ack op code to buffer
	if err != nil {
		return nil, err
	}

	err = binary.Write(b, binary.BigEndian, &a) // Now write block number
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (a *Ack) UnmarshalBinary(p []byte) error {
	var code OpCode

	r := bytes.NewReader(p)

	if err := binary.Read(r, binary.BigEndian, &code); err != nil {
		return err
	}

	if code != OpAck {
		return errors.New("invalid ACK")
	}

	return binary.Read(r, binary.BigEndian, a)
}

// Err packet
// 2 bytes     2 bytes       string    1 byte
// -----------------------------------------
// | Opcode |  ErrorCode |   ErrMsg   |   0  |
// -----------------------------------------
type Err struct {
	Error   ErrCode
	Message string
}

func (e Err) MarshalBinary() ([]byte, error) {
	capacity := 2 + 2 + len(e.Message) + 1

	b := new(bytes.Buffer)
	b.Grow(capacity)

	err := binary.Write(b, binary.BigEndian, OpErr) // Write OpErr op code to buffer
	if err != nil {
		return nil, err
	}

	// Now write error code
	if err = binary.Write(b, binary.BigEndian, e.Error); err != nil {
		return nil, err
	}

	_, err = b.WriteString(e.Message)
	if err != nil {
		return nil, err
	}

	if err = b.WriteByte(0); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (e Err) UnmarshalBinary(p []byte) error {
	r := bytes.NewBuffer(p)

	var code OpCode

	if err := binary.Read(r, binary.BigEndian, &code); err != nil { // read op code
		return err
	}

	if code != OpErr {
		return errors.New("invalid ERROR")
	}

	if err := binary.Read(r, binary.BigEndian, &e.Error); err != nil {
		return err
	}

	var err error
	e.Message, err = r.ReadString(0)
	e.Message = strings.TrimRight(e.Message, "\x00") // remove the 0-byte

	return err
}
