package CatTest

import (
	"encoding/json"
	"fmt"
	"main/Model"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/nsf/termbox-go"
)

type Coordinate struct {
	X int
	Y int
}

type Pokemon struct {
	Pokemon Model.Pokemon
	Coords  Coordinate
}

type World struct {
	Size    int
	Cells   map[string]*Pokemon // Map to track Pokemon in each cell
	Players map[string]*Player  // Map to track all player join game
	Mutex   sync.Mutex
}

type Player struct {
	Credential Credential
	PlayerPos  Coordinate
}

var pokedex []Model.Pokemon
var world World
var playerPosition Coordinate

const (
	DefaultWorldSize = 20
	PokemonWave      = 2
)

var spawnMutex sync.Mutex

func InitializePokedex() []Model.Pokemon {
	filename := "./POKEMONS.json"

	// Read pokemon file
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("Failed to read POKEMONS.json: ", err.Error())
		return nil
	}
	// Write pokemon from JSON to slice
	if err = json.Unmarshal(data, &pokedex); err != nil {
		fmt.Println("Fail to parse POKEMON data to slice: ", err.Error())
		return nil
	}
	return pokedex
}

func InitializeWorld(size int) *World {
	world := &World{
		Size:    size,
		Cells:   make(map[string]*Pokemon),
		Players: make(map[string]*Player), // Initialize map for Players
		Mutex:   sync.Mutex{},
	}
	return world
}

func InitializePlayer(PlayerCredential *Credential, world *World) *Player {
	player := Player{
		Credential: *PlayerCredential,
		PlayerPos: Coordinate{
			X: rand.Intn(world.Size),
			Y: rand.Intn(world.Size),
		},
	}
	playerPosition = Coordinate{X: player.PlayerPos.X, Y: player.PlayerPos.Y}
	return &player
}

func SpawnPokemon(world *World, DEX []Model.Pokemon) *World {
	// spawnMutex.Lock()
	// defer spawnMutex.Unlock()

	world.Mutex.Lock()
	defer world.Mutex.Unlock()
	ReleasePokemon := 5
	for i := 0; i < ReleasePokemon; i++ {
		pokemon := Pokemon{
			Pokemon: pokedex[rand.Intn(len(pokedex))],
			Coords: Coordinate{
				X: rand.Intn(world.Size),
				Y: rand.Intn(world.Size),
			},
		}
		key := fmt.Sprintf("%d,%d", pokemon.Coords.X, pokemon.Coords.Y)
		world.Cells[key] = &pokemon
	}
	return world
}

func CapturePokemon(newworld *World, player *Player) {
	key := fmt.Sprintf("%d,%d", player.PlayerPos.X, player.PlayerPos.Y)
	spawnMutex.Lock()
	defer spawnMutex.Unlock()
	world.Mutex.Lock()
	defer world.Mutex.Unlock()

	if pokemon, exists := newworld.Cells[key]; exists {
		delete(newworld.Cells, key)
		err := UpdatePlayerInfor(player, pokemon)
		if err != nil {
			fmt.Println("Fail to update player information: ", err.Error())
		} else {
			fmt.Printf("Captured Pokemon: %s!\n", pokemon.Pokemon.Name)
		}
	}
}

func UpdatePlayerInfor(player *Player, pokemon *Pokemon) error {
	filename := "./Users.json"
	var UserSlice []Model.User

	// Read the existing Users.json file
	file, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("Cannot read Users.json: ", err.Error())
		return err
	}

	// Unmarshal the file content into UserSlice
	if err = json.Unmarshal(file, &UserSlice); err != nil {
		fmt.Println("Fail to parse users from Users.json: ", err.Error())
		return err
	}

	// Iterate through the UserSlice to find the player by username
	for i := 0; i < len(UserSlice); i++ {
		users := &UserSlice[i]
		if player.Credential.User.Username == users.Username {
			// Check if the player already owns this Pokemon
			pokemonExists := false
			for _, ownPoke := range users.OwnPokemon {
				if ownPoke.Name == pokemon.Pokemon.Name {
					pokemonExists = true
					break
				}
			}

			// If the player doesn't own the Pokemon, add it
			if !pokemonExists {
				users.OwnPokemon = append(users.OwnPokemon, pokemon.Pokemon)
				fmt.Printf("Pokemon %s added to %s's collection.\n", pokemon.Pokemon.Name, player.Credential.User.Username)
			} else {
				fmt.Printf("Player %s already owns this Pokemon %s.\n", player.Credential.User.Username, pokemon.Pokemon.Name)
			}
			break
		}
	}

	// Re-serialize the updated UserSlice and write it back to Users.json
	updatedData, err := json.MarshalIndent(UserSlice, "", "  ")
	if err != nil {
		fmt.Println("Failed to marshal updated user data: ", err.Error())
		return err
	}

	// Write the updated data back to Users.json
	err = os.WriteFile(filename, updatedData, 0644)
	if err != nil {
		fmt.Println("Failed to write to Users.json: ", err.Error())
		return err
	}

	return nil
}

// RenderWorld function (with improvements to fix flashing issue)
func RenderWorld(players []*Player, world *World) {
	// Only clear the screen once at the start of each frame
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	// Lock the world to avoid concurrent updates during rendering
	world.Mutex.Lock()
	defer world.Mutex.Unlock()

	// Render the world grid
	for y := 0; y < world.Size; y++ {
		for x := 0; x < world.Size; x++ {
			key := fmt.Sprintf("%d,%d", x, y)
			if _, exists := world.Cells[key]; exists {
				// Render pokemon in the world with red "O" character
				termbox.SetCell(x, y, 'O', termbox.ColorRed, termbox.ColorDefault)
			} else {
				// Render empty spaces with white "."
				termbox.SetCell(x, y, '.', termbox.ColorWhite, termbox.ColorDefault)
			}
		}
	}

	// Render each player with a unique color
	for i, player := range players {
		playerColor := []termbox.Attribute{termbox.ColorRed, termbox.ColorGreen}[i]
		if i == 0 {
			// Player 1 with red color
			playerColor = termbox.ColorGreen
		}
		// Render player position as 'P'
		termbox.SetCell(player.PlayerPos.X, player.PlayerPos.Y, 'P', playerColor, termbox.ColorDefault)
	}

	// Flush the buffer to the terminal
	termbox.Flush()
}

// Handle player movement
// Handle player movement
func handlePlayerInput(player *Player, world *World, stopChan chan struct{}, playerIndex int) {
	for {
		select {
		case <-stopChan:
			return
		default:
			event := termbox.PollEvent()
			if event.Type == termbox.EventKey {
				world.Mutex.Lock()

				// Distinguish between Player 1 and Player 2 based on playerIndex
				switch playerIndex {
				case 0: // Player 1 controls using arrow keys
					switch event.Key {
					case termbox.KeyArrowUp:
						if player.PlayerPos.Y > 0 {
							player.PlayerPos.Y--
						}
					case termbox.KeyArrowDown:
						if player.PlayerPos.Y < world.Size-1 {
							player.PlayerPos.Y++
						}
					case termbox.KeyArrowLeft:
						if player.PlayerPos.X > 0 {
							player.PlayerPos.X--
						}
					case termbox.KeyArrowRight:
						if player.PlayerPos.X < world.Size-1 {
							player.PlayerPos.X++
						}
					case termbox.KeyEsc:
						close(stopChan)
						return
					}
				case 1: // Player 2 controls using WASD keys
					switch event.Ch {
					case 'w', 'W':
						if player.PlayerPos.Y > 0 {
							player.PlayerPos.Y--
						}
					case 's', 'S':
						if player.PlayerPos.Y < world.Size-1 {
							player.PlayerPos.Y++
						}
					case 'a', 'A':
						if player.PlayerPos.X > 0 {
							player.PlayerPos.X--
						}
					case 'd', 'D':
						if player.PlayerPos.X < world.Size-1 {
							player.PlayerPos.X++
						}
					case 'q': // Quit player 2
						close(stopChan)
						return
					}
				}

				// Check for capturing a pokemon
				CapturePokemon(world, player)
				world.Mutex.Unlock()
			}
		}
	}
}

func StartGameService(world *World, players []*Player) {
	if err := termbox.Init(); err != nil {
		fmt.Println("Failed to initialize termbox:", err)
		return
	}
	defer termbox.Close()

	stopChans := make([]chan struct{}, len(players))
	for i := range players {
		stopChans[i] = make(chan struct{})
		go handlePlayerInput(players[i], world, stopChans[i], i)
	}

	// Game loop with smoother frame rate to reduce flicker
	for {
		RenderWorld(players, world)

		// Sleep for 100 milliseconds (adjust as necessary for smoother experience)
		time.Sleep(200 * time.Millisecond)
	}
}
