package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func explore(location string) error {
	if cachedData, found := cache.Get(location); found {
		fmt.Println("Exploring " + location + " from cache")
		fmt.Println("Found Pokemon:")
		var cachedPokemons pokemons
		if err := json.Unmarshal(cachedData, &cachedPokemons); err != nil {
			return fmt.Errorf("cache unmarshal error: %v", err)
		}

		for _, encounter := range cachedPokemons.PokemonEncounters {
			fmt.Println(encounter.Pokemon.Name)
		}
		return nil
	}

	url := "https://pokeapi.co/api/v2/location-area/" + location

	pokemon, err := parselocationJSON(url)
	if err != nil {
		return err
	}

	fmt.Println("Exploring " + location)
	fmt.Println("Found Pokemon:")
	for _, encounter := range pokemon.PokemonEncounters {
		fmt.Println(encounter.Pokemon.Name)
	}
	resultBytes, err := json.Marshal(pokemon)
	if err != nil {
		return fmt.Errorf("error marshaling data for cache: %v", err)
	}
	cache.Add(location, resultBytes)

	return nil
}

func parselocationJSON(url string) (pokemons, error) {
	res, err := http.Get(url)
	if err != nil {
		return pokemons{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return pokemons{}, err
	}

	if res.StatusCode > 299 {
		return pokemons{}, fmt.Errorf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
	}

	var l pokemons
	err = json.Unmarshal(body, &l)
	if err != nil {
		return pokemons{}, err
	}

	return l, nil
}

type pokemons struct {
	EncounterMethodRates []struct {
		EncounterMethod struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"encounter_method"`
		VersionDetails []struct {
			Rate    int `json:"rate"`
			Version struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"version"`
		} `json:"version_details"`
	} `json:"encounter_method_rates"`
	GameIndex int `json:"game_index"`
	ID        int `json:"id"`
	Location  struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"location"`
	Name  string `json:"name"`
	Names []struct {
		Language struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"language"`
		Name string `json:"name"`
	} `json:"names"`
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"pokemon"`
		VersionDetails []struct {
			EncounterDetails []struct {
				Chance          int   `json:"chance"`
				ConditionValues []any `json:"condition_values"`
				MaxLevel        int   `json:"max_level"`
				Method          struct {
					Name string `json:"name"`
					URL  string `json:"url"`
				} `json:"method"`
				MinLevel int `json:"min_level"`
			} `json:"encounter_details"`
			MaxChance int `json:"max_chance"`
			Version   struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"version"`
		} `json:"version_details"`
	} `json:"pokemon_encounters"`
}
