package main

import "fmt"

func handlerLogin(s *State, cmd Command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("invalid argument for 'username'. Provide a username without spaces")
	}
	loginName := cmd.args[0]
	err := s.Config.SetUser(loginName)
	if err != nil {
		return fmt.Errorf("could not set username to: %v", loginName)
	}
	fmt.Printf("successfully set user to: %v\n", loginName)
	return nil
}
