package preparation

import (
	"context"
	"fmt"

	"github.com/Leantar/fimserver/models"
	casbinadapter "github.com/Leantar/fimserver/modules/casbin"
	"github.com/Leantar/fimserver/repository"
	"github.com/rs/zerolog/log"
)

func Setup(repo *repository.PgRepository) error {
	err := repo.ApplySchema()
	if err != nil {
		return fmt.Errorf("schema: %w", err)
	}

	err = createCasbinPolicy(repo)
	if err != nil {
		return fmt.Errorf("schema: %w", err)
	}

	err = createAdminUser(repo)
	if err != nil {
		return fmt.Errorf("schema: %w", err)
	}

	log.Info().Msg("finished setup. please restart without setup mode")

	return nil
}

func createCasbinPolicy(repo *repository.PgRepository) error {
	ctx := context.Background()

	err := repo.Rules().Create(ctx, casbinadapter.Rule{
		PType: "p",
		V0:    "reporter",
		V1:    "GetStartupInfo",
	})
	if err != nil {
		return err
	}

	err = repo.Rules().Create(context.Background(), casbinadapter.Rule{
		PType: "p",
		V0:    "baseline",
		V1:    "CreateBaseline",
	})
	if err != nil {
		return err
	}

	err = repo.Rules().Create(context.Background(), casbinadapter.Rule{
		PType: "p",
		V0:    "updater",
		V1:    "UpdateBaseline",
	})
	if err != nil {
		return err
	}

	err = repo.Rules().Create(context.Background(), casbinadapter.Rule{
		PType: "p",
		V0:    "reporter",
		V1:    "ReportFsStatus",
	})
	if err != nil {
		return err
	}

	err = repo.Rules().Create(context.Background(), casbinadapter.Rule{
		PType: "p",
		V0:    "reporter",
		V1:    "ReportFsEvent",
	})
	if err != nil {
		return err
	}

	err = repo.Rules().Create(context.Background(), casbinadapter.Rule{
		PType: "p",
		V0:    "viewer",
		V1:    "GetAgents",
	})
	if err != nil {
		return err
	}

	err = repo.Rules().Create(context.Background(), casbinadapter.Rule{
		PType: "p",
		V0:    "viewer",
		V1:    "GetAlertsByAgent",
	})
	if err != nil {
		return err
	}

	err = repo.Rules().Create(context.Background(), casbinadapter.Rule{
		PType: "p",
		V0:    "approver",
		V1:    "CreateBaselineUpdateApproval",
	})
	if err != nil {
		return err
	}

	err = repo.Rules().Create(context.Background(), casbinadapter.Rule{
		PType: "p",
		V0:    "user_admin",
		V1:    "CreateAgentEndpoint",
	})
	if err != nil {
		return err
	}

	err = repo.Rules().Create(context.Background(), casbinadapter.Rule{
		PType: "p",
		V0:    "user_admin",
		V1:    "CreateClientEndpoint",
	})
	if err != nil {
		return err
	}

	err = repo.Rules().Create(context.Background(), casbinadapter.Rule{
		PType: "p",
		V0:    "user_admin",
		V1:    "DeleteEndpoint",
	})
	if err != nil {
		return err
	}

	err = repo.Rules().Create(context.Background(), casbinadapter.Rule{
		PType: "p",
		V0:    "user_admin",
		V1:    "UpdateEndpointWatchedPaths",
	})
	if err != nil {
		return err
	}

	return nil
}

func createAdminUser(repo *repository.PgRepository) error {
	ctx := context.Background()

	err := repo.Endpoints().Create(ctx, models.Endpoint{
		Name:              "admin",
		Kind:              "client",
		HasBaseline:       false,
		BaselineIsCurrent: false,
		WatchedPaths:      []string{""},
	})
	if err != nil {
		return err
	}

	err = repo.Rules().Create(context.Background(), casbinadapter.Rule{
		PType: "g",
		V0:    "admin",
		V1:    "user_admin",
	})
	if err != nil {
		return err
	}

	err = repo.Rules().Create(context.Background(), casbinadapter.Rule{
		PType: "g",
		V0:    "admin",
		V1:    "viewer",
	})
	if err != nil {
		return err
	}

	err = repo.Rules().Create(context.Background(), casbinadapter.Rule{
		PType: "g",
		V0:    "admin",
		V1:    "approver",
	})
	if err != nil {
		return err
	}

	return nil
}
