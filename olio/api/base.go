package api

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/j0ni/service-skeleton-go/olio/dao"
	"github.com/j0ni/service-skeleton-go/olio/db"
)

type OlioBaseCoreAPI struct {
	ConnectionManager *dao.ConnectionManager
}

func (api *OlioBaseCoreAPI) Init() {
	log.Info("Initializing core api.")
}

func (api *OlioBaseCoreAPI) RunMigrations(migrations []db.Migration, tableName ...string) {
	if len(tableName) > 1 {
		log.Fatalf("Wrong number of args (%d), function takes 1 or 2 args", 1+len(tableName))
	}

	var migrationManager db.MigrationManager
	if len(tableName) == 1 {
		migrationManager = db.NewMigrationManager(api.ConnectionManager, migrations, tableName[0])
	} else {
		migrationManager = db.NewMigrationManager(api.ConnectionManager, migrations)
	}

	if err := migrationManager.Migrate(); err != nil {
		log.Fatal("Failed to run migrations: ", err)
		os.Exit(1)
	}
}
