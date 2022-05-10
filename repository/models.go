package repository

import (
	"github.com/Leantar/fimserver/models"
	"github.com/Leantar/fimserver/modules/casbin"
	"github.com/lib/pq"
)

type dbEndpoint struct {
	ID                uint64         `db:"id"`
	Name              string         `db:"name"`
	Kind              string         `db:"kind"`
	Roles             pq.StringArray `db:"roles"`
	HasBaseline       bool           `db:"has_baseline"`
	BaselineIsCurrent bool           `db:"baseline_is_current"`
	WatchedPaths      pq.StringArray `db:"watched_paths"`
}

func (d dbEndpoint) toEndpoint() models.Endpoint {
	return models.Endpoint{
		ID:                d.ID,
		Name:              d.Name,
		Kind:              d.Kind,
		Roles:             d.Roles,
		HasBaseline:       d.HasBaseline,
		BaselineIsCurrent: d.BaselineIsCurrent,
		WatchedPaths:      d.WatchedPaths,
	}
}

type dbEndpoints []dbEndpoint

func (d dbEndpoints) toEndpoints() []models.Endpoint {
	conv := make([]models.Endpoint, len(d))
	for i, ep := range d {
		conv[i] = ep.toEndpoint()
	}

	return conv
}

type dbAlert struct {
	ID         uint64 `db:"id"`
	Kind       string `db:"kind"`
	Difference string `db:"difference"`
	IssuedAt   int64  `db:"issued_at"`
	Path       string `db:"path"`
	Modified   int64  `db:"modified"`
	AgentID    uint64 `db:"fk_agent_id"`
}

func (d dbAlert) toAlert() models.Alert {
	return models.Alert(d)
}

type dbAlerts []dbAlert

func (d dbAlerts) toAlerts() []models.Alert {
	conv := make([]models.Alert, len(d))
	for i, al := range d {
		conv[i] = al.toAlert()
	}

	return conv
}

type dbFsObject struct {
	ID       uint64 `db:"id"`
	Path     string `db:"path"`
	Hash     string `db:"hash"`
	Created  int64  `db:"created"`
	Modified int64  `db:"modified"`
	Uid      uint32 `db:"uid"`
	Gid      uint32 `db:"gid"`
	Mode     uint32 `db:"mode"`
	AgentID  uint64 `db:"fk_agent_id"`
}

func (d dbFsObject) toFsObject() models.FsObject {
	return models.FsObject(d)
}

type dbFsObjects []dbFsObject

func (d dbFsObjects) toFsObjects() []models.FsObject {
	conv := make([]models.FsObject, len(d))
	for i, obj := range d {
		conv[i] = obj.toFsObject()
	}

	return conv
}

type dbRule struct {
	ID    uint64 `db:"id"`
	PType string `db:"p_type"`
	V0    string `db:"v0"`
	V1    string `db:"v1"`
	V2    string `db:"v2"`
	V3    string `db:"v3"`
	V4    string `db:"v4"`
	V5    string `db:"v5"`
}

func (d *dbRule) toRule() casbin.Rule {
	return casbin.Rule{
		PType: d.PType,
		V0:    d.V0,
		V1:    d.V1,
		V2:    d.V2,
		V3:    d.V3,
		V4:    d.V4,
		V5:    d.V5,
	}
}

type dbRules []dbRule

func (d dbRules) toRules() []casbin.Rule {
	conv := make([]casbin.Rule, len(d))
	for i, rule := range d {
		conv[i] = rule.toRule()
	}

	return conv
}
