package CatTest

var LoginChannel = make(chan bool)

// THIS IS USED TO CALL LOG IN SERVICE AND GAME SERVICE OF POKECAT
func StartGame() {
	DEX := InitializePokedex()
	world := InitializeWorld(20)
	newworld := SpawnPokemon(world, DEX)
	var players []*Player
	// Start Login Service
	go StartServiceServer()
	for {
		PlayerCredential := <-HasLoggedInChannel
		// Initialize the player and add them to the shared world
		go func(credential *Credential) {
			player := InitializePlayer(credential, newworld)
			players = append(players, player)
			StartGameService(newworld, players)
		}(PlayerCredential)
	}

}
