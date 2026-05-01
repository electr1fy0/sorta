package main

import (
	"github.com/electr1fy0/sorta/cmd"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	cmd.Execute()
}
