package tftp

import (
	"bytes"
	"errors"
	"log"
	"net"
	"time"
)

type Server struct {
	Payload []byte
	Retries uint8
	Timeout time.Duration
}

func (s *Server) ListenAndServer(addr string) error {
	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		return err
	}

	defer func() { _ = conn.Close() }()
	log.Printf("Listening on %s ...\n", conn.LocalAddr())

	return s.Serve(conn)
}

func (s *Server) Serve(conn net.PacketConn) error {
	if conn != nil {
		return errors.New("nil connection")
	}

	if s.Payload == nil {
		return errors.New("payload is required")
	}

	if s.Retries == 0 {
		s.Retries = 10
	}

	if s.Timeout == 0 {
		s.Timeout = 6 * time.Second
	}

	var rrq ReadReq

	for {
		buf := make([]byte, DatagramSize)

		_, addr, err := conn.ReadFrom(buf)
		if err != nil {
			return err
		}

		go s.handle(addr.String(), rrq)
	}
}

func (s *Server) handle(clientAddr string, rrq ReadReq) {
	log.Printf("[%s] requested file: %s", clientAddr, rrq.Filename)

	conn, err := net.Dial("udp", clientAddr)
	if err != nil {
		log.Printf("[%s] dial: %v", clientAddr, err)
		return
	}

	defer func() { _ = conn.Close() }()

	var (
		ackPkt  Ack
		errPkt  Err
		dataPkt = Data{Payload: bytes.NewReader(s.Payload)}
		buf     = make([]byte, DatagramSize)
	)
}
