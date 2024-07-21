package main

import "fmt"

func inspect(name string) error {
	if val, ok := pokedex[name]; ok {
		fmt.Println("Height:" + string(val.Height))
		fmt.Println("Weight:" + string(val.Weight))
		fmt.Println("Stats:")
		fmt.Println("  -hp:" + string(val.Stats[0].BaseStat))
		fmt.Println("  -attack:" + string(val.Stats[1].BaseStat))
		fmt.Println("  -defense:" + string(val.Stats[2].BaseStat))
		fmt.Println("  -special-attack:" + string(val.Stats[3].BaseStat))
		fmt.Println("  -special-defense:" + string(val.Stats[4].BaseStat))
		fmt.Println("  -speed" + string(val.Stats[5].BaseStat))
		fmt.Println("Types:")
		for i := range val.Types {
			fmt.Println("  - " + string(val.Types[i].Type.Name))
		}
		return nil
	}

	fmt.Println("pokemon not caught yet")
	return nil
}
