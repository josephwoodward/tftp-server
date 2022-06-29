# Trivial File Transfer Protocol Server (TFTP) implementation in Go

TFTP is a (trivial) File Transfer Protocol that ensures reliable data transfer over UDP. It does this by using a subset of mechanics that make TCP reliable.  
The full RFC can be found below, with the packet structure highlighted in this read me.

### Usage

```shell
$ tftp -e 127.0.0.1
get payload.jpeg
```

https://datatracker.ietf.org/doc/html/rfc1350

### Packet structure

#### Initial Read Request

ReadReq acts as the initial read request packet (RRQ) informing the server which file it would like to read
```
2 bytes     string    1 byte     string   1 byte
------------------------------------------------
| Opcode |  Filename  |   0  |    Mode    |   0  |
------------------------------------------------
```

### Data Packet

The data packet contains an opcode, the block number (a monotonically incrementing number used by the server to ensure reliable transfer by waiting for the client to ack the packet by sending the block number back to the server)

```
2 bytes     2 bytes      n bytes
----------------------------------
| Opcode |   Block #  |   Data     |
----------------------------------
```

### ACK Packet

 Ack responds to the server with a block number to inform the server which packet it just received
```
2 bytes     2 bytes
---------------------
| Opcode |   Block #  |
---------------------
```
