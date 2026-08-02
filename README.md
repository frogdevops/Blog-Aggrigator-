# gator

A command-line RSS feed aggregator written in Go. Follow blogs and news feeds, let `gator` scrape them on a schedule, and read the collected posts from your terminal.

Posts are stored in Postgres, so the aggregator can run continuously in the background while you browse from another shell.

## Requirements

- **Go 1.22+** — needed to install the CLI
- **PostgreSQL 15+** — the aggregator stores feeds and posts here

Optional, only if you want to work on gator itself:

- [`goose`](https://github.com/pressly/goose) — applies the database migrations
- [`sqlc`](https://sqlc.dev/) — regenerates the typed query layer from SQL

## Installation

```sh
go install github.com/frogdevops/gator@latest
```

This builds a statically linked binary and drops it in `$GOPATH/bin` (usually `~/go/bin`). Make sure that directory is on your `PATH`:

```sh
export PATH="$PATH:$(go env GOPATH)/bin"
```

## Setup

### 1. Start Postgres

Any Postgres instance works. With Docker:

```sh
docker run -d \
  --name gator-db \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=gator \
  -p 5432:5432 \
  -v gator-pgdata:/var/lib/postgresql \
  postgres:17
```

### 2. Run the migrations

From a clone of this repo:

```sh
goose -dir sql/schema postgres "postgres://postgres:postgres@localhost:5432/gator?sslmode=disable" up
```

### 3. Create the config file

`gator` reads its configuration from `~/.gatorconfig.json`:

```json
{
  "db_url": "postgres://postgres:postgres@localhost:5432/gator?sslmode=disable",
  "current_user_name": ""
}
```

`db_url` is the only field you need to set by hand. `current_user_name` is managed by the `login` and `register` commands.

## Usage

### Getting started

```sh
gator register jan          # create a user and log in as them
gator login jan             # switch to an existing user
```

Add a feed and start collecting:

```sh
gator addfeed "Hacker News" https://news.ycombinator.com/rss
gator addfeed "TechCrunch" https://techcrunch.com/feed/
gator agg 1m                # scrape one feed every minute, runs until Ctrl+C
```

Leave `agg` running in its own terminal. In another shell:

```sh
gator browse 10             # show your 10 most recent posts
```

### Commands

| Command | Description |
| --- | --- |
| `register <name>` | Create a new user and log in as them |
| `login <name>` | Set the current user |
| `users` | List all users, marking the current one |
| `addfeed <name> <url>` | Add a feed and follow it |
| `feeds` | List every feed and who added it |
| `follow <url>` | Follow a feed someone else already added |
| `following` | List the feeds the current user follows |
| `unfollow <url>` | Stop following a feed |
| `agg <interval>` | Scrape feeds continuously (e.g. `30s`, `1m`, `1h`) |
| `browse [limit]` | Show recent posts from followed feeds (default 2) |
| `reset` | Delete all users and their data |

### A note on `agg`

`agg` fetches **one** feed per tick, always choosing whichever feed was scraped
longest ago. With five feeds and a `1m` interval, each feed is polled roughly
every five minutes.

Be considerate of the servers you're fetching from — a short interval with only
one or two feeds means a lot of requests to the same host. `1m` is a reasonable
starting point.

## How it works

```
addfeed / follow  ──►  Postgres  ◄──  agg (long-running scraper)
                          │
                       browse
```

`agg` is a writer process; the other commands are readers. They share nothing but
the database, so the scraper can run indefinitely while you use the CLI normally.

Feed scheduling is handled in SQL rather than in Go: the next feed to fetch is
whichever has the oldest `last_fetched_at`, with never-fetched feeds sorting
first. Marking a feed as fetched pushes it to the back of the queue, so the
rotation is a consequence of the ordering rather than logic that has to be
maintained.

## Development

```
sql/schema/     goose migrations
sql/queries/    SQL that sqlc compiles into typed Go
internal/       generated database layer and config handling
```

After changing anything in `sql/`:

```sh
sqlc generate
```

To roll back the most recent migration:

```sh
goose -dir sql/schema postgres "$DB_URL" down
```

## License

MIT