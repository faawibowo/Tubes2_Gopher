package main

import (
	"log"

	"github.com/faawibowo/Tubes2_Gopher/cmd/cmds"
	_ "github.com/faawibowo/Tubes2_Gopher/server/docs"
)

func main() {
	if err := cmds.Execute(); err != nil {
		log.Fatal(err)
	}
}
