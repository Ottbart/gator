package main

import (
	"context"
	"fmt"
	"time"

	"github.com/Ottbart/gator/internal/database"
	"github.com/google/uuid"
)

func handlerRegister(s *State, cmd Command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("invalid argument for 'username'. Provide a username without spaces")
	}
	username := cmd.args[0]

	userParams := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      username,
	}
	user, err := s.db.CreateUser(context.Background(), userParams)
	if err != nil {
		return fmt.Errorf("error creating user in db: %v", err)
	}

	err = s.Config.SetUser(user.Name)
	if err != nil {
		return fmt.Errorf("could not set username to: %v", user.Name)
	}
	fmt.Printf("successfully created user: \n")
	fmt.Printf(" * ID:      %v\n", user.ID)
	fmt.Printf(" * Name:    %v\n", user.Name)
	return nil
}
