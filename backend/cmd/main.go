package main

import (
	"fmt"
	"log"

	"github.com/faawibowo/Tubes2_Gopher/configs"
	"github.com/faawibowo/Tubes2_Gopher/internal/algorithm/bfs"
	"github.com/faawibowo/Tubes2_Gopher/internal/algorithm/dfs"
	"github.com/faawibowo/Tubes2_Gopher/pkg/DataStructure/Graph"
	"github.com/faawibowo/Tubes2_Gopher/pkg/DataStructure/Tree"

	_ "github.com/faawibowo/Tubes2_Gopher/server/docs"
)

func main() {
	// Load
	elements, err := configs.LoadElementsJSON("configs/elements.json")
	if err != nil {
		log.Fatal("Failed to load elements.json:", err)
	}

	scraped := configs.ToScrapedMapWithTiers(elements)

	// Extract
	tierMap := make(map[string]int)
	ingredientMap := make(map[string][]string)
	for name, data := range scraped {
		tierMap[name] = data.Tier
		ingredientMap[name] = data.Ingredients
	}

	graphMap := Graph.CreateElementGraphMap(ingredientMap, tierMap)
	target := graphMap["Brick"] // <UBAH>

	// count := 3
	// result := bfs.BuildRecipeTree(target, graphMap, count)
	// fmt.Println("===== BFS BuildRecipeTree =====")
	// fmt.Printf("Node Count: %d\n", result.NodeCount)
	// fmt.Printf("Complete Paths: %d\n", result.CompletePaths)
	// fmt.Printf("Execution Time (ms): %d\n", result.ExecutionTimeMs)
	// fmt.Println("Resulting Tree:")
	// fmt.Println()
	// Tree.PrintTree(result.Tree.First, 0, true, []bool{})
	// fmt.Println()
	// fmt.Println()

	// result2 := dfs.BuildRecipeTree(target, graphMap, count)
	// fmt.Println("===== DFS BuildRecipeTree =====")
	// fmt.Printf("Node Count: %d\n", result2.NodeCount)
	// fmt.Printf("Complete Paths: %d\n", result2.CompletePaths)
	// fmt.Printf("Execution Time (ms): %d\n", result2.ExecutionTimeMs)
	// fmt.Println("Resulting Tree:")
	// fmt.Println()
	// Tree.PrintTree(result2.Tree.First, 0, true, []bool{})
	// fmt.Println()
	// fmt.Println()

	result4 := dfs.FindShortestPath(target, graphMap)
	fmt.Println("===== DFS FindShortestPath =====")
	fmt.Printf("Node Count: %d\n", result4.NodeCount)
	fmt.Printf("Complete Paths: %d\n", result4.CompletePaths)
	fmt.Printf("Execution Time (ms): %d\n", result4.ExecutionTimeMs)
	fmt.Println("Resulting Tree:")
	fmt.Println()
	Tree.PrintTree(result4.Tree.First, 0, true, []bool{})
	fmt.Println()
	fmt.Println()

	result3 := bfs.FindShortestPath(target, graphMap)
	fmt.Println("===== BFS FindShortestPath =====")
	fmt.Printf("Node Count: %d\n", result3.NodeCount)
	fmt.Printf("Complete Paths: %d\n", result3.CompletePaths)
	fmt.Printf("Execution Time (ms): %d\n", result3.ExecutionTimeMs)
	fmt.Println("Resulting Tree:")
	fmt.Println()
	Tree.PrintTree(result3.Tree.First, 0, true, []bool{})
	fmt.Println()
	fmt.Println()

}
