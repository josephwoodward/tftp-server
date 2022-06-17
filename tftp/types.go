package tftp

const (
	DatagramSize = 516 // Maximum supported datagram size
	BlockSize    = DatagramSize
)

type OpCode uint16

const (
	OpRRQ = iota + 1
	_
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
