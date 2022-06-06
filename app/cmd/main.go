package main

import (
	"app/internal/app"
	"fmt"
	"os"
)

func main() {
	arguments := os.Args
	if len(arguments) == 1 {
		fmt.Println("Address is needed as an argument")
		return
	}
	addr := arguments[1]
	app.Run(addr)
}
