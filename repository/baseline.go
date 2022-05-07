package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Leantar/fimserver/models"
	"github.com/jmoiron/sqlx"
)

type PgBaselineRepository struct {
	db *sqlx.DB
}

func (f *PgBaselineRepository) CreateMany(ctx context.Context, wfs []models.FsObject) (err error) {
	const query = "INSERT INTO baseline_fs_objects(path, hash, created, modified, uid, gid, mode, fk_agent_id) VALUES(:path,:hash,:created,:modified,:uid,:gid,:mode,:agentid)"

	wfsLen := len(wfs)
	step := 5000
	lowerBound := 0
	upperBound := getMinimum(step, wfsLen)

	for lowerBound < wfsLen {
		_, err = f.db.NamedExecContext(ctx, query, wfs[lowerBound:upperBound])
		if err != nil {
			return
		}

		lowerBound = upperBound
		upperBound = getMinimum(upperBound+step, wfsLen)
	}

	return
}

func (f *PgBaselineRepository) GetByPathAndAgentID(ctx context.Context, path string, agentID uint64) (models.FsObject, error) {
	const query = "SELECT * FROM baseline_fs_objects WHERE path = $1 AND fk_agent_id = $2 LIMIT 1"
	var fsObject dbFsObject

	err := f.db.GetContext(ctx, &fsObject, query, path, agentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.FsObject{}, errEmptyResultSet
		}
		return models.FsObject{}, err
	}

	return models.FsObject(fsObject), nil
}

func (f *PgBaselineRepository) GetBaselineByAgent(ctx context.Context, agentID uint64) ([]models.FsObject, error) {
	const query = "SELECT * FROM baseline_fs_objects  WHERE fk_agent_id = $1 ORDER BY id ASC"
	objs := make(dbFsObjects, 0)

	err := f.db.SelectContext(ctx, &objs, query, agentID)
	if err != nil {
		return nil, err
	}

	if len(objs) == 0 {
		return nil, errEmptyResultSet
	}

	return objs.toFsObjects(), nil
}

func (f *PgBaselineRepository) DeleteBaselineForAgent(ctx context.Context, agentID uint64) (err error) {
	const query = "DELETE FROM baseline_fs_objects WHERE fk_agent_id = $1"

	_, err = f.db.ExecContext(ctx, query, agentID)

	return
}
