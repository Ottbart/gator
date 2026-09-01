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

	err = helperFeedFollow(s, currentUserID, feed.ID)
	if err != nil {
		return fmt.Errorf("can't create feed_follow record. Error: %v", err)
	}

	fmt.Printf("%v is now following %v", s.Config.CurrentUserName, feed.Name)

	return nil
}

func handlerListFeeds(s *State, cmd Command) error {
	feeds, err := s.db.ListFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("can'get feeds from database. Error: %v", err)
	}

	for _, feed := range feeds {
		userID := feed.UserID
		username, err := s.db.GetUserName(context.Background(), userID)
		if err != nil {
			return fmt.Errorf("error getting username with id: %v from database. Error: %v", userID, err)
		}
		fmt.Printf("Feed: %v\n", feed.Name)
		fmt.Printf("URL: %v\n", feed.Url)
		fmt.Printf("User: %v\n", username)
	}
	return nil
}

func handlerFollow(s *State, cmd Command) error {
	input := cmd.args
	if len(input) < 1 {
		return fmt.Errorf("not enough arguments. Use with 'follow <url>'")
	}
	url := cmd.args[0]

	//get feedID from url
	getFeed, err := s.db.GetFeedFromUrl(context.Background(), url)
	if err != nil {
		return fmt.Errorf("Can't get feed id. Please check provided URL. Error: %v", err)
	}

	// get user userID from current username
	currentUser, err := s.db.GetUser(context.Background(), s.Config.CurrentUserName)
	if err != nil {
		return fmt.Errorf("Can't get current user from user db. Error: %v", err)
	}

	err = helperFeedFollow(s, currentUser.ID, getFeed.ID)

	fmt.Printf("%v is now following %v", s.Config.CurrentUserName, getFeed.Name)

	return nil
}

func handlerFollowing(s *State, cmd Command) error {
	// get user userID from current username
	currentUser, err := s.db.GetUser(context.Background(), s.Config.CurrentUserName)
	if err != nil {
		return fmt.Errorf("Can't get current user from user db. Error: %v", err)
	}
	fmt.Printf(" User %v is following:\n", currentUser.Name)

	sliceOfFeeds, err := s.db.GetFeedFollowsForUser(context.Background(), currentUser.ID)
	if err != nil {
		return fmt.Errorf("Can't get feeds for user %v. Error: %v", currentUser.Name, err)
	}

	for _, feed := range sliceOfFeeds {
		feed_follow, err := s.db.GetFeedFromId(context.Background(), feed.FeedID)
		if err != nil {
			return fmt.Errorf("Cant't get feedId from feedFollow. Error: %v", err)
		}
		fmt.Printf("- %v\n", feed_follow.Name)
	}

	return nil
}

func helperFeedFollow(s *State, userId, feedId uuid.UUID) error {
	feed := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    userId,
		FeedID:    feedId,
	}
	_, err := s.db.CreateFeedFollow(context.Background(), feed)
	if err != nil {
		return fmt.Errorf("Can't create feed  follow. Error: %v", err)
	}

	return nil
}
