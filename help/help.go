package help

import (
	"fmt"
)

func Menu(){
	fmt.Println("Print this menu: ./privy --help")
	fmt.Println("")
	fmt.Println("Host on PORT: ./privy -h -p <PORT>")
	fmt.Println("Connect to PORT on localhost: ./privy -c -p <PORT>")
	fmt.Println("Connect to PORT on IPv4 or IPv6: ./privy -c -p <PORT> -ip <IP>")
}
