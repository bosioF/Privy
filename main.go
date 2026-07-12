package main

import (
	"fmt"
	"net"
	"bufio"
	"errors"
	"strconv"
	"os"
	"strings"
	"privy/crypto"
	"encoding/base64"
)

func ConnInit(Uport int) (net.Listener, error){
	port := ":" + strconv.Itoa(Uport)
	
	listener, err := net.Listen("tcp", port)
	if err != nil {
	 return nil, errors.New("err while trying to listen, port occupied?")
	}

	return listener, nil
}

func ConnAccept(listener net.Listener) (net.Conn, error){
		conn, err := listener.Accept()
		if err != nil {
			return nil, errors.New("err while accepting connection")
		}
		
		fmt.Println("Connection successful!", conn.RemoteAddr().String())

		return conn, nil
}

func CheckPort(port int, flag int)(int){
	if port < 1024 || port > 65535 {
		if flag == 1 {
			return 0
		}

		fmt.Println("Port not valid")
		return 0
	}

	return 1
}

func Listen() (net.Conn, error){
	var port int = 0
	var flag int = 1
	
	for CheckPort(port, flag) == 0 {
		flag = 0
		fmt.Print("On what port do you want to listen? (1024-65535) ")
		_, err := fmt.Scanf("%d", &port)
		if err != nil {
			return nil, err
		}
	}
	
	listener, err := ConnInit(port)
	if err != nil {
		return nil, err
	}

	fmt.Println("Listening")
	
	conn, err := ConnAccept(listener)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func Connect() (net.Conn, error){
	reader := bufio.NewReader(os.Stdin)
	
	var port int = 0
	var flag int = 1

	for CheckPort(port, flag) == 0 {
		flag = 0
		fmt.Print("On what port do you want to connect? (1024-65535) ")
		portInput, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		port, _ = strconv.Atoi(strings.TrimSpace(portInput))
	}

	fmt.Print("What is the IP? (Press Enter for localhost): ")
	ipInput, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	ip := strings.TrimSpace(ipInput)

	if ip == "" {
		ip = "127.0.0.1"
	}

	targetAddr := ip + ":" + strconv.Itoa(port)

	fmt.Println("Connecting to", targetAddr, "...")
	conn, err := net.Dial("tcp", targetAddr)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func GetConn() (net.Conn, error){
	fmt.Println("You want to host(h) or connect(c)?")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return nil, err 
	}

	switch strings.TrimSpace(input){
		case "h":
			conn, err := Listen()
			if err != nil {
				return nil, err
			}

			return conn, nil
		case "c":
			conn, err := Connect()
			if err != nil {
				return nil, err
			}

			return conn, nil
		default:
			return nil, errors.New("input err")
	}
	
	return nil, nil
}

func SendToConn(conn net.Conn, key []byte) error {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Send: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return err
		}

		byteInput := []byte(strings.TrimSpace(input))
		if len(byteInput) == 0 {
			continue
		}

		EncMsg, err := crypto.EncryptMsg(byteInput, key)
		if err != nil {
			return err
		}

		b64EncMsg := base64.StdEncoding.EncodeToString(EncMsg)
		fmt.Fprintln(conn, b64EncMsg)
	}
}

func PrintRecvdLine(conn net.Conn, key []byte, scanner *bufio.Scanner){
	for scanner.Scan(){
		b64DecMsg, err := base64.StdEncoding.DecodeString(scanner.Text())
		if err != nil {
			fmt.Println("\rerror while dec b64")
			continue
		}

		DecMsg, err := crypto.DecryptMsg(b64DecMsg, key)
		if err != nil {
			fmt.Println("\rerror while dec")
			continue
		}
		
		fmt.Print("\r\033[K") //puts the cursors at the start of the line, and then deletes all the lines text
		fmt.Println("Received: ", string(DecMsg))
		fmt.Print("Send: ")
	}
}

func HandleConn(conn net.Conn, key []byte, scanner *bufio.Scanner){
	PrintRecvdLine(conn, key, scanner)
}

func main(){	
	conn, err := GetConn()
	if err != nil {
		fmt.Println(err)
		return
	}
	
	scanner := bufio.NewScanner(conn)

	key, err := crypto.Handshake(conn, scanner)
	if err != nil {
		return
	}

	go HandleConn(conn, key, scanner)
	
	SendToConn(conn, key)
}

