package main

import (
	"fmt"
	"log"

	"github.com/faawibowo/Tubes2_Gopher/configs"
	"github.com/faawibowo/Tubes2_Gopher/pkg/DataStructure/Graph"

	_ "github.com/faawibowo/Tubes2_Gopher/server/docs"
)

func main() {
	// Load elements
	elements, err := configs.LoadElementsJSON("configs/elements.json")
	if err != nil {
		log.Fatal("Failed to load elements.json:", err)
	}

	// Get ingredient map with tiers
	scraped := configs.ToScrapedMapWithTiers(elements)

	// Extract tier map and flattened ingredient map
	tierMap := make(map[string]int)
	ingredientMap := make(map[string][]string)
	for name, data := range scraped {
		tierMap[name] = data.Tier
		ingredientMap[name] = data.Ingredients
	}

	// Build the graph using both
	graphMap := Graph.CreateElementGraphMap(ingredientMap, tierMap)

	fmt.Printf("Brick Tier: %d\n", graphMap["Chicken coop"].Tier)
	fmt.Printf("Mud Tier:   %d\n", graphMap["Planet"].Tier)

	// // Build and print tree (example: for element "Brick")
	// target := graphMap["Mirror"]
	// count := 4

	// tree := bidirectional.BuildUnifiedRecipeTree(target, count, 0, nil, graphMap)
	// Tree.PrintTree(tree.First, 0, true, []bool{})

	// fmt.Println("Paths: ", Tree.CountPaths(tree.First))

	// result := dfs.BuildRecipeTree(target, graphMap, count)

	// fmt.Println("===== DFS BuildRecipeTree =====")
	// fmt.Printf("Node Count: %d\n", result.NodeCount)
	// fmt.Printf("Complete Paths: %d\n", result.CompletePaths)
	// fmt.Printf("Execution Time (ms): %d\n", result.ExecutionTimeMs)
	// fmt.Println("Resulting Tree:")
	// Tree.PrintTree(result.Tree.First, 0, true, []bool{})

	// fmt.Println("Paths: ", Tree.CountPaths(result.Tree.First))

	// // ✅ SHORTEST PATH
	// shortest := dfs.FindShortestPath(target, graphMap)

	// fmt.Println("\n===== DFS Shortest Path =====")
	// fmt.Printf("Node Count: %d\n", shortest.NodeCount)
	// fmt.Printf("Complete Paths: %d\n", shortest.CompletePaths)
	// fmt.Printf("Execution Time (ms): %d\n", shortest.ExecutionTimeMs)
	// fmt.Println("Shortest Tree:")
	// Tree.PrintTree(shortest.Tree.First, 0, true, []bool{})
}
