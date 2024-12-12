package main

import (
	"main/Pokedex"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	Pokedex.CrawlDriver()
	wg.Wait()

}
