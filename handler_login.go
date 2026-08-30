package main

import (
	"context"
	"fmt"
)

func handlerLogin(s *State, cmd Command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("invalid argument for 'username'. Provide a username without spaces")
	}
	loginName := cmd.args[0]

	user, err := s.db.GetUser(context.Background(), loginName)
	if err != nil {
		return fmt.Errorf("user does not exist. please register user first with 'register' command")
	}

	err = s.Config.SetUser(user.Name)
	if err != nil {
		return fmt.Errorf("could not set username to: %v", user.Name)
	}
	fmt.Printf("successfully set user to: %v\n", user.Name)
	return nil
}
