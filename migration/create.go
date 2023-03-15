package migration

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"time"

	"github.com/victorvbello/go-sql-migration/helpers"
)

const fileTemplate = `package migrations

import (
	"database/sql"
	"log"
	SqlMigration "github.com/victorvbello/go-sql-migration/migration"
)

/*
File create by sql-migration flow
Version: {{.Version}}
Name: {{.OriginalName}}
*/

func init() {
	SqlMigration.AddMigration(upAction{{.Name}}, downAction{{.Name}})
}

func upAction{{.Name}}(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	log.Println("\t\tAction up of {{.Name}}")
	return nil
}

func downAction{{.Name}}(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	log.Println("\t\tAction down of {{.Name}}")
	return nil
}
`

func CreateMigrationFile(name string, migrationDir string) error {
	version := time.Now().Format(timeFormat)

	finalFilename := fmt.Sprintf("%s_%s.%s", version, helpers.StrSnakeCase(name), migrationFileType)

	path := filepath.Join(migrationDir, finalFilename)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return fmt.Errorf("[CreateMigrationFile] failed to create migration file: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("[CreateMigrationFile] failed to create migration file: %w", err)
	}
	defer f.Close()

	vars := map[string]interface{}{
		"Version":      version,
		"OriginalName": name,
		"Name":         helpers.StrCamelCase(name, ""),
	}

	migrationTemplate, err := template.New("migration-file-template").Parse(fileTemplate)

	if err != nil {
		return fmt.Errorf("[CreateMigrationFile] failed to create migration file: %w", err)
	}

	if err := migrationTemplate.Execute(f, vars); err != nil {
		return fmt.Errorf("[CreateMigrationFile] failed to execute template: %w", err)
	}
	return nil
}
