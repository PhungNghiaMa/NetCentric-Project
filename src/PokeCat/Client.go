package Pokecat

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func StartPokeCat() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Please enter your usename: ")
	Username, _ := reader.ReadString('\n')
	Username = strings.TrimSpace(Username)

	serverAddr, err := net.ResolveUDPAddr()

}
