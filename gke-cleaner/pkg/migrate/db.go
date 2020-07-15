package migrate

import (
	"database/sql"
	"os"

	"github.com/go-logr/logr"
)

type DB struct {
	Log logr.Logger
	DB  *sql.DB
}

func (d *DB) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	statement, err := d.DB.Prepare(`CREATE TABLE IF NOT EXISTS Clusters (
		ID INT NOT NULL AUTO_INCREMENT,
		Name TEXT NOT NULL,
		ExpirationDate DATETIME,
		IgnoreMe BOOLEAN,
		PRIMARY KEY (ID)
	)`)
	if err != nil {
		return err
	}
	_, err = statement.Exec()
	if err != nil {
		return err
	}

	close(ready)
	d.Log.Info("Migrated database")

	select {
	case <-signals:
		return nil
	}
}
