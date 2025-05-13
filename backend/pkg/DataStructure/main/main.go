package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/faawibowo/Tubes2_Gopher/configs"
	"github.com/faawibowo/Tubes2_Gopher/pkg/DataStructure/Dree"
	"github.com/faawibowo/Tubes2_Gopher/pkg/DataStructure/Graph"
	"github.com/faawibowo/Tubes2_Gopher/pkg/scraping"
)

func main() {
	// Step 1: scrape data
	fmt.Println("Scraping data...")
	elements, err := scraping.ScrapeDatafromWeb()
	if err != nil {
		fmt.Println("❌ Failed to scrape:", err)
		return
	}

	// Step 2: convert to tierMap and ingredientMap
	fmt.Println("Converting to graph...")
	scraped := configs.ToScrapedMapWithTiers(elements)
	tierMap := make(map[string]int)
	ingredientMap := make(map[string][]string)

	for name, data := range scraped {
		tierMap[name] = data.Tier
		ingredientMap[name] = data.Ingredients
	}

	// Step 3: build graph map
	graphMap := Graph.CreateElementGraphMap(ingredientMap, tierMap)

	// Step 4: pilih target
	targetName := "Treasure" // ← ganti nama elemen target di sini
	target, ok := graphMap[targetName]
	if !ok {
		fmt.Println("❌ Target element not found:", targetName)
		return
	}

	// Step 5: build Dree tree
	maxPaths := 1 // ← ganti sesuai batas maksimal path
	fmt.Println("Building DFS tree for", targetName, "with maxPaths =", maxPaths)
	tree := Dree.BuildDreeDFSConcurrent(target, maxPaths, graphMap)

	// Step 6: print hasil sebagai JSON
	output, err := json.MarshalIndent(tree, "", "  ")
	if err != nil {
		fmt.Println("❌ Failed to marshal tree:", err)
		return
	}

	// Output ke file atau terminal
	err = os.WriteFile("dree_output.json", output, 0644)
	if err != nil {
		fmt.Println("❌ Failed to write output file:", err)
		return
	}

	Dree.PrintDreeTree(tree.First, 0, true, []bool{})

	fmt.Println("✅ Tree generated and saved to dree_output.json")
}
