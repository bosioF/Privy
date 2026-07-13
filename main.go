package main

import (
	"fmt"
	"bufio"

	"privy/crypto"
	"privy/net"
)

func main(){	
	conn, err := net.GetConn()
	if err != nil {
		fmt.Println(err)
		return
	}
	
	scanner := bufio.NewScanner(conn)

	key, err := crypto.Handshake(conn, scanner)
	if err != nil {
		return
	}

	go net.HandleConn(conn, key, scanner)
	
	net.SendToConn(conn, key)
}

