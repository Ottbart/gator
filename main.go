package main

import (
	"fmt"
)

func main() error {
	config, err := cfg.read_config()
	if err != nil {
		return fmt.Errorf("error reading config file")
	}
}
