# Gator

Gator is a command-line RSS feed aggregator backed by PostgreSQL.

## Requirements

Install the following before running Gator:

- [Go](https://go.dev/dl/)
- [PostgreSQL](https://www.postgresql.org/download/)
- [Goose](https://github.com/pressly/goose), for applying the database migrations

## Database Setup

Create a PostgreSQL database named `gator`:

```sh
createdb gator
```

From the repository root, apply the schema migrations. The connection string should match your local PostgreSQL setup:

```sh
cd sql/schema
goose postgres "postgres://postgres:postgres@localhost:5432/gator" up
cd ../..
```

If your PostgreSQL username, password, host, port, or database name differs, update the connection string accordingly.

## Configuration

Gator reads its configuration from `~/.gatorconfig.json`. Create that file with your PostgreSQL connection URL and an empty current user name:

```json
{
  "db_url": "postgres://postgres:postgres@localhost:5432/gator",
  "current_user_name": ""
}
```

Set `current_user_name` to the name of a registered user after registration, or let the `register` and `login` commands update it for you.

## Install

Install the Gator CLI with [`go install`](https://pkg.go.dev/cmd/go#hdr-Compile_and_install_packages_and_dependencies):

```sh
go install github.com/Ottbart/gator@latest
```

Make sure Go's install directory is on your `PATH`. You can also run the program directly from the repository with:

```sh
go run . <command> [arguments]
```

## Commands

Register a user and switch the current user:

```sh
gator register alice
```

Log in as an existing user and list all users:

```sh
gator login alice
gator users
```

Add and list RSS feeds:

```sh
gator addfeed "Hacker News" "https://news.ycombinator.com/rss"
gator feeds
```

Follow or unfollow a feed, then list the feeds followed by the current user:

```sh
gator follow "https://news.ycombinator.com/rss"
gator following
gator unfollow "https://news.ycombinator.com/rss"
```

Start the feed aggregator. The argument is the interval between requests, using Go duration syntax:

```sh
gator agg 1m
```

Browse the current user's posts. The optional argument controls how many posts are returned and defaults to 2:

```sh
gator browse
gator browse 10
```

To delete all users and their related data:

```sh
gator reset
```
