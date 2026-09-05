package main

import (
	"fmt"
)

type Command struct {
	name string
	args []string
}

type Commands struct {
	handlers map[string]func(*State, Command) error
}

func (c *Commands) register(name string, f func(*State, Command) error) {
	c.handlers[name] = f
}

func (c *Commands) run(s *State, cmd Command) error {
	function, ok := c.handlers[cmd.name]
	if ok {
		return function(s, cmd)
	}
	return fmt.Errorf("command not found")
}
