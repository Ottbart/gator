package main

import (
	"context"
	"fmt"
)

func handlerDelete(s *State, cmd Command) error {
	err := s.db.DeleteAllUsers(context.Background())
	if err != nil {
		return err
	}
	fmt.Println("successfully delete all users from database")
	return nil
}
