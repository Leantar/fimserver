package repository

import (
	"context"
	"database/sql"

	"github.com/Leantar/fimserver/models"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type PgEndpointRepository struct {
	db *sqlx.DB
}

func (e *PgEndpointRepository) Create(ctx context.Context, ep models.Endpoint) (err error) {
	const query = "INSERT INTO endpoints(name, kind, roles, has_baseline, baseline_is_current, watched_paths) VALUES($1,$2,$3,$4,$5,$6)"

	_, err = e.db.ExecContext(ctx, query, ep.Name, ep.Kind, pq.Array(ep.Roles), ep.HasBaseline, ep.BaselineIsCurrent, pq.Array(ep.WatchedPaths))

	return
}

func (e *PgEndpointRepository) GetByName(ctx context.Context, name string) (models.Endpoint, error) {
	const query = "SELECT * FROM endpoints WHERE name = $1 LIMIT 1"
	var endpoint dbEndpoint

	err := e.db.GetContext(ctx, &endpoint, query, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Endpoint{}, errEmptyResultSet
		}
		return models.Endpoint{}, err
	}

	return endpoint.toEndpoint(), nil
}

func (e *PgEndpointRepository) GetAgents(ctx context.Context) ([]models.Endpoint, error) {
	const query = "SELECT * FROM endpoints WHERE kind = 'agent'"
	endpoints := make(dbEndpoints, 0)

	err := e.db.SelectContext(ctx, &endpoints, query)
	if err != nil {
		return nil, err
	}

	if len(endpoints) == 0 {
		return nil, errEmptyResultSet
	}

	return endpoints.toEndpoints(), nil
}

func (e *PgEndpointRepository) Update(ctx context.Context, ep models.Endpoint) (err error) {
	const query = "UPDATE endpoints SET name = $1, kind = $2, roles = $3, has_baseline = $4, baseline_is_current = $5, watched_paths = $6 WHERE id = $7"

	_, err = e.db.ExecContext(ctx, query, ep.Name, ep.Kind, pq.Array(ep.Roles), ep.HasBaseline, ep.BaselineIsCurrent, pq.Array(ep.WatchedPaths), ep.ID)

	return
}

func (e *PgEndpointRepository) Delete(ctx context.Context, name string) (err error) {
	const query = "DELETE FROM endpoints WHERE name = $1"

	_, err = e.db.ExecContext(ctx, query, name)

	return
}
