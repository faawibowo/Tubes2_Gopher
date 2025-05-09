package Graph

import "fmt"

type ElementGraph struct {
	Name    string   // element name
	UsedIn  []Recipe // which recipes this element is a part of
	Recipes []Recipe // the recipes of this element
}

type Recipe struct {
	FirstElement  *ElementGraph
	SecondElement *ElementGraph
	ResultElement *ElementGraph
}

func (r Recipe) String() string {
	return fmt.Sprintf("%s + %s → %s",
		r.FirstElement.Name,
		r.SecondElement.Name,
		r.ResultElement.Name)
}

// Create Map String & Graph
func CreateElementGraphMap(scraped map[string][]string) map[string]*ElementGraph {
	elementMap := make(map[string]*ElementGraph)

	for resultName, inputs := range scraped {
		// node hasil
		if _, ok := elementMap[resultName]; !ok {
			elementMap[resultName] = &ElementGraph{Name: resultName}
		}
		// node bahan
		for _, in := range inputs {
			if _, ok := elementMap[in]; !ok {
				elementMap[in] = &ElementGraph{Name: in}
			}
		}
	}

	for resultName, inputs := range scraped {
		for i := 0; i+1 < len(inputs); i += 2 {
			first := elementMap[inputs[i]]
			second := elementMap[inputs[i+1]]
			result := elementMap[resultName]

			recipe := Recipe{
				FirstElement:  first,
				SecondElement: second,
				ResultElement: result,
			}

			// simpan ke daftar “cara membuat” si hasil
			result.Recipes = append(result.Recipes, recipe)
			// simpan ke daftar “dipakai di” kedua bahan
			first.UsedIn = append(first.UsedIn, recipe)
			second.UsedIn = append(second.UsedIn, recipe)
		}
	}

	return elementMap
}

// basic Element
func GetBasicElement(graphMap map[string]*ElementGraph) []*ElementGraph {
	return []*ElementGraph{graphMap["Air"], graphMap["Earth"], graphMap["Water"], graphMap["Fire"]} // starter
}

func IsLeaf(node *ElementGraph, graphMap map[string]*ElementGraph) bool {
	return node == graphMap["Air"] || node == graphMap["Earth"] || node == graphMap["Water"] || node == graphMap["Fire"]
}
