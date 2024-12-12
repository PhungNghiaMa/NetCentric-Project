package Pokedex

import (
	"fmt"
	"main/Model"
	"strconv"
	"strings"
	"sync"
	"time"
	"github.com/playwright-community/playwright-go"
)

const (
	url = "https://pokedex.org/#/"
)

var pokemons []Model.Pokemon

func CrawlDriver() {
	var wg sync.WaitGroup
	pw, err := playwright.Run()
	if err != nil {
		fmt.Println("Cannot start playwright instance: ", err.Error())
		return
	}
	browser, err := pw.Chromium.Launch()
	if err != nil {
		fmt.Println("Cannot lauch browser: ", err.Error())
		return
	}

	page, err := browser.NewPage()
	if err != nil {
		fmt.Println("Cannot create page: ", err.Error())
		return
	}

	page.Goto(url)

	// Get total pokemon number need to extract to use for looping
	totalPokemons := ExtractPokemonNumber(page)
	fmt.Println("TOTAL POKEMON WILL BE EXTRACTED: ", totalPokemons)

	// Call function to crawl all pokemon
	locator := "button.sprite-1"
	wg.Add(1)
	go func() {
		defer wg.Done()
		button := page.Locator(locator).First()
		button.Click()
		time.Sleep(500 * time.Millisecond)
		fmt.Println("FOUND POKEMON: ")
		ExtractPokemon(page)
		page.Goto(url)
		page.Reload()
	}()

	wg.Wait()

}

func ExtractPokemonNumber(page playwright.Page) int {
	li, _ := page.Locator("div#monsters-list-wrapper > ul#monsters-list > li").All()
	return len(li)
}

func ExtractPokemon(page playwright.Page) {
	pokemon := Model.Pokemon{} // pokemon model

	stats := Model.Stats{} // pokemon feature model ( hp , attack , def , ....)

	// EXTRACT POKEMON STAT
	fmt.Println("EXTRACT STAT !")
	entries, _ := page.Locator("div.detail-panel-content > div.detail-header > dev.detail-infobox >div.detail-stats > div.detail-stats-rows").All()
	for _, entry := range entries {
		title, _ := entry.Locator("span:not([class])").TextContent() // find all "span" in each entry which not have the class
		switch title {
		case "HP":
			hp, _ := entry.Locator("span.stat-bar > dev.stat-bar-fg").TextContent()
			stats.HP, _ = strconv.Atoi(hp)
		case "Attack":
			atk, _ := entry.Locator("span.stat-bar > div.stat-bar-fg").TextContent()
			stats.Attack, _ = strconv.Atoi(atk)
		case "Defense":
			def, _ := entry.Locator("span.stat-bar > div.stat-bar-fg").TextContent()
			stats.Defense, _ = strconv.Atoi(def)
		case "Speed":
			speed, _ := entry.Locator("span.stat-bar > div.stat-bar-fg").TextContent()
			stats.Speed, _ = strconv.Atoi(speed)
		case "Sp Atk":
			sp_atk, _ := entry.Locator("span.stat-bar > div.stat-bar-fg").TextContent()
			stats.Sp_Attack, _ = strconv.Atoi(sp_atk)
		case "Sp Def":
			sp_def, _ := entry.Locator("span.stat-bar > div.stat-bar-fg").TextContent()
			stats.Sp_Defense, _ = strconv.Atoi(sp_def)
		default:
			fmt.Println("Unknown title: ", title)
		}
	}
	pokemon.Stats = Model.Stats(stats)

	// EXTRACT POKEMON NAME
	fmt.Println("EXTRACT NAME !")
	name, _ := page.Locator("div.detail-panel > h1.detail-panel-header").TextContent()
	pokemon.Name = name
	fmt.Println("NAME: ", name)

	// EXTRACT GENDER RATIO
	genderRatio := Model.GenderRatio{}
	profile := Model.Profile{}
	entries, _ = page.Locator("div.detail-panel-content > div.detail-below-header > div.monster-minutia").All()
	for _, entry := range entries {
		title1, _ := entry.Locator("strong:not([class]):nth-child(1)").TextContent() // get the first strong element in each div.monster--minutia
		stat1, _ := entry.Locator("span:not([class]):nth-child(2)").TextContent()    // Extract value of stat1 with the corresponding name to the title1
		title2, _ := entry.Locator("strong:not([class]):nth-child(3)").TextContent() // get the second strong element in each div.monster--minutia
		stat2, _ := entry.Locator("span:not([class]):nth-child(4)").TextContent()    // Extract value of stat1 with the corresponding name to the title1
		switch title1 {
		case "Height:":
			heightSplit := strings.Split(stat1, " ")            // split at the space to take the number without the unit , E.g: 0.7 m => Take 0.7 and store to heightSplit
			height, _ := strconv.ParseFloat(heightSplit[0], 32) // 32 bit-float number but the real value store to height still float64
			profile.Height = float32(height)
		case "Catch Rate:":
			cacheRateSplit := strings.Split(stat1, "%")
			cachRate, _ := strconv.ParseFloat(cacheRateSplit[0], 32)
			profile.CatchRate = float32(cachRate)
		case "Egg Groups:":
			eggGroupSplit := strings.Split(stat1, "]")
			profile.EggGroup = eggGroupSplit[1]
		case "Abilities:":
			profile.Abilities = stat1
		default:
			fmt.Println("Unknow title: ", title1)
		}

		switch title2 {
		case "Weight:":
			weightSplit := strings.Split(stat2, " ")
			weight, _ := strconv.ParseFloat(weightSplit[0], 32) // 32 bit-float number but the real value store to weight still float64
			profile.Weight = float32(weight)
		case "Gender Ratio:":
			if stat2 == "N/A" {
				genderRatio.MaleRatio = 0
				genderRatio.FemaleRatio = 0
			} else {
				ratios := strings.Split(stat2, " ")
				maleRatioString := strings.Split(ratios[0], "%")
				femaleRatioString := strings.Split(ratios[2], "%")
				maleRatio, _ := strconv.ParseFloat(maleRatioString[0], 32)
				femaleRatio, _ := strconv.ParseFloat(femaleRatioString[0], 32)
				genderRatio.MaleRatio = float32(maleRatio)
				genderRatio.FemaleRatio = float32(femaleRatio)
			}
			profile.GenderRatio = genderRatio
		case "Hatch Steps:":
			profile.HatchSteps, _ = strconv.Atoi(stat2)
		}
	}
	pokemon.Profile = profile
	// fmt.Println("PROFILE: ", pokemon.Profile)

	// EXTRACT DAMAMGE WHEN ATTACK
	damageWhenAttacked := []Model.DamageWhenAttacked{}
	entries, _ = page.Locator("div.when-attacked > div.when-attacked-row").All()
	fmt.Println("TOTAL MAPPING COMPONENT DMWA: ", len(entries))
	for _, entry := range entries {
		element1, _ := entry.Locator("span.monster-type:nth-child(1)").TextContent()
		damage1, _ := entry.Locator("span.monster-multiplier:nth-child(2)").TextContent()
		element2, _ := entry.Locator("span.monster-type:nth-child(3)").TextContent()
		damage2, _ := entry.Locator("span.monster-multiplier:nth-child(4)").TextContent()
		Damage1, _ := strconv.ParseFloat(strings.Split(damage1, "x")[0], 32)
		Damage2, _ := strconv.ParseFloat(strings.Split(damage2, "x")[0], 32)
		// fmt.Println("\nDAMAGE WHEN ATTACK: \n")
		// fmt.Printf("Element %s: %s", element1, damage1)
		// fmt.Printf("Element %s: %s", element2, damage2)
		damageWhenAttacked = append(damageWhenAttacked, Model.DamageWhenAttacked{Element: element1, Coefficient: float32(Damage1)})
		// fmt.Println("DMWA-1: ", damageWhenAttacked)
		damageWhenAttacked = append(damageWhenAttacked, Model.DamageWhenAttacked{Element: element2, Coefficient: float32(Damage2)})
		// fmt.Println("DMWA-2: ", damageWhenAttacked[1])
	}
	pokemon.DamageWhenAttacked = damageWhenAttacked
	// fmt.Println("DAMAGE WHEN ATTACK: ", pokemon.DamageWhenAttacked)

	// EXTRACT EVOLUTION
	fmt.Println("EXTRACT EVOLUTION")
	entries, _ = page.Locator("div.evolutions > div.evolution-row").All()
	for _, entry := range entries {
		evolutionLabel, _ := entry.Locator("div.evolution-label > span").TextContent()
		evolutionLabelSplit := strings.Split(evolutionLabel, " ")
		if evolutionLabelSplit[0] == name { // compare the name of first evolution level with the extract name above
			LevelString := strings.Split(evolutionLabel, "level ")
			evolutionLevel, _ := strconv.Atoi(strings.Split(LevelString[1], ".")[0])
			pokemon.EvolutionLevel = evolutionLevel
		}
	}
	// fmt.Println("\nEVOLUTION LEVEL: ", pokemon.EvolutionLevel)

	// EXTRACT MOVES
	fmt.Println("EXTRACT MOVES")
	moves := []Model.Moves{}
	entries, _ = page.Locator("div.monster-moves > div.moves-row").All()
	for _, entry := range entries {
		// Simulate click the expand button in the div.moves-inner-rows to get data
		expandButton := entry.Locator("div.moves-inner-row > button.dropdown-button").First()
		expandButton.Click()

		move_name, _ := entry.Locator("div.moves-inner-row > span:nth-child(2)").TextContent()
		element, _ := entry.Locator("div.moves-inner-row > span.monster-type").TextContent()

		// Extract each component when click the button
		powerSplit, _ := entry.Locator("div.moves-row-detail > div.moves-row-stats > span:nth-child(1)").TextContent()
		power := strings.Split(powerSplit, ": ")
		power[1] = strings.TrimSpace(power[1])

		accSplit, _ := entry.Locator("div.moves-row-detail > div.moves-row-stats > span:nth-child(2)").TextContent()
		acc := strings.Split(accSplit, ": ")
		accValue, _ := strconv.Atoi(strings.Split(acc[1], "%")[0])

		PPSplit, _ := entry.Locator("div.moves-row-detail > div.moves-row-stats > span:nth-child(3)").TextContent()
		pp := strings.Split(PPSplit, ": ")
		ppValue, _ := strconv.Atoi(pp[1])

		description, _ := entry.Locator("div.moves-row-detail > div.move-description").TextContent()
		moves = append(moves, Model.Moves{Name: move_name, Element: element, Power: power[1], Acc: accValue, PP: ppValue, Description: description})
	}
	pokemon.Moves = moves
	// fmt.Println("\nMOVE: ", moves)

	// EXTRACT POKEMON TYPES
	fmt.Println("EXTRACT TYPES")
	entries, _ = page.Locator("div.detail-types > span.monster-type").All()
	for _, entry := range entries {
		element, _ := entry.TextContent()
		pokemon.Elements = append(pokemon.Elements, element)
	}
	// fmt.Println("\nTYPE: ", pokemon.Elements)

	fmt.Println("Finish add new Pokemon !")
	pokemons = append(pokemons, pokemon)

}
