package main

import "fmt"

func checkPokedex() {
	if len(pokedex) == 0 {
		fmt.Println("Your Pokedex: is empty")
	} else {
		fmt.Println("Your Pokedex:")
		for _, pokemon := range pokedex {
			fmt.Println(pokemon.Name)
		}
	}
}
