package CatTest

import (
	"encoding/json"
	"fmt"
	"main/Model"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

type Credential struct { // This struct is created to support LOGOUT
	Address *net.UDPAddr
	User    Model.User
}

var activePlayer = make(map[string]*Credential)
var ActivePlayer *Credential
var HasLoggedInChannel = make(chan *Credential)
var playerMutex sync.Mutex

// HANDLE SERVICE
func StartServiceServer() {
	addr, err := net.ResolveUDPAddr("udp", ":8080")
	if err != nil {
		fmt.Println("Error resolving address: ", err.Error())
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Error starting UDP server: ", err.Error())
		return
	}
	defer conn.Close()

	fmt.Println("UDP server start on port :8080 ")
	for {
		handleClient(conn)
	}
}

func handleClient(conn *net.UDPConn) {
	buffer := make([]byte, 1024)
	n, clientAddr, err := conn.ReadFromUDP([]byte(buffer))
	if err != nil {
		fmt.Println("Error reading from UDP: ", err)
		return
	}

	message := strings.TrimSpace(string(buffer[:n]))
	// Handle authorization
	if strings.Contains(message, "REGISTER") {
		handleAuthorization(conn, clientAddr, message)
	} else if strings.Contains(message, "LOGOUT") {
		delete(activePlayer, clientAddr.String())
		fmt.Println("<<SERVER>>: Client ", clientAddr, "leaved !")
		// LogActiveUser()
		_, err = conn.WriteToUDP([]byte("Logout successfully !"), clientAddr)
		if err != nil {
			fmt.Println("<<ERROR>>: Fail to send logout confirm to player: ", err.Error())
			return
		}
		return
	}
}

func handleAuthorization(conn *net.UDPConn, clientAddr *net.UDPAddr, message string) {
	token := CreateToken()
	users, err := loadUsers()
	if err != nil {
		fmt.Println("Fail to load user list: ", err.Error())
	}
	splitString := strings.Split(message, ":")
	credential := strings.Split(splitString[1], "-")
	username := credential[0] // get username from sent message from client
	password := credential[1] // get password from sent message from client
	// Handle login logic
	if strings.Contains(message, "Login") {
		// Check if user has exist already
		for index, user := range users {
			if user.Username == username && user.Password == password {
				conn.WriteToUDP([]byte("--> Login successfully !"+strconv.Itoa(token)), clientAddr)
				playerMutex.Lock()
				activePlayer[clientAddr.String()] = &Credential{
					Address: clientAddr,
					User:    user,
				}
				ActivePlayer = &Credential{
					Address: clientAddr,
					User:    user,
				}
				HasLoggedInChannel <- ActivePlayer
				defer playerMutex.Unlock()
				fmt.Println("<<CONFRIM MESSAGE>>: Player", username, "joined the game!")
				// LogActiveUser()
				break
			} else if index == len(users)-1 && user.Username != username && user.Password != password {
				conn.WriteToUDP([]byte("--> Incorrect password / username"), clientAddr)
			}
		}
	}

	// Handle create account logic
	if strings.Contains(message, "CreateAccount") {
		for index, user := range users {
			if user.Username == username {
				conn.WriteToUDP([]byte("<<ERROR>>: Username already exists !"), clientAddr)
				break
			} else if index == len(users)-1 && user.Username != username {
				conn.WriteToUDP([]byte("--> Create account successfully !\nPlease login to play !"), clientAddr)
				newUser := Model.User{
					Username:   username,
					Password:   password,
					OwnPokemon: []Model.Pokemon{},
				}
				users = append(users, newUser)
				err = saveUsers(users)
				if err != nil {
					conn.WriteToUDP([]byte("<<ERROR>>: Fail to save user !"), clientAddr)
					return
				}
			}
		}
	}

}

// Load all users from Users.json
func loadUsers() ([]Model.User, error) {
	file, err := os.Open("Users.json")
	if err != nil {
		if os.IsNotExist(err) {
			return []Model.User{}, nil
		}
		return nil, err
	}
	defer file.Close()

	var users []Model.User
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&users)
	if err != nil && err.Error() != "EOF" {
		return nil, err
	}

	return users, nil
}

// Save user to User.json
func saveUsers(users []Model.User) error {
	file, err := os.Create("Users.json")
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(users)
	if err != nil {
		return err
	}

	return nil
}

// Create token
func CreateToken() int {
	token := rand.Intn(899) + 100
	return token
}

// func LogActiveUser() {
// 	fmt.Println("Active user:")
// 	for index, user := range activePlayers {
// 		fmt.Println("User ", index, user.User.Username, " is in game")
// 	}
// }

func GetCredential() *Credential {
	return ActivePlayer
}
