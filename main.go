package main

import (
	"log"
	"os"

	cfg "github.com/Ottbart/gator/internal/config"
)

type State struct {
	Config *cfg.Config
}

func main() {

	read, err := cfg.ReadConfig()
	if err != nil {
		log.Fatalf("error reading config: %v", err)
		return
	}
	programState := &State{
		Config: &read,
	}

	cmds := Commands{
		handlers: make(map[string]func(*State, Command) error),
	}
	cmds.register("login", handlerLogin)

	input := os.Args
	if len(input) < 2 {
		log.Fatalf("no command found")
	} else if len(input) < 3 {
		log.Fatalf("no arguments given")
	} else {
		cmd := Command{
			name: input[1],
			args: input[2:],
		}
		err := cmds.run(programState, cmd)
		if err != nil {
			log.Fatal(err)
		}
	}
}
