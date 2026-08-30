package main

import (
	"database/sql"
	"log"
	"os"

	cfg "github.com/Ottbart/gator/internal/config"
	database "github.com/Ottbart/gator/internal/database"
	_ "github.com/lib/pq"
)

type State struct {
	Config *cfg.Config
	db     *database.Queries
}

func main() {
	// read config
	appConfig, err := cfg.ReadConfig()
	if err != nil {
		log.Fatalf("error reading config: %v", err)
		return
	}

	//open a connection to the database...
	db, err := sql.Open("postgres", appConfig.DBURL)
	if err != nil {
		log.Fatal("error opening sql connection")
	}
	defer db.Close()

	//...and create a new database
	dbQueries := database.New(db)

	//??
	programState := &State{
		Config: &appConfig,
		db:     dbQueries,
	}

	//register commands
	cmds := Commands{
		handlers: make(map[string]func(*State, Command) error),
	}
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerDelete)
	cmds.register("users", handlerListUsers)
	cmds.register("agg", handlerAgg)
	cmds.register("addfeed", handlerAddFeed)

	//read inputs from the console
	input := os.Args
	if len(input) < 2 {
		log.Fatalf("no command found")
	}

	cmd := Command{
		name: input[1],
		args: input[2:],
	}
	err = cmds.run(programState, cmd)
	if err != nil {
		log.Fatal(err)
	}
}
