package casbin

import (
	"context"
	"errors"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
)

type RuleRepository interface {
	Create(ctx context.Context, line Rule) error
	GetAll(ctx context.Context) ([]Rule, error)
	Delete(ctx context.Context, line Rule) error
	DeleteAll(ctx context.Context) error
}

type Adapter struct {
	repo RuleRepository
}

type Rule struct {
	PType string
	V0    string
	V1    string
	V2    string
	V3    string
	V4    string
	V5    string
}

func NewAdapter(repo RuleRepository) *Adapter {
	return &Adapter{
		repo: repo,
	}
}

func (a *Adapter) LoadPolicy(model model.Model) error {
	lines, err := a.repo.GetAll(context.Background())
	if err != nil {
		return err
	}

	for _, line := range lines {
		loadPolicyLine(line, model)
	}

	return nil
}

func (a *Adapter) SavePolicy(model model.Model) error {
	err := a.repo.DeleteAll(context.Background())
	if err != nil {
		return err
	}

	for _, char := range "pg" {
		for ptype, ast := range model[string(char)] {
			for _, rule := range ast.Policy {
				line := savePolicyLine(ptype, rule)
				err := a.repo.Create(context.Background(), line)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (a *Adapter) AddPolicy(_ string, ptype string, rule []string) error {
	line := savePolicyLine(ptype, rule)
	err := a.repo.Create(context.Background(), line)
	if err != nil {
		return err
	}

	return nil
}

func (a *Adapter) RemovePolicy(_ string, ptype string, rule []string) error {
	line := savePolicyLine(ptype, rule)
	err := a.repo.Delete(context.Background(), line)
	if err != nil {
		return err
	}

	return nil
}

func (a *Adapter) RemoveFilteredPolicy(_ string, _ string, _ int, _ ...string) error {
	return errors.New("not implemented")
}

func savePolicyLine(ptype string, rule []string) (line Rule) {
	line.PType = ptype

	if len(rule) > 0 {
		line.V0 = rule[0]
	}
	if len(rule) > 1 {
		line.V1 = rule[1]
	}
	if len(rule) > 2 {
		line.V2 = rule[2]
	}
	if len(rule) > 3 {
		line.V3 = rule[3]
	}
	if len(rule) > 4 {
		line.V4 = rule[4]
	}
	if len(rule) > 5 {
		line.V5 = rule[5]
	}

	return
}

func loadPolicyLine(line Rule, model model.Model) {
	p := []string{line.PType, line.V0, line.V1, line.V2,
		line.V3, line.V4, line.V5}

	for i := len(p) - 1; i > 0; i-- {
		if p[i] != "" {
			p = p[:i+1]
			break
		}
	}

	persist.LoadPolicyArray(p, model)
}
