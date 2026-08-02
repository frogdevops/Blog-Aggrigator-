package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/frogdevops/gator/internal/config"
	"github.com/frogdevops/gator/internal/database"
	"github.com/google/uuid"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
}

type command struct {
	Name string
	Args []string
}

type commands struct {
	registry map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	handler, ok := c.registry[cmd.Name]
	if !ok {
		return fmt.Errorf("unknown command: %s", cmd.Name)
	}
	return handler(s, cmd)
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.registry[name] = f
}

func parseCommands(args []string) (command, error) {
	if len(args) < 2 {
		return command{}, fmt.Errorf("usage: gator <command> [args...]")
	}
	return command{Name: args[1], Args: args[2:]}, nil
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: login <username>")
	}
	name := cmd.Args[0]
	user, err := s.db.GetUser(context.Background(), name)
	if err != nil {
		return fmt.Errorf("user %s not found: %w", name, err)
	}
	if err := s.cfg.SetUser(user.Name); err != nil {
		return fmt.Errorf("set user: %w", err)
	}

	fmt.Printf("logged in as %s\n", user.Name)
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: register <name>")
	}

	name := cmd.Args[0]

	dbParam := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      name,
	}

	user, err := s.db.CreateUser(context.Background(), dbParam)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	if err := s.cfg.SetUser(user.Name); err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	fmt.Printf("user %s created\n", user.Name)
	return nil
}

func handlerReset(s *state, _ command) error {
	if err := s.db.Reset(context.Background()); err != nil {
		return fmt.Errorf("reset: %w", err)
	}
	fmt.Println("all users has been deleted.")
	return nil
}

func handlerGetUsers(s *state, _ command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("get users: %w", err)
	}

	for _, user := range users {
		if s.cfg.CurrentUserName == user.Name {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}
	return nil
}

func scrapeFeeds(s *state) {
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		log.Printf("get next feed: %v", err)
		return
	}

	if err := s.db.MarkFeedFetched(context.Background(), feed.ID); err != nil {
		log.Printf("mark feed %s fetched: %v", feed.Name, err)
		return
	}

	rss, err := fetchFeed(context.Background(), feed.Url)
	if err != nil {
		log.Printf("fetch %s: %v", feed.Url, err)
		return
	}

	for _, item := range rss.Channel.Item {
		publishedAt := sql.NullTime{}
		if t, err := time.Parse(time.RFC1123Z, item.PubDate); err == nil {
			publishedAt = sql.NullTime{Time: t, Valid: true}
		}

		dbParams := database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			Title:       item.Title,
			Url:         item.Link,
			Description: item.Description,
			PublishedAt: publishedAt,
			FeedID:      feed.ID,
		}

		if err := s.db.CreatePost(context.Background(), dbParams); err != nil {
			log.Printf("create post %s: %v", item.Link, err)
		}
	}
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: agg <time_between_reqs>")
	}
	time_between_request, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", cmd.Args[0], err)
	}
	fmt.Printf("Collecting feeds every %s\n", time_between_request)

	ticker := time.NewTicker(time_between_request)

	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}

}

func handlerAddfeed(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 2 {
		return fmt.Errorf("usage: addfeed <Name> <Url>")
	}
	user, err := s.db.GetUser(context.Background(), user.Name)
	if err != nil {
		return fmt.Errorf("get current user: %w", err)
	}
	name := cmd.Args[0]
	url := cmd.Args[1]

	dbParams := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      name,
		Url:       url,
		UserID:    user.ID,
	}

	feed, err := s.db.CreateFeed(context.Background(), dbParams)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	follow, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	fmt.Printf("%s is now following %s\n", follow.UserName, follow.FeedName)

	fmt.Printf("%+v\n", feed)
	return nil
}

func handlerFeeds(s *state, _ command) error {
	feed, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	for _, f := range feed {
		fmt.Printf("%s\n%s\n%s\n", f.FeedName, f.FeedUrl, f.UserName)

	}
	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: follow <url>")
	}

	user, err := s.db.GetUser(context.Background(), user.Name)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	url := cmd.Args[0]
	feed, err := s.db.GetFeed(context.Background(), url)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("no feed with url %s — add it first", url)
		}
		return fmt.Errorf("get feed: %w", err)
	}

	dbParams := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	}

	follow, err := s.db.CreateFeedFollow(context.Background(), dbParams)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	fmt.Printf("%s is now following %s\n", follow.UserName, follow.FeedName)
	return nil
}

func handlerFollowing(s *state, _ command, user database.User) error {
	follows, err := s.db.GetFeedFollowsForUser(context.Background(), user.Name)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	for _, f := range follows {
		fmt.Println(f.FeedName)
	}
	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: unfollow <url>")
	}

	dbParams := database.DeleteFeedFollowParams{
		UserID: user.ID,
		Url:    cmd.Args[0],
	}

	if err := s.db.DeleteFeedFollow(context.Background(), dbParams); err != nil {
		return fmt.Errorf("unfollow: %w", err)
	}

	fmt.Printf("unfollowed %s\n", cmd.Args[0])
	return nil
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	limit := 2
	if len(cmd.Args) == 1 {
		n, err := strconv.Atoi(cmd.Args[0])
		if err != nil {
			return fmt.Errorf("invalid limit %q: %w", cmd.Args[0], err)
		}
		limit = n
	}
	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
	})
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	for _, post := range posts {
		if post.PublishedAt.Valid {
			fmt.Printf("  %s\n", post.PublishedAt.Time.Format("2006-01-02"))
		}
		fmt.Printf("%s\n", post.Title)
		fmt.Printf("  %s\n", post.Url)
		fmt.Println()
	}
	return nil
}
