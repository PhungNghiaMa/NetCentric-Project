package Pokedex

import (
	"encoding/json"
	"fmt"
	"log"
	"main/Model"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

const (
	url = "https://pokedex.org/#/"
)

type POKEMON struct {
	Pokelist []Model.Pokemon
}

var pokemons []Model.Pokemon

func CrawlDriver() {
	// var wg sync.WaitGroup
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

	for i := range 10 {
		// simulate clicking the button to open the pokemon details
		locator := fmt.Sprintf("button.sprite-%d", i+1)
		fmt.Println("Attempting to click:", locator)
		button := page.Locator(locator).First()

		// Check visibility with error handling
		isVisible, err := button.IsVisible()
		if err != nil {
			fmt.Printf("Error checking visibility for locator %s: %v\n", locator, err)
			continue
		}
		if !isVisible {
			fmt.Println("Button not visible, skipping:", locator)
			continue
		}
		time.Sleep(1000 * time.Millisecond)
		button.Click()

		fmt.Print("Pokemon ", i+1, " ")
		ExtractPokemon(page)

		// Return to the list without reloading the entire page
		page.GoBack()

	}

	CreateJson(pokemons)
	if err = browser.Close(); err != nil {
		log.Fatalf("Cannot close the browser: %v", err.Error())
		os.Exit(1)
	}

	if err = pw.Stop(); err != nil {
		log.Fatalf("Fail to stop Playwright: %v", err.Error())
	}
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
	entries, _ := page.Locator("div.detail-panel-content > div.detail-header > div.detail-infobox >div.detail-stats > div.detail-stats-row").All()
	fmt.Println("Total entries STAT: ", len(entries))
	for _, entry := range entries {
		title, _ := entry.Locator("span:not([class])").TextContent() // find all "span" in each entry which not have the class
		switch title {
		case "HP":
			hp, _ := entry.Locator("span.stat-bar > div.stat-bar-fg").TextContent()
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
	pokemon.Stats = stats
	fmt.Println("STATS: ", pokemon.Stats)

	// EXTRACT POKEMON NAME
	fmt.Println("EXTRACT NAME !")
	name, _ := page.Locator("div.detail-panel > h1.detail-panel-header").TextContent()
	pokemon.Name = name
	fmt.Println("NAME: ", name)

	// EXTRACT GENDER RATIO
	fmt.Println("EXTRACT GENDER RATIO !")
	genderRatio := Model.GenderRatio{}
	profile := Model.Profile{}
	entries, _ = page.Locator("div.detail-panel-content > div.detail-below-header > div.monster-minutia").All()
	fmt.Println("Total entries Gender: ", len(entries))
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
			fmt.Println("Height: ", profile.Height)
		case "Catch Rate:":
			cacheRateSplit := strings.Split(stat1, "%")
			cachRate, _ := strconv.ParseFloat(cacheRateSplit[0], 32)
			profile.CatchRate = float32(cachRate)
			fmt.Println("Catch-Rate: ", profile.CatchRate)
		case "Egg Groups:":
			var eggGroupSplit string
			if strings.Contains(stat1, "]") {
				eggGroupSplit = strings.Split(stat1, "]")[1]

			} else {
				eggGroupSplit = stat1
			}
			profile.EggGroup = eggGroupSplit
			fmt.Println("Egg-group: ", profile.EggGroup)
		case "Abilities:":
			profile.Abilities = stat1
			fmt.Println("Abilities: ", profile.EggGroup)
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
	fmt.Println("PROFILE: ", pokemon.Profile)

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
	//fmt.Println("DAMAGE WHEN ATTACK: ", pokemon.DamageWhenAttacked)

	// EXTRACT EVOLUTION
	fmt.Println("EXTRACT EVOLUTION")
	entries, _ = page.Locator("div.evolutions > div.evolution-row").All()
	for _, entry := range entries {
		evolutionLabel, _ := entry.Locator("div.evolution-label > span").First().TextContent()
		evolutionLabelSplit := strings.Split(evolutionLabel, " ")
		if evolutionLabelSplit[0] == name { // compare the name of first evolution level with the extract name above
			if strings.Contains(evolutionLabel, "at level ") {
				LevelString := strings.Split(evolutionLabel, "level ")
				evolutionLevel, _ := strconv.Atoi(strings.Split(LevelString[1], ".")[0])
				pokemon.EvolutionLevel = evolutionLevel
				pokemon.EvolutionCondition = "N/A"
			} else if strings.Contains(evolutionLabel, "using") {
				pokemon.EvolutionLevel = 0
				ConditionSplit := strings.Split(evolutionLabel, "using")
				pokemon.EvolutionCondition = "using" + ConditionSplit[1]
			}

		} else {
			pokemon.EvolutionLevel = 0
			pokemon.EvolutionCondition = "N/A"
		}
	}
	//fmt.Println("\nEVOLUTION LEVEL: ", pokemon.EvolutionLevel)

	// // EXTRACT MOVES
	// fmt.Println("EXTRACT MOVES")
	// moves := []Model.Moves{}
	// entries, _ = page.Locator("div.detail-below-header > div.monster-moves > div.moves-row").All()
	// fmt.Println("Total move row: ", len(entries))
	// for _, entry := range entries {
	// 	// Simulate click the expand button in the div.moves-inner-rows to get data
	// 	expandButton := entry.Locator("div.moves-inner-row > button.dropdown-button").First()
	// 	expandButton.Click()

	// 	move_name, _ := entry.Locator("div.moves-inner-row > span:nth-child(2)").TextContent()
	// 	element, _ := entry.Locator("div.moves-inner-row > span.monster-type").TextContent()

	// 	// Extract each component when click the button
	// 	powerSplit, _ := entry.Locator("div.moves-row-detail > div.moves-row-stats > span:nth-child(1)").TextContent()
	// 	power := strings.Split(powerSplit, ": ")
	// 	power[1] = strings.TrimSpace(power[1])
	// 	fmt.Println("POWER:", power[1])

	// 	accSplit, _ := entry.Locator("div.moves-row-detail > div.moves-row-stats > span:nth-child(2)").TextContent()
	// 	acc := strings.Split(accSplit, ": ")
	// 	accValue, _ := strconv.Atoi(strings.Split(acc[1], "%")[0])
	// 	fmt.Println("ACC:", accValue)

	// 	PPSplit, _ := entry.Locator("div.moves-row-detail > div.moves-row-stats > span:nth-child(3)").TextContent()
	// 	pp := strings.Split(PPSplit, ": ")
	// 	ppValue, _ := strconv.Atoi(pp[1])
	// 	fmt.Println("PP:", ppValue)

	// 	description, _ := entry.Locator("div.moves-row-detail > div.move-description").TextContent()
	// 	moves = append(moves, Model.Moves{Name: move_name, Element: element, Power: power[1], Acc: accValue, PP: ppValue, Description: description})
	// }
	// pokemon.Moves = moves
	// fmt.Println("\nMOVE: ", pokemon.Moves)

	// EXTRACT POKEMON TYPES
	fmt.Println("EXTRACT TYPES")
	entries, _ = page.Locator("div.detail-types > span.monster-type").All()
	for _, entry := range entries {
		element, _ := entry.TextContent()
		pokemon.Elements = append(pokemon.Elements, element)
	}
	//fmt.Println("\nTYPE: ", pokemon.Elements)

	fmt.Println("Finish add new Pokemon !")
	pokemons = append(pokemons, pokemon)
}

func CreateJson(pokemons []Model.Pokemon) {
	data, err := json.MarshalIndent(pokemons, "", "    ")
	if err != nil {
		fmt.Println("Error marshalling pokemons: ", err.Error())
		return
	}

	file, err := os.Create("POKEMONS.json")
	if err != nil {
		fmt.Println("Fail to create json: ", err.Error())
		return
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		fmt.Println("Error writing to file: ", err.Error())
		return
	}
}
