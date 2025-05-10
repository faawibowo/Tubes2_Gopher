package Graph

import "fmt"

type ElementGraph struct {
	Name    string   // element name
	UsedIn  []Recipe // which recipes this element is a part of
	Recipes []Recipe // the recipes of this element
	Tier    int      // tier of this element
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

func CreateElementGraphMap(
	scraped map[string][]string,
	tierMap map[string]int,
) map[string]*ElementGraph {

	elementMap := make(map[string]*ElementGraph)

	for resultName, inputs := range scraped {
		// node hasil
		if _, ok := elementMap[resultName]; !ok {
			elementMap[resultName] = &ElementGraph{
				Name: resultName,
				Tier: tierMap[resultName],
			}
		}
		// node bahan
		for _, in := range inputs {
			if _, ok := elementMap[in]; !ok {
				elementMap[in] = &ElementGraph{
					Name: in,
					Tier: tierMap[in],
				}
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

			result.Recipes = append(result.Recipes, recipe)
			first.UsedIn = append(first.UsedIn, recipe)
			second.UsedIn = append(second.UsedIn, recipe)
		}
	}

	return elementMap
}

func GetBasicElement(graphMap map[string]*ElementGraph) []*ElementGraph {
	return []*ElementGraph{graphMap["Air"], graphMap["Earth"], graphMap["Water"], graphMap["Fire"]} // starter
}

func IsLeaf(node *ElementGraph, graphMap map[string]*ElementGraph) bool {
	return node == graphMap["Air"] || node == graphMap["Earth"] || node == graphMap["Water"] || node == graphMap["Fire"]
}
