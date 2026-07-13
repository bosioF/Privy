package types

import (
	"net"
)

type SendingRatchet struct {
	ChainKey []byte
	SequenceNum int
} 

type ReceivingRatchet struct {
	ChainKey []byte
	ExpectedSeqNum int
	SkippedKeys map[int][]byte
}

type PrivySession struct {
	Conn net.Conn
	SendingRatchet *SendingRatchet
	ReceivingRatchet *ReceivingRatchet
}
