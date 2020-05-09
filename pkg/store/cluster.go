package store

import (
	"context"
	"database/sql"
	"time"
)

type Cluster struct {
	DB *sql.DB
}

type ClusterRecord struct {
	ID             int
	Name           string
	ExpirationDate time.Time
	Ignore         bool
}

func (c *Cluster) Insert(ctx context.Context, name string, expirationDate time.Time, ignore bool) error {
	statement, err := c.DB.Prepare(`
		INSERT INTO Clusters (Name, ExpirationDate, Ignore)
		VALUES (?, ?, ?)
	`)
	if err != nil {
		return err
	}

	_, err = statement.ExecContext(ctx, name, expirationDate, ignore)
	if err != nil {
		return err
	}

	return nil
}

func (c *Cluster) Delete(ctx context.Context, name string) error {
	statement, err := c.DB.Prepare(`
		DELETE FROM Clusters
		WHERE Name=?
	`)
	if err != nil {
		return err
	}

	_, err = statement.ExecContext(ctx, name)
	if err != nil {
		return err
	}

	return nil
}

func (c *Cluster) List(ctx context.Context) ([]ClusterRecord, error) {
	var clusters []ClusterRecord

	rows, err := c.DB.QueryContext(ctx, `
		SELECT
			ID,
			Name,
			ExpirationDate,
			Ignore
		FROM Clusters`)
	if err != nil {
		return []ClusterRecord{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name string
		var expirationDate time.Time
		var ignore bool

		err = rows.Scan(&id, &name, &expirationDate, &ignore)
		if err != nil {
			return []ClusterRecord{}, err
		}

		clusters = append(clusters, ClusterRecord{
			ID:             id,
			Name:           name,
			ExpirationDate: expirationDate,
			Ignore:         ignore,
		})
	}

	err = rows.Err()
	if err != nil {
		return []ClusterRecord{}, err
	}

	return clusters, nil
}

func (c *Cluster) ListExpired(ctx context.Context) ([]ClusterRecord, error) {
	var clusters []ClusterRecord

	rows, err := c.DB.QueryContext(ctx, `
		SELECT
			ID,
			Name,
			ExpirationDate,
			Ignore
		FROM Clusters
		WHERE ExpirationDate < ?`, time.Now())
	if err != nil {
		return []ClusterRecord{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name string
		var expirationDate time.Time
		var ignore bool

		err = rows.Scan(&id, &name, &expirationDate, &ignore)
		if err != nil {
			return []ClusterRecord{}, err
		}

		clusters = append(clusters, ClusterRecord{
			ID:             id,
			Name:           name,
			ExpirationDate: expirationDate,
			Ignore:         ignore,
		})
	}

	err = rows.Err()
	if err != nil {
		return []ClusterRecord{}, err
	}

	return clusters, nil
}

func (c *Cluster) UpdateIgnore(ctx context.Context, name string, ignore bool) error {
	statement, err := c.DB.Prepare(`
		UPDATE Clusters
		SET Ignore = ?
		WHERE Name = ?
	`)
	if err != nil {
		return err
	}

	_, err = statement.ExecContext(ctx, ignore, name)
	if err != nil {
		return err
	}

	return nil
}

func (c *Cluster) UpdateExpirationDate(ctx context.Context, name string, expirationDate time.Time) error {
	statement, err := c.DB.Prepare(`
		UPDATE Clusters
		SET ExpirationDate = ?
		WHERE Name = ?
	`)
	if err != nil {
		return err
	}

	_, err = statement.ExecContext(ctx, expirationDate, name)
	if err != nil {
		return err
	}

	return nil
}
