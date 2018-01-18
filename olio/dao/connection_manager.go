package dao

import (
	"errors"
	"os"
	"time"

	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"

	"github.com/j0ni/service-skeleton-go/olio/extractors"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

type ConnectionProvider interface {
	GetDb() *gorm.DB
}

type ConnectionManager struct {
	db *gorm.DB
}

func init() {
	gorm.NowFunc = func() time.Time {
		return time.Now().UTC()
	}
}

func NewGormProvider(db *gorm.DB) *ConnectionManager {
	connectionManager := ConnectionManager{}
	connectionManager.db = db
	return &connectionManager
}

func (self *ConnectionManager) createDb(dbDialect string, dbConnectionString string) (*gorm.DB, error) {
	db, err := gorm.Open(dbDialect, dbConnectionString)
	if err != nil {
		return nil, errors.New("failed to connect database")
	}

	env := os.Getenv("GIN_ENV")
	if env == "development" || env == "test" {
		db.LogMode(true)
	}

	return db, nil
}

func (self *ConnectionManager) GetDb() *gorm.DB {
	return self.db
}

func (self *ConnectionManager) Ping() error {
	return self.db.DB().Ping()
}

func (self *ConnectionManager) Close() error {
	return self.db.DB().Close()
}

func NewConnectionManager(dbExtractor extractors.DbExtractor) (*ConnectionManager, error) {
	connectionManager := ConnectionManager{}

	dbConnectionString := dbExtractor.ExtractConnectionString()
	dialect := dbExtractor.ExtractDialect()

	log.Info("Connecting to [", dbConnectionString, "], a [", dialect, "] database")
	db, err := connectionManager.createDb(string(dialect), dbConnectionString)
	if err != nil {
		return nil, err
	}
	connectionManager.db = db
	return &connectionManager, nil
}
