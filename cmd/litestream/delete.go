package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/benbjohnson/litestream"
)

// DeleteCommand represents a command to remove everything from a replica.
type DeleteCommand struct{}

// Run executes the command.
func (c *DeleteCommand) Run(ctx context.Context, args []string) (err error) {
	fs := flag.NewFlagSet("litestream-delete", flag.ContinueOnError)
	configPath, noExpandEnv := registerConfigFlag(fs)
	replicaName := fs.String("replica", "", "replica name")
	generationName := fs.String("generation", "", "generation name")
	allFiles := fs.Bool("all-files", false, "remove all files")
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return err
	} else if fs.NArg() == 0 || fs.Arg(0) == "" {
		return fmt.Errorf("database path or replica URL required")
	} else if fs.NArg() > 1 {
		return fmt.Errorf("too many arguments")
	}

	if *generationName == "" && !*allFiles {
		return fmt.Errorf("either generation name or -all-files must be specified")
	}

	var db *litestream.DB
	var r *litestream.Replica
	if isURL(fs.Arg(0)) {
		if *configPath != "" {
			return fmt.Errorf("cannot specify a replica URL and the -config flag")
		}
		if r, err = NewReplicaFromConfig(&ReplicaConfig{URL: fs.Arg(0)}, nil); err != nil {
			return err
		}
	} else {
		if *configPath == "" {
			*configPath = DefaultConfigPath()
		}

		// Load configuration.
		config, err := ReadConfigFile(*configPath, !*noExpandEnv)
		if err != nil {
			return err
		}

		// Lookup database from configuration file by path.
		if path, err := expand(fs.Arg(0)); err != nil {
			return err
		} else if dbc := config.DBConfig(path); dbc == nil {
			return fmt.Errorf("database not found in config: %s", path)
		} else if db, err = NewDBFromConfig(dbc); err != nil {
			return err
		}

		// Filter by replica, if specified.
		if *replicaName != "" {
			if r = db.Replica(*replicaName); r == nil {
				return fmt.Errorf("replica %q not found for database %q", *replicaName, db.Path())
			}
		}
	}

	var replicas []*litestream.Replica
	if r != nil {
		replicas = []*litestream.Replica{r}
	} else {
		replicas = db.Replicas
	}

	// Delete everything from each replica.
	for _, r := range replicas {
		if *allFiles {
			log.Printf("%s: removing all files from replica", r.Name())

			if err := r.Client.DeleteAll(ctx); err != nil {
				return err
			}
		} else if *generationName != "" {
			log.Printf("%s: removing generation %s", r.Name(), *generationName)

			if err := r.Client.DeleteGeneration(ctx, *generationName); err != nil {
				return err
			}
		}
	}

	return nil
}

// Usage prints the help message to STDOUT.
func (c *DeleteCommand) Usage() {
	fmt.Printf(`
The delete command removes data from replicas. Either a generation or -all-files
must be specified.

Usage:

	litestream delete [arguments] DB_PATH
	
	litestream delete [arguments] REPLICA_URL

Arguments:

	-config PATH
	    Specifies the configuration file.
	    Defaults to %s

	-no-expand-env
	    Disables environment variable expansion in configuration file.

	-replica NAME
	    Optional, filters by replica.

	-generation NAME
	    Optional, selects a generation.

	-all-files
	    Optional, removes everything on replica path recursively.

`[1:],
		DefaultConfigPath(),
	)
}
