package main

import (
	"fmt"
	"bufio"
	"os"
	"net"

	"privy/crypto"
	"privy/netw"
	"privy/types"
	"privy/errs"
	"privy/parsing"
	"privy/help"
)

func main(){
	var conn net.Conn
	var isHost bool
	var err error

	argCount := len(os.Args) - 1
	if argCount == 0 {
		conn, isHost, err = netw.GetConn()
		if err != nil {
			fmt.Println(err)
			return
		}	
	} else if argCount == 1 {
		help.Menu()
		return
	} else if argCount == 3 || argCount == 5 {
		conn, isHost, err = parsing.ParseArgs(os.Args[1:])
		if err != nil {
			fmt.Println(err)
			return
		}
	} else {
		fmt.Println(errs.WRONG_ARGS)
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

	go netw.HandleConn(session, scanner)
	
	netw.SendToConn(session)
}

