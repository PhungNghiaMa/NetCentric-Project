package main

import (
	CatTest "main/PokeCat"
	"main/Pokedex"
)

func main() {
	Pokedex.CrawlDriver()
	CatTest.StartGame()
	// Get boolean variable "done" from Pokedex
	// If "done" variable is true in Pokedex then call PokeCat.Start()
}
