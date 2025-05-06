package graph

type Graph struct {
	Element      string       // element
	Results      []*Graph     // what the element can make
	Combinations []GraphTuple // combinations to make this element
}

type GraphTuple struct {
	FirstElement  *Graph
	SecondElement *Graph
}

// Create Map String & Graph
func createGraphMap(scrapedMap map[string][]string) map[string]*Graph {

	graphMap := make(map[string]*Graph)

	// Create graphMap kalo nil
	for key, elements := range scrapedMap {
		if graphMap[key] == nil {
			graphMap[key] = &Graph{Element: key}
		}
	}

	// Add Combinations & Results
	for key, values := range scrapedMap {

		if len(values) < 2 {
			continue
		}

		firstKey := values[0]
		secondKey := values[1]

		// Recipe blom punya graph node -> create one
		firstNode, ok := graphMap[firstKey]
		if !ok {
			firstNode = &Graph{Element: firstKey}
			graphMap[firstKey] = firstNode
		}

		secondNode, ok := graphMap[secondKey]
		if !ok {
			secondNode = &Graph{Element: secondKey}
			graphMap[secondKey] = secondNode
		}

		// Combination Tuple
		combination := GraphTuple{
			FirstElement:  firstNode,
			SecondElement: secondNode,
		}
		// Append to Combinations
		graphMap[key].Combinations = append(graphMap[key].Combinations, combination)

		// Append Results ke Graph recipe (for bidirectional)
		if graphMap[firstKey].Element != graphMap[key].Element {
			graphMap[firstKey].Results = append(graphMap[firstKey].Results, graphMap[key])
		}

		if graphMap[secondKey].Element != graphMap[key].Element {
			graphMap[secondKey].Results = append(graphMap[secondKey].Results, graphMap[key])
		}
	}

	return graphMap
}

// basic Element
func getBasicElement(map[string]*Graph graphMap) []*Graph {
	return graphMap["Air"], graphMap["Earth"], graphMap["Water"], graphMap["Fire"] // starter
}

