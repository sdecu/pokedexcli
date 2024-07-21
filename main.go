package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/sdecu/pokedexcli/internal/pokecache"
)

var cache *pokecache.Cache
var pokedex = make(map[string]pokemonInfo)

func init() {
	cache = pokecache.NewCache(5 * time.Minute)
}

func main() {
	scanner()

}

type cliCommand struct {
	name        string
	description string
	callback    func(string) error
}

func scanner() {
	scanner := bufio.NewScanner(os.Stdin)
	current := "https://pokeapi.co/api/v2/location-area"

	commands := map[string]cliCommand{
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback: func(string) error {
				fmt.Println("\nWelcome to the Pokedex!")
				fmt.Println("Usage:")
				fmt.Println("\nhelp: Displays a help message")
				fmt.Println("map: Lists the next 20 location areas")
				fmt.Println("map: Lists the previous 20 location areas")
				fmt.Println("explore: List all possible encounters for a given location 'explore <locationName>'")
				fmt.Println("catch: attempt to catch a pokemon 'catch <pokemonName>'")
				fmt.Println("inspect: inspect a pokemon to see its stats if you have caught it before 'inspect <pokemonName>'")
				fmt.Println("pokedex: list every pokemon you've caught")
				fmt.Println("exit: Exit the Pokedex")
				return nil
			},
		},
		"map": {
			name:        "map",
			description: "Lists the next 20 location areas",
			callback: func(url string) error {
				newURL, err := maps(url)
				if err != nil {
					return err
				}
				current = newURL
				return nil
			},
		},
		"mapb": {
			name:        "mapb",
			description: "List the previous 20 location areas",
			callback: func(url string) error {
				newURL, err := mapb(url)
				if err != nil {
					return err
				}
				current = newURL
				return nil
			},
		},
		"explore": {
			name:        "explore",
			description: "shows a list of all the pokemon in a given area",
			callback: func(location string) error {
				return explore(location)
			},
		},
		"catch": {
			name:        "catch",
			description: "tries to catch a pokemon when passed a pokemon name",
			callback: func(name string) error {
				return catch(name)
			},
		},
		"inspect": {
			name:        "inspect",
			description: "tries to inspect a pokemon when passed a pokemon name",
			callback: func(name string) error {
				return inspect(name)
			},
		},
		"pokedex": {
			name:        "pokedex",
			description: "list every pokemon you've caught",
			callback: func(name string) error {
				checkPokedex()
				return nil
			},
		},
	}

	for {
		fmt.Print("pokedex> ")
		scanner.Scan()
		input := scanner.Text()

		if input == "exit" {
			break
		}

		words := strings.Fields(input)
		if len(words) == 0 {
			continue
		}

		commandName := words[0]
		args := strings.Join(words[1:], " ")

		if cmd, ok := commands[commandName]; ok {
			var err error
			if commandName == "explore" ||
				commandName == "catch" ||
				commandName == "inspect" {
				if args == "" {
					fmt.Println("Please provide a location to explore")
					continue
				}
				err = cmd.callback(args)
			} else {
				err = cmd.callback(current)
			}
			if err != nil {
				fmt.Println("Error:", err)
			}
		} else {
			fmt.Println("Unknown command. Type 'help' for a list of commands.")
		}
	}
}

func parseJSON(url string) (locations, error) {
	res, err := http.Get(url)
	if err != nil {
		return locations{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return locations{}, err
	}

	if res.StatusCode > 299 {
		return locations{}, fmt.Errorf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
	}

	var l locations
	err = json.Unmarshal(body, &l)
	if err != nil {
		return locations{}, err
	}

	return l, nil
}

func maps(url string) (string, error) {
	if val, ok := cache.Get(url); ok {
		var cachedLocation locations
		err := json.Unmarshal(val, &cachedLocation)
		if err != nil {
			return url, nil
		}
		for _, place := range cachedLocation.Results {
			fmt.Println(place.Name)
		}
		return cachedLocation.Next, nil
	}

	location, err := parseJSON(url)
	if err != nil {
		return url, err
	}

	for _, place := range location.Results {
		fmt.Println(place.Name)
	}
	resultBytes, err := json.Marshal(location)
	if err != nil {
		return url, err
	}
	cache.Add(url, resultBytes)

	return location.Next, nil
}

func mapb(url string) (string, error) {
	if next, ok := cache.Get(url); ok {
		var nextLocation locations
		err := json.Unmarshal(next, &nextLocation)
		if err != nil {
			return useapi(url)
		}

		if nextLocation.Previous == nil {
			return useapi(url)
		}

		cur, ok := cache.Get(*nextLocation.Previous)
		if !ok {
			return useapi(url)
		}

		var currentLocation locations
		if err := json.Unmarshal(cur, &currentLocation); err != nil {
			return useapi(url)
		}

		if currentLocation.Previous == nil {
			return useapi(url)
		}

		prev, ok := cache.Get(*currentLocation.Previous)
		if !ok {
			return useapi(url)
		}

		var prevLocation locations
		if err := json.Unmarshal(prev, &prevLocation); err != nil {
			return useapi(url)
		}

		for _, place := range prevLocation.Results {
			fmt.Println(place.Name)
		}

		return *nextLocation.Previous, nil
	}
	return useapi(url)
}

func useapi(url string) (string, error) {

	location, err := parseJSON(url)
	if err != nil {
		return url, err
	}

	if location.Previous == nil {
		return url, errors.New("you are at the starting location")
	}

	previousLocation, err := parseJSON(*location.Previous)
	if err != nil {
		return url, err
	}

	for _, place := range previousLocation.Results {
		fmt.Println(place.Name)
	}

	resultBytes, err := json.Marshal(previousLocation)
	if err != nil {
		return url, err
	}
	cache.Add(*location.Previous, resultBytes)

	return *location.Previous, nil
}

type locations struct {
	Count    int     `json:"count"`
	Next     string  `json:"next"`
	Previous *string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}
