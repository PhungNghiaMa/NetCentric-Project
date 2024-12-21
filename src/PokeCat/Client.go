package Pokecat

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

// START CLIENT
func StartClient() {
	done := make(chan bool)
	messageChannel := make(chan string)
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("> Do you have have account(Y/N) ?: ")
	Confirmation, _ := reader.ReadString('\n')

	// HANDLE CONNECT TO SERVER
	serverAddr, err := net.ResolveUDPAddr("udp", "localhost:8080") // create UDP enpoint address [IP:port]
	if err != nil {
		log.Fatal("Error resolving address: ", err.Error())
		os.Exit(1)
	}

	conn, err := net.DialUDP("udp", nil, serverAddr) // create a conection to server
	if err != nil {
		log.Fatal("Fail to connect to server: ", err.Error())
		os.Exit(1)
	}
	defer conn.Close()

	// HANDLE LOGIN
	authorizeError := Authorization(serverAddr, conn, strings.TrimSpace(Confirmation))
	if authorizeError != nil {
		fmt.Println("<<ERROR>>: ", authorizeError.Error())
		StartClient()
	}
	go ReceiveMessage(conn, messageChannel, done)
	// Handle logic of authorzation based on server response
	for {
		message := <-messageChannel
		if strings.Contains(message, "Incorrect password / username") { // Handle incorrect credential inputs
			fmt.Println("<<RETRY>>")
			Authorization(serverAddr, conn, "Y")
		} else if strings.Contains(message, "Create account successfully !") { // Handle create successfully new account !
			fmt.Println(message)
			Authorization(serverAddr, conn, "Y")
			return
		} else if strings.Contains(message, "<<ERROR>>: Username already exists !") {
			fmt.Println("<<RETRY>>")
			Authorization(serverAddr, conn, "N") // Handle input clash (same username exist in DB)
			Authorization(serverAddr, conn, "Y")
		} else if strings.Contains(message, "Logout successfully !") {
			<-done
			fmt.Println(message)
			return
		}
		fmt.Print("> ")
		LogoutMsg, _ := reader.ReadString('\n')
		LogoutMsg = strings.TrimSpace(LogoutMsg)
		_, err = conn.Write([]byte(LogoutMsg))
		if err != nil {
			fmt.Println("<<ERROR>>: Fail to sent log out message: ", err.Error())
		}
		if LogoutMsg == "LOGOUT" {
			time.Sleep(600 * time.Millisecond)
			break
		}
	}
}

// AUTHORIZATION FUNCTION
func Authorization(serverAddr *net.UDPAddr, conn *net.UDPConn, answer string) error {
	var err error
	if strings.EqualFold(answer, "Y") {
		Username, Password := Login()
		_, err = conn.Write([]byte("Login-REGISTER:" + Username + "-" + Password))
		if err != nil {
			fmt.Println("<<ERROR>>: Fail to send credential to server: ", err.Error())
			return err
		}
	} else {
		Username, Password := CreateAccount()
		_, err = conn.Write([]byte("CreateAccount-REGISTER:" + Username + "-" + Password))
		if err != nil {
			fmt.Println("Fail to send credential to server: ", err.Error())
			return err
		}
	}
	return nil
}

func Login() (string, string) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("> Enter Username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Print("> Enter Password: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	return username, password
}

func CreateAccount() (string, string) {

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("> Enter Username to create account: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Print("> Enter Password to create account: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	if len(password) < 6 {
		fmt.Println("<<ERROR>>: Password must be at least 6 characters long !")
	}
	return username, password
}

func ReceiveMessage(conn *net.UDPConn, messageChannel chan string, done chan bool) {
	buffer := make([]byte, 1024)
	for {
		select {
		case <-done: // Exit the goroutine if "done" is signaled
			return
		default:
			n, _, err := conn.ReadFromUDP(buffer)
			if err != nil {
				fmt.Println("Error receiving message: ", err)
				return
			}
			message := strings.TrimSpace(string(buffer[:n]))
			fmt.Println("<<SERVER>>: " + message + "\n")
			messageChannel <- message
		}
	}
}
