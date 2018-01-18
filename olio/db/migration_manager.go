package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/j0ni/service-skeleton-go/olio/dao"
	"github.com/j0ni/service-skeleton-go/olio/util"
	log "github.com/sirupsen/logrus"
)

type MigrationManager struct {
	connectionManager *dao.ConnectionManager
	migrations        []Migration
	tableName         string
}

type Migration func() error

func (self *MigrationManager) getRequiredSchemaVersion() int {
	return len(self.migrations)
}

func (self *MigrationManager) getCurrentVersion() (int, error) {
	db := self.connectionManager.GetDb()
	var version int
	row := db.Table(self.tableName).Select("version").Row()
	row.Scan(&version)
	return version, db.Error
}

func (self *MigrationManager) perequisites() error {
	db := self.connectionManager.GetDb()
	sqlStr := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (version int)", self.tableName)
	if db.Exec(sqlStr).Error != nil {
		return db.Error
	}

	rows, error := db.Table(self.tableName).Select("version").Rows()
	if error != nil {
		return error
	}
	if rows == nil || rows.Err() == sql.ErrNoRows || !rows.Next() {
		sqlStr = fmt.Sprintf("INSERT INTO %s values (?)", self.tableName)
		return db.Exec(sqlStr, 0).Error
	}
	rows.Close()
	return nil
}

func (self *MigrationManager) incrementVersion(targetVersion int) error {
	db := self.connectionManager.GetDb()
	sqlStr := fmt.Sprintf("UPDATE %s SET version = ?", self.tableName)
	return db.Exec(sqlStr, targetVersion).Error
}

func (self *MigrationManager) Migrate() error {
	err := self.perequisites()
	if err != nil {
		return err
	}

	currentVersion, err := self.getCurrentVersion()
	if err != nil {
		return err
	}

	requiredVersion := self.getRequiredSchemaVersion()
	log.Info("Required schema version is ", requiredVersion)
	log.Info("Database is at schema version ", currentVersion)

	if currentVersion == requiredVersion {
		log.Info("Database schema is up to date.")
		return nil
	}

	if currentVersion > requiredVersion {
		return errors.New("Current schema cannot be ahead of code!")
	}

	targetVersion := currentVersion + 1

	if targetVersion > len(self.migrations) {
		return errors.New("Invalid target version " + util.Int64ToString(int64(targetVersion)) + "; no migration exists.")
	}

	err = self.migrations[targetVersion-1]()

	if err != nil {
		return errors.New("Failed to migrate from " + strconv.Itoa(currentVersion) + " to " + strconv.Itoa(targetVersion) + ": " + err.Error())
	}

	if err := self.incrementVersion(targetVersion); err != nil {
		panic(err)
	}

	return self.Migrate()
}

func NewMigrationManager(connectionManager *dao.ConnectionManager, migrations []Migration, tableName ...string) *MigrationManager {
	if len(tableName) > 1 {
		log.Fatalf("Wrong number of args (%d), function takes 2 or 3 args", 2+len(tableName))
	}

	migrationManager := MigrationManager{}
	migrationManager.connectionManager = connectionManager
	migrationManager.migrations = migrations

	if len(tableName) == 1 {
		migrationManager.tableName = tableName[0]
	}

	return &migrationManager
}
