package main

import (
	"fmt"
	"os"

	cfg "github.com/Ottbart/gator/internal/config"
)

func main() {
	_, err := cfg.ReadConfig()
	if err != nil {
		println("error reading config")
	}
	currentUser := "sascha"
	err = cfg.SetUser(currentUser)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error setting user %s\n", currentUser)
	}

	config, err := cfg.ReadConfig()
	if err != nil {
		println("error reading config")
	}
	print(config)
}
