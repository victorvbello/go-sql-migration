package main

import (
	"database/sql"
	"flag"
	"log"
	"os"

	"github.com/victorvbello/go-sql-migration/migration"
	_ "github.com/victorvbello/go-sql-migration/migrations"
)

func SQLOpenConnection(connStr string) (*sql.DB, error) {
	return sql.Open("mysql", connStr)
}

func main() {
	var command, templateName, templateDir, cnnString string
	var versionTo int
	// create command
	createFlags := flag.NewFlagSet("create", flag.ExitOnError)
	createFlags.StringVar(&templateName, "name", "", "Name of template file")
	createFlags.StringVar(&templateDir, "dir", "./migrations", "Migration destination directory")

	// up command
	upFlags := flag.NewFlagSet("up", flag.ExitOnError)
	upFlags.StringVar(&cnnString, "conn", "", "String connection")

	// up-to command
	upToFlags := flag.NewFlagSet("up-to", flag.ExitOnError)
	upToFlags.StringVar(&cnnString, "conn", "", "String connection")
	upToFlags.IntVar(&versionTo, "version", 0, "Version to")

	// down command
	downFlags := flag.NewFlagSet("down", flag.ExitOnError)
	downFlags.StringVar(&cnnString, "conn", "", "String connection")

	// down-to command
	downToFlags := flag.NewFlagSet("down-to", flag.ExitOnError)
	downToFlags.StringVar(&cnnString, "conn", "", "String connection")
	downToFlags.IntVar(&versionTo, "version", 0, "Version to")

	if len(os.Args[1:]) < 1 {
		log.Fatal("You must pass a command")
	}

	command = os.Args[1]

	switch command {
	case "create":
		createFlags.Parse(os.Args[2:])
		if templateName == "" {
			log.Fatal("template name is required")
		}
		if templateDir == "" {
			log.Fatal("template dir is required")
		}
		err := migration.CreateMigrationFile(templateName, templateDir)
		if err != nil {
			log.Fatalf("error on create template [%s] in dir [%s] error: %v", templateName, templateDir, err)
		}
	case "up":
		upFlags.Parse(os.Args[2:])
		if cnnString == "" {
			log.Fatal("String connection is required")
		}
		db, err := SQLOpenConnection(cnnString)
		if err != nil {
			log.Fatalf("error on SQLOpenConnection: %v", err)
		}
		err = migration.RunUp(db)
		if err != nil {
			log.Fatalf("error on run action up error: %v", err)
		}
	case "up-to":
		upToFlags.Parse(os.Args[2:])
		if cnnString == "" {
			log.Fatal("String connection is required")
		}
		if versionTo == 0 {
			log.Fatal("Version to is required")
		}
		db, err := SQLOpenConnection(cnnString)
		if err != nil {
			log.Fatalf("error on SQLOpenConnection: %v", err)
		}
		err = migration.RunUpTo(db, int64(versionTo))
		if err != nil {
			log.Fatalf("error on run action up-to error: %v", err)
		}
	case "down":
		downFlags.Parse(os.Args[2:])
		if cnnString == "" {
			log.Fatal("String connection is required")
		}
		db, err := SQLOpenConnection(cnnString)
		if err != nil {
			log.Fatalf("error on SQLOpenConnection: %v", err)
		}
		err = migration.RunDown(db)
		if err != nil {
			log.Fatalf("error on run action down error: %v", err)
		}
	case "down-to":
		downToFlags.Parse(os.Args[2:])
		if cnnString == "" {
			log.Fatal("String connection is required")
		}
		if versionTo == 0 {
			log.Fatal("Version to is required")
		}
		db, err := SQLOpenConnection(cnnString)
		if err != nil {
			log.Fatalf("error on SQLOpenConnection: %v", err)
		}
		err = migration.RunDownTo(db, int64(versionTo))
		if err != nil {
			log.Fatalf("error on run action down-to error: %v", err)
		}
	default:
		log.Fatalf("invalid command %s", command)
	}
}
