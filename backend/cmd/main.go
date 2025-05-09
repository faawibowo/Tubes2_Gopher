package main

import (
	"fmt"
	"log"

	"github.com/faawibowo/Tubes2_Gopher/configs"
	"github.com/faawibowo/Tubes2_Gopher/pkg/DataStructure/Graph"
	_ "github.com/faawibowo/Tubes2_Gopher/server/docs"
)

func main() {
	// if err := cmds.Execute(); err != nil {
	// 	log.Fatal(err)
	// }
	elements, err := configs.LoadElementsJSON("configs/elements.json")
	if err != nil {
		log.Fatal("Failed to load elements.json:", err)
	}

	scraped := configs.ToScrapedMap(elements)

	g := Graph.CreateElementGraphMap(scraped)

	fmt.Println("== Cara Membuat Dust ==")
	for _, r := range g["Dust"].Recipes {
		fmt.Println("  ", r)
	}

	fmt.Println("== Air dipakai di ==")
	for _, r := range g["Air"].UsedIn {
		fmt.Println("  ", r)
	}
}
