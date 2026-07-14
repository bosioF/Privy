package netw

import (
	"strings"
	"fmt"
	"bufio"
	"errors"
	"strconv"
	"os"
	"net"

	"encoding/base64"

	"privy/crypto"
	"privy/types"
	"privy/errs"
)

func ConnInit(Uport int) (net.Listener, error){
	port := ":" + strconv.Itoa(Uport)
	
	listener, err := net.Listen("tcp", port)
	if err != nil {
	 return nil, errors.New(errs.PORT_OCCUPIED)
	}

	return listener, nil
}

func ConnAccept(listener net.Listener) (net.Conn, error){
		conn, err := listener.Accept()
		if err != nil {
			return nil, errors.New(errs.ACCEPT_ERR)
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

func GetConn() (net.Conn, bool, error){
	fmt.Println("You want to host(h) or connect(c)?")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return nil, false, err 
	}

	switch strings.TrimSpace(input){
		case "h":
			conn, err := Listen()
			if err != nil {
				return nil, false, err
			}

			return conn, true, nil
		case "c":
			conn, err := Connect()
			if err != nil {
				return nil, false, err
			}

			return conn, false, nil
		default:
			return nil, false, errors.New("input err")
	}
	
	return nil, false, nil
}

func SendToConn(session *types.PrivySession) error {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Send: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return err
		}

		byteInput := []byte(strings.TrimSpace(input))
		if len(byteInput) == 0 {
			fmt.Print("\033[1A\033[K") //if user pressed enter we clear the line and reprint the prompt to not accumulate "Send" vertically
			continue
		}

		nextChainKey, msgKey, err := crypto.StepRatchet(session.SendingRatchet.ChainKey)
		if err != nil {
			return err
		}
		session.SendingRatchet.ChainKey = nextChainKey

		EncMsg, err := crypto.EncryptMsg(byteInput, msgKey)
		if err != nil {
			return err
		}
		
		b64EncMsg := base64.StdEncoding.EncodeToString(EncMsg)
		finalMsg := fmt.Sprintf("%08x:%s", session.SendingRatchet.SequenceNum, b64EncMsg)
		fmt.Fprintln(session.Conn, finalMsg)

		session.SendingRatchet.SequenceNum++

		fmt.Print("\033[1A\r\033[K") //up one line, go to start of line, remove "Send: "
		fmt.Printf("Sent! -> %s\n", strings.TrimSpace(input)) //reprint "Sent!" w msg
	}
}

func CheckRecvdSeqNum(session *types.PrivySession, seqNum int) ([]byte, error){
	if seqNum == session.ReceivingRatchet.ExpectedSeqNum {
		nextChainKey, msgKey, err := crypto.StepRatchet(session.ReceivingRatchet.ChainKey)
		if err != nil {
			return nil, err
		}

		session.ReceivingRatchet.ChainKey = nextChainKey
		session.ReceivingRatchet.ExpectedSeqNum++

		return msgKey, err
	}

	if seqNum > session.ReceivingRatchet.ExpectedSeqNum {
		if seqNum - session.ReceivingRatchet.ExpectedSeqNum > 100 {
			return nil, errors.New(errs.SEQ_NUM_TOO_LARGE)
		} else {
			var msgKey []byte
			for session.ReceivingRatchet.ExpectedSeqNum < seqNum {
				nextChainKey, msgKey, err := crypto.StepRatchet(session.ReceivingRatchet.ChainKey)
				if err != nil {
					return nil, err
				}

				session.ReceivingRatchet.ChainKey = nextChainKey
				session.ReceivingRatchet.SkippedKeys[session.ReceivingRatchet.ExpectedSeqNum] = msgKey
				session.ReceivingRatchet.ExpectedSeqNum++
			}
			
			nextChainKey, msgKey, err := crypto.StepRatchet(session.ReceivingRatchet.ChainKey)
			if err != nil {
				return nil, err
			}

			session.ReceivingRatchet.ChainKey = nextChainKey
			session.ReceivingRatchet.ExpectedSeqNum++

			return msgKey, nil
		}
	}

	if seqNum < session.ReceivingRatchet.ExpectedSeqNum {
		msgKey, ok := session.ReceivingRatchet.SkippedKeys[seqNum]
		if ok {
			delete(session.ReceivingRatchet.SkippedKeys, seqNum)
			return msgKey, nil
		} else {
			return nil, errors.New(errs.SEQ_NUM_NOT_FOUND_CACHE)
		}
	}

	return nil, errors.New(errs.WTF_ERR)
}

func PrintRecvdLine(session *types.PrivySession, scanner *bufio.Scanner) {
	for scanner.Scan(){
		recvdPayload := strings.Split(scanner.Text(), ":")
		recvdSeqNum, err := strconv.ParseInt(recvdPayload[0], 16, 0)
		if err != nil {
			continue
		}

		msgKey, err := CheckRecvdSeqNum(session, int(recvdSeqNum))
		if err != nil {
			continue
		}

		b64DecMsg, err := base64.StdEncoding.DecodeString(recvdPayload[1])
		if err != nil {
			fmt.Println(errs.B64_ERR)
			continue
		}

		DecMsg, err := crypto.DecryptMsg(b64DecMsg, msgKey)
		if err != nil {
			fmt.Println(errs.DEC_ERR)
			continue
		}
		
		fmt.Print("\r\033[K") //puts the cursors at the start of the line, and then deletes all the lines text
		fmt.Printf("Received -> %s\n", string(DecMsg))
		fmt.Print("Send: ")
	}
}

func HandleConn(session *types.PrivySession, scanner *bufio.Scanner){
	PrintRecvdLine(session, scanner)
}
