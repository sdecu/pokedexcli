package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sdecu/internal/pokeapi"
	"io"
	"net/http"
	"os"
)

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
	}

	for {
		fmt.Print("pokedex> ")
		scanner.Scan()
		input := scanner.Text()

		if input == "exit" {
			break
		}

		if cmd, ok := commands[input]; ok {
			err := cmd.callback(current)
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
	location, err := parseJSON(url)
	if err != nil {
		return "", err
	}

	for _, place := range location.Results {
		fmt.Println(place.Name)
	}

	return location.Next, nil
}

func mapb(url string) (string, error) {
	if url == "https://pokeapi.co/api/v2/location-area" {
		return maps(url)
	}
	location, err := parseJSON(url)
	if err != nil {
		return "", err
	}

	location, err = parseJSON(*location.Previous)
	if err != nil {
		return "", err
	}

	if location.Previous == nil {
		return "", errors.New("you are at the starting location")
	}

	return maps(*location.Previous)
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
