package main

import (
	"fmt"
	"bufio"

	"privy/crypto"
	"privy/net"
	"privy/types"
)

func main(){	
	conn, isHost, err := net.GetConn()
	if err != nil {
		fmt.Println(err)
		return
	}
	
	scanner := bufio.NewScanner(conn)

	key, err := crypto.Handshake(conn, scanner)
	if err != nil {
		return
	}

	sendRatchet, recvRatchet, err := crypto.InitRatchet(key, isHost)
	if err != nil {
		return
	}

	session := &types.PrivySession{
		conn,
		sendRatchet,
		recvRatchet,
	}

	go net.HandleConn(session, scanner)
	
	net.SendToConn(session)
}

