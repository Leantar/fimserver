package repository

import (
	"context"

	"github.com/Leantar/fimserver/modules/casbin"
	"github.com/jmoiron/sqlx"
)

type PgRuleRepository struct {
	db *sqlx.DB
	sqlx.Tx
}

func (r *PgRuleRepository) Create(ctx context.Context, li casbin.Rule) (err error) {
	const query = "INSERT INTO rules(p_type,v0,v1,v2,v3,v4,v5) VALUES($1,$2,$3,$4,$5,$6,$7)"

	_, err = r.db.ExecContext(ctx, query, li.PType, li.V0, li.V1, li.V2, li.V3, li.V4, li.V5)

	return
}

func (r *PgRuleRepository) GetAll(ctx context.Context) ([]casbin.Rule, error) {
	const query = "SELECT * from rules"
	rules := make(dbRules, 0)

	err := r.db.SelectContext(ctx, &rules, query)
	if err != nil {
		return nil, err
	}

	if len(rules) == 0 {
		return nil, errEmptyResultSet
	}

	return rules.toRules(), nil
}

func (r *PgRuleRepository) Delete(ctx context.Context, li casbin.Rule) (err error) {
	const query = "DELETE FROM rules WHERE p_type = $1 AND v0 = $2 AND v1 = $3 AND v2 = $4 AND v3 = $5 AND v4 = $6 AND v5 = $7"

	_, err = r.db.ExecContext(ctx, query, li.PType, li.V0, li.V1, li.V2, li.V3, li.V4, li.V5)

	return
}

func (r *PgRuleRepository) DeleteAll(ctx context.Context) (err error) {
	const query = "DELETE FROM rules"

	_, err = r.db.ExecContext(ctx, query)

	return
}
