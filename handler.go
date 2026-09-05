package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/Ottbart/gator/internal/database"
	"github.com/google/uuid"
)

func middlewareLoggedIn(handler func(s *State, cmd Command, user database.User) error) func(*State, Command) error {
	return func(s *State, cmd Command) error {
		currentUser, err := s.db.GetUser(context.Background(), s.Config.CurrentUserName)
		if err != nil {
			return fmt.Errorf("can't get current user from user db. Error %v", err)
		}
		return handler(s, cmd, currentUser)
	}
}

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
	if len(cmd.args) < 1 || len(cmd.args) > 2 {
		return fmt.Errorf("usage: %v <time_between_reqs>", cmd.name)
	}

	timeBetweenRequests, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return fmt.Errorf("invalid duration: %w", err)
	}

	log.Printf("Collecting feeds every %s...", timeBetweenRequests)

	ticker := time.NewTicker(timeBetweenRequests)

	for ; ; <-ticker.C {
		err = scrapeFeeds(s)
		if err != nil {
			return fmt.Errorf("error scraping feeds. Error: %v", err)
		}
	}
}

func scrapeFeeds(s *State) error {
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("can't get list of feeds to fetch from database. Error: %v", err)
	}
	fmt.Println("Found a feed to fetch!")

	s.db.MarkFeedFetched(context.Background(), feed.ID)

	feedData, err := fetchFeed(context.Background(), feed.Url)
	if err != nil {
		return fmt.Errorf("can't fetch feed. Error %v", err)
	}
	fmt.Printf("writing all posts from '%v' to the database", feedData.Channel.Title)

	for _, item := range feedData.Channel.Item {
		publishedAt := sql.NullTime{}
		if t, err := time.Parse(time.RFC1123Z, item.PubDate); err == nil {
			publishedAt = sql.NullTime{
				Time:  t,
				Valid: true,
			}
		}
		params := database.CreatePostParams{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			FeedID:    feed.ID,
			Title:     item.Title,
			Url:       item.Link,
			Description: sql.NullString{
				String: item.Description,
				Valid:  true,
			},
			PublishedAt: publishedAt,
		}

		_, err := s.db.CreatePost(context.Background(), params)
		if err != nil {
			return fmt.Errorf("error writing post to database. Error: %v", err)
		}
	}
	return nil
}

func handlerAddFeed(s *State, cmd Command, user database.User) error {

	input := cmd.args

	if len(input) < 2 {
		return fmt.Errorf("could not add feed. Usage: addfeed <name> <url>")
	}
	feedName := input[0]
	feedUrl := input[1]

	currentUserID := user.ID

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

func handlerFollow(s *State, cmd Command, user database.User) error {
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

	err = helperFeedFollow(s, user.ID, getFeed.ID)

	fmt.Printf("%v is now following %v", user.Name, getFeed.Name)

	return nil
}

func handlerFollowing(s *State, cmd Command, user database.User) error {
	//fetch feeds for user
	sliceOfFeeds, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("Can't get feeds for user %v. Error: %v", user.Name, err)
	}

	fmt.Printf(" User %v is following:\n", user.Name)

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

func handlerUnfollow(s *State, cmd Command, user database.User) error {
	input := cmd.args
	if len(input) < 1 {
		return fmt.Errorf("No arguments given. Use with 'unfollow <url>'")
	}
	url := input[0]
	feed, err := s.db.GetFeedFromUrl(context.Background(), url)
	if err != nil {
		return fmt.Errorf("couldn't get feedID from url %v\nError: %v", url, err)
	}

	params := database.DeleteFeedFollowsParams{
		UserID: user.ID,
		FeedID: feed.ID,
	}
	err = s.db.DeleteFeedFollows(context.Background(), params)
	if err != nil {
		return fmt.Errorf("Couldn't delete FeedFollows record. Error: %v", err)
	}

	return nil
}

func handlerBrowse(s *State, cmd Command, user database.User) error {
	input := cmd.args
	limit := 2
	if len(input) > 0 {
		parsedLimit, err := strconv.Atoi(input[0])
		if err != nil {
			return fmt.Errorf("Invalid limit argument. Error: %v", err)
		} else {
			limit = parsedLimit
		}
	}
	params := database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
	}
	posts, err := s.db.GetPostsForUser(context.Background(), params)
	if err != nil {
		return fmt.Errorf("error getting posts for user. Error: %v", err)
	}
	fmt.Printf("found %v posts for user %v\n", len(posts), user.Name)
	for _, post := range posts {
		fmt.Printf("Title: %v\n", post.Title)
	}
	return nil
}
