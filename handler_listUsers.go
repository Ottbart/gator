package main

import (
	"context"
	"fmt"
)

func handlerListUsers(s *State, cmd Command) error {
	users, err := s.db.ListUsers(context.Background())
	if err != nil {
		return fmt.Errorf("couldn't list users from database")
	}
	currentUser := s.Config.CurrentUserName

	for _, user := range users {
		if user == currentUser {
			fmt.Printf("- %v (current)\n", user)
		} else {
			fmt.Printf("- %v\n", user)
		}
	}
	return nil
}
