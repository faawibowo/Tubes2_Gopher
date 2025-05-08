package mapping

import (
	"encoding/json"
	"os"
	"strings"
)

// Buat decode dari json
type Element struct {
	Name      string     `json:"name"`
	Recipes   [][]string `json:"recipes"`
	ImageLink string     `json:"link"`
}

// Value dari map
type Value struct {
	Recipes   [2]string
	ImageLink string
}

func loadRecipes(filename string) (map[string]Value, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var elements []Element
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&elements); err != nil {
		return nil, err
	}

	recipeMap := make(map[string]Value)

	for _, elem := range elements {
		for _, recipe := range elem.Recipes {
			if len(recipe) == 2 {
				first := strings.TrimSpace(recipe[0])
				second := strings.TrimSpace(recipe[1])

				if first > second {
					first, second = second, first
				}

				key := elem.Name
				recipeMap[key] = [2]string{first, second}
			} else if recipe == nil {
				key := elem.Name
				recipeMap[key] = [2]string{"", ""}
			}
		}
	}

	return recipeMap, nil
}
