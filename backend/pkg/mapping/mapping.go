package mapping

import (
	"encoding/json"
	"os"
	"strings"
)

type Element struct {
	Name    string     `json:"name"`
	Recipes [][]string `json:"recipes"`
}

func loadRecipes(filename string) (map[string][2]string, error) {
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

	recipeMap := make(map[string][2]string)

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
