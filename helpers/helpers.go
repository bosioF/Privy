package helpers

import (
	"net"
	"errors"
	"strconv"
	"fmt"

	"privy/netw"
	"privy/errs"
)

func ParseArgs(args []string)(net.Conn, bool, error) {
	if len(args) < 3 {
		return nil, false, errors.New(errs.WRONG_ARGS)
	}

	switch args[0] {
		case "-h":
			if args[1] == "-p" {
				port, err := strconv.Atoi(args[2])
				if err != nil {
					return nil, false, errors.New(errs.CONV_ERR)
				}

				if netw.CheckPort(port, false) {
					listener, err := netw.ConnInit(port)
					if err != nil {
						return nil, false, err
					}

					fmt.Println("Listening...")

					conn, err := netw.ConnAccept(listener)
					if err != nil {
						return nil, false, err
					}
					
					return conn, true, nil
				}

				return nil, false, errors.New(errs.INVALID_PORT)
			}

			return nil, false, errors.New(errs.WRONG_ARGS)

		case "-c":
			if args[1] == "-p" {
				port, err := strconv.Atoi(args[2])
				if err != nil {
					return nil, false, errors.New(errs.CONV_ERR)
				}

				if netw.CheckPort(port, false) {
					portStr := args[2]
					ip := "::1"

					if len(args) >= 5 && args[3] == "-ip" {
						ip = args[4]
						if netw.CheckIp([]byte(ip), false, true) == "" {
							return nil, false, errors.New(errs.GEN_IP_ERR)
						}
					} else if len(args) != 3 {
						return nil, false, errors.New(errs.WRONG_ARGS)
					}
					
					var conn net.Conn
					var err error
					switch netw.CheckIp([]byte(ip), false, false) {
						case "v4":
							targetAddr := ip + ":" + portStr
				
							fmt.Println("Connecting to", targetAddr, "...")
							conn, err = net.Dial("tcp", targetAddr)
							if err != nil {
								return nil, false, errors.New(errs.DIAL_ERR)
							}

						case "v6":
							targetAddr := "[" + ip + "]" + ":" + portStr

							fmt.Println("Connecting to", targetAddr, "...")
							conn, err = net.Dial("tcp6", targetAddr)
							if err != nil {
								return nil, false, errors.New(errs.DIAL_ERR)
							}

						default:
							return nil, false, errors.New(errs.GEN_IP_ERR)
					}

					return conn, false, nil
				}

				return nil, false, errors.New(errs.INVALID_PORT)
			}

			return nil, false, errors.New(errs.WRONG_ARGS)

				
		default:
			return nil, false, errors.New(errs.WRONG_ARGS)
	}

	return nil, false, errors.New(errs.WRONG_ARGS)
}
