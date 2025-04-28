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

func loadRecipes(filename string) (map[string]string, error) {
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

	recipeMap := make(map[string]string)

	for _, elem := range elements {
		for _, recipe := range elem.Recipes {
			if len(recipe) == 2 {
				first := strings.TrimSpace(recipe[0])
				second := strings.TrimSpace(recipe[1])

				if first > second {
					first, second = second, first
				}

				key := first + "+" + second
				recipeMap[key] = elem.Name
			}
		}
	}

	return recipeMap, nil
}

func findResult(recipeMap map[string]string, component1, component2 string) (string, bool) {
	first := strings.TrimSpace(component1)
	second := strings.TrimSpace(component2)

	if first > second {
		first, second = second, first
	}

	key := first + "+" + second

	result, found := recipeMap[key]
	return result, found
}
