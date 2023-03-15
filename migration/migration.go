package migration

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type MigrationAction func(tx *sql.Tx) error

type migration struct {
	fileName   string
	version    int64
	upAction   MigrationAction
	downAction MigrationAction
}

type migrations []*migration

const (
	_MIGRATION_TABLE = "migration_db_version"
)

var (
	migrationFilesAdds = map[int64]*migration{}
	migrationAction    = struct {
		UP   string
		DOWN string
	}{
		"UP",
		"DOWN",
	}
)

func (m migrations) Len() int      { return len(m) }
func (m migrations) Swap(i, j int) { m[i], m[j] = m[j], m[i] }
func (m migrations) Less(i, j int) bool {
	if m[i].version == m[j].version {
		log.Fatalf("[migrations.Less] duplicate version %s detected:\n%s\n%s", m[i].version, m[i].fileName, m[j].fileName)
	}
	return m[i].version < m[j].version
}

func extractVersionNumberFromFileName(path string) (int64, error) {
	fileBasePath := filepath.Base(path)

	underScoreIndex := strings.Index(fileBasePath, "_")
	if underScoreIndex < 0 {
		return 0, errors.New("[extractVersionNumberFromFileName] separator '_' not found")
	}

	version, err := strconv.ParseInt(fileBasePath[:underScoreIndex], 10, 64)
	if err == nil && version <= 0 {
		return 0, errors.New("[extractVersionNumberFromFileName] migration version must be greater than zero")
	}

	return version, err
}

func prepareMigrationList() (migrations, error) {
	var migrationList migrations
	if len(migrationFilesAdds) < 0 {
		return migrationList, errors.New("[prepareMigrationList] migration file list is empty")
	}
	for _, m := range migrationFilesAdds {
		migrationList = append(migrationList, m)
	}
	sort.Sort(migrationList)
	return migrationList, nil
}

func createMigrationTable(db *sql.DB) error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id INT NOT NULL AUTO_INCREMENT,
			version_id bigint NOT NULL,
			created timestamp NULL default now(),
			UNIQUE KEY version_id (version_id) USING BTREE,
			PRIMARY KEY(id)
		)`,
		_MIGRATION_TABLE)
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("[createMigrationTable] error :%v", err)
	}
	return nil
}

func insertMigration(tx *sql.Tx, version int64) error {
	query := fmt.Sprintf(`INSERT INTO %s (version_id) VALUES (?)`, _MIGRATION_TABLE)
	_, err := tx.Exec(query, version)
	if err != nil {
		return fmt.Errorf("[insertMigration] error :%v", err)
	}
	return nil
}

func deleteMigration(tx *sql.Tx, version int64) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE version_id = ?`, _MIGRATION_TABLE)
	_, err := tx.Exec(query, version)
	if err != nil {
		return fmt.Errorf("[deleteMigration] error :%v", err)
	}
	return nil
}

func listMigrateVersionsFrom(db *sql.DB, versionTo int64) ([]int64, error) {
	var versionList []int64
	query := fmt.Sprintf(`SELECT version_id FROM %s WHERE version_id > ? ORDER BY version_id ASC`, _MIGRATION_TABLE)
	rows, err := db.Query(query, versionTo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var version int64
		err = rows.Scan(
			&version,
		)
		if err != nil {
			return nil, fmt.Errorf("[listMigrateVersionsFrom] error :%v", err)
		}
		versionList = append(versionList, version)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("[listMigrateVersionsFrom] error :%v", err)
	}

	return versionList, nil
}

func existMigrate(tx *sql.Tx, version int64) (bool, error) {
	var result int
	var exist bool
	query := fmt.Sprintf(`SELECT 1 FROM %s WHERE version_id = ? LIMIT 1`, _MIGRATION_TABLE)
	row := tx.QueryRow(query, version)
	err := row.Scan(
		&result,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return exist, nil
		}
		return exist, fmt.Errorf("[existMigrate] error :%v", err)
	}
	exist = result == 1
	return exist, nil
}

func runAction(db *sql.DB, version int64, action MigrationAction, actionType string) error {
	var err error

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("[runAction] error %v", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
		tx.Commit()
	}()

	switch actionType {
	case migrationAction.UP:
		migrateApplied, err := existMigrate(tx, version)
		if err != nil {
			return fmt.Errorf("[runAction] error %v", err)
		}
		if migrateApplied {
			log.Printf("\t\tMigration %d applied\n", version)
			return nil
		}
		err = insertMigration(tx, version)
	case migrationAction.DOWN:
		err = deleteMigration(tx, version)
	}

	if err != nil {
		return fmt.Errorf("[runAction] error %v", err)
	}

	err = action(tx)
	if err != nil {
		return fmt.Errorf("[runAction] error action %v", err)
	}
	return nil
}

func runUp(db *sql.DB, m *migration) error {
	log.Printf("\t- %s\n", m.fileName)
	err := runAction(db, m.version, m.upAction, migrationAction.UP)
	if err != nil {
		return fmt.Errorf("[runUp] error %v", err)
	}
	return nil
}

func runDown(db *sql.DB, m *migration) error {
	log.Printf("\t- %s\n", m.fileName)
	err := runAction(db, m.version, m.downAction, migrationAction.DOWN)
	if err != nil {
		return fmt.Errorf("[runDown] runAction error %v", err)
	}
	return nil
}

func AddMigration(upAction MigrationAction, downAction MigrationAction) error {
	_, filename, _, _ := runtime.Caller(1)

	currentVersionFile, err := extractVersionNumberFromFileName(filename)
	if err != nil {
		return fmt.Errorf("[AddMigration] error on get current version file %v", err)
	}

	if existing, ok := migrationFilesAdds[currentVersionFile]; ok {
		return fmt.Errorf("[AddMigration] error to add migration [%s] version [%d] conflicts with [%s]",
			filename,
			currentVersionFile,
			existing.fileName,
		)
	}

	migrationFilesAdds[currentVersionFile] = &migration{
		fileName:   filename,
		version:    currentVersionFile,
		upAction:   upAction,
		downAction: downAction,
	}
	return nil
}

func RunUp(db *sql.DB) error {
	err := createMigrationTable(db)
	if err != nil {
		return fmt.Errorf("[RunUp] error %v", err)
	}

	migrationList, err := prepareMigrationList()
	if err != nil {
		return fmt.Errorf("[RunUp] error %v", err)
	}
	if len(migrationList) == 0 {
		return errors.New("[RunUp] migration list is empty")
	}

	log.Println("Run up migrations")
	for _, m := range migrationList {
		err := runUp(db, m)
		if err != nil {
			return fmt.Errorf("[RunUp] error %v", err)
		}
	}
	return nil
}

func RunDown(db *sql.DB) error {
	err := createMigrationTable(db)
	if err != nil {
		return fmt.Errorf("[RunDown] error %v", err)
	}

	migrationList, err := prepareMigrationList()
	if err != nil {
		return fmt.Errorf("[RunUp] error %v", err)
	}
	if len(migrationList) == 0 {
		return errors.New("[RunDown] migration list is empty")
	}

	log.Println("Run down migrations")
	for _, m := range migrationList {
		err := runDown(db, m)
		if err != nil {
			return fmt.Errorf("[RunDown] error %v", err)
		}
	}
	return nil
}

func RunUpTo(db *sql.DB, versionTo int64) error {
	err := createMigrationTable(db)
	if err != nil {
		return fmt.Errorf("[RunUpTo] error %v", err)
	}
	m, ok := migrationFilesAdds[versionTo]
	if !ok {
		return errors.New("[RunUpTo] version %d not found")
	}
	log.Printf("Run up migration %d\n", versionTo)
	err = runUp(db, m)
	if err != nil {
		return fmt.Errorf("[RunUpTo] error %v", err)
	}
	return nil
}

func RunDownTo(db *sql.DB, versionTo int64) error {
	err := createMigrationTable(db)
	if err != nil {
		return fmt.Errorf("[RunDownTo] error %v", err)
	}

	versionList, err := listMigrateVersionsFrom(db, versionTo)
	if err != nil {
		return fmt.Errorf("[RunDownTo] error %v", err)
	}
	if len(versionList) == 0 {
		return errors.New("[RunDownTo] version list is empty")
	}

	for _, v := range versionList {
		m, ok := migrationFilesAdds[v]
		if !ok {
			return errors.New("[RunDownTo] version %d not found")
		}
		err = runDown(db, m)
		if err != nil {
			return fmt.Errorf("[RunDownTo] error %v", err)
		}
	}
	return nil
}
