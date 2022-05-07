package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Leantar/fimserver/models"
	"github.com/jmoiron/sqlx"
)

type PgAlertRepository struct {
	db *sqlx.DB
}

func (a *PgAlertRepository) Create(ctx context.Context, al models.Alert) (err error) {
	const query = "INSERT INTO alerts(kind, difference, issued_at, path, modified, fk_agent_id) VALUES($1,$2,$3,$4,$5,$6)"

	_, err = a.db.ExecContext(ctx, query, al.Kind, al.Difference, al.IssuedAt, al.Path, al.Modified, al.AgentID)

	return
}

func (a *PgAlertRepository) GetAllByAgent(ctx context.Context, agentID uint64) ([]models.Alert, error) {
	const query = "SELECT * from alerts WHERE fk_agent_id = $1"
	alerts := make(dbAlerts, 0)

	err := a.db.SelectContext(ctx, &alerts, query, agentID)
	if err != nil {
		return nil, err
	}

	if len(alerts) == 0 {
		return nil, errEmptyResultSet
	}

	return alerts.toAlerts(), nil
}

func (a *PgAlertRepository) GetLatestByPathAndAgent(ctx context.Context, path string, agentID uint64) (models.Alert, error) {
	const query = "SELECT * from alerts WHERE path = $1 AND fk_agent_id = $2 ORDER BY issued_at DESC LIMIT 1"
	var alert dbAlert

	err := a.db.GetContext(ctx, &alert, query, path, agentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Alert{}, errEmptyResultSet
		}
		return models.Alert{}, err
	}

	return alert.toAlert(), nil
}

func (a *PgAlertRepository) DeleteAll(ctx context.Context, agentID uint64) (err error) {
	const query = "DELETE FROM alerts WHERE fk_agent_id = $1"

	_, err = a.db.ExecContext(ctx, query, agentID)

	return
}
