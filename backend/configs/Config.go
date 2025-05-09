package configs

import (
	"encoding/json"
	"fmt"
	"os"
)

type ElementJSON struct {
	Name    string     `json:"name"`
	Recipes [][]string `json:"recipes"`
}

func LoadElementsJSON(path string) ([]ElementJSON, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading json: %w", err)
	}

	var result []ElementJSON
	if err := json.Unmarshal(file, &result); err != nil {
		return nil, fmt.Errorf("error parsing json: %w", err)
	}

	return result, nil
}

func ToScrapedMap(elements []ElementJSON) map[string][]string {
	scraped := make(map[string][]string)

	for _, el := range elements {
		if el.Recipes == nil {
			continue
		}
		for _, pair := range el.Recipes {
			if len(pair) == 2 {
				scraped[el.Name] = append(scraped[el.Name], pair[0], pair[1])
			}
		}
	}
	return scraped
}
