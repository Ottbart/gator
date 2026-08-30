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

func handlerDelete(s *State, cmd Command) error {
	err := s.db.DeleteAllUsers(context.Background())
	if err != nil {
		return err
	}
	fmt.Println("successfully delete all users from database")
	return nil
}

func handlerAgg(s *State, cmd Command) error {
	feed, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		return fmt.Errorf("error fetching rss data with error: %v", err)
	}
	fmt.Printf("Feed: %+v\n", feed)

	return nil
}

func handlerAddFeed(s *State, cmd Command) error {

	input := cmd.args

	if len(input) < 2 {
		return fmt.Errorf("could not add feed. Usage: addfeed <name> <url>")
	}
	feedName := input[0]
	feedUrl := input[1]

	currentUser, err := s.db.GetUser(context.Background(), s.Config.CurrentUserName)
	if err != nil {
		return fmt.Errorf("can't get current user from user db")
	}
	currentUserID := currentUser.ID

	feedParams := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      feedName,
		Url:       feedUrl,
		UserID:    currentUserID,
	}

	feed, err := s.db.CreateFeed(context.Background(), feedParams)
	fmt.Printf("successfully added feed %v to the database\n", feed)

	return nil
}
