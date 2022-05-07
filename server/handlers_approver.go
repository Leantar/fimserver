package server

import (
	"context"

	"github.com/Leantar/fimproto/proto"
	"github.com/Leantar/fimserver/models"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) CreateBaselineUpdateApproval(ctx context.Context, endpointName *proto.EndpointName) (*proto.Empty, error) {
	approver := ctx.Value(endpointKey("endpoint")).(models.Endpoint)

	agent, err := s.repo.Endpoints().GetByName(ctx, endpointName.Name)
	if err != nil {
		if s.repo.IsEmptyResultSetError(err) {
			return nil, status.Error(codes.NotFound, "agent was not found")
		}
		log.Error().Caller().Err(err).Msg("failed to get agent")
		return nil, status.Error(codes.Internal, "internal error")
	}

	ok, err := s.enforcer.AddRoleForUser(agent.Name, "updater")
	if err != nil {
		log.Error().Caller().Err(err).Msg("failed to assign updater role to agent")
		return nil, status.Error(codes.Internal, "internal error")
	}
	if !ok {
		return nil, status.Error(codes.AlreadyExists, "agent already has permission to update baseline")
	}

	agent.BaselineIsCurrent = false
	err = s.repo.Endpoints().Update(ctx, agent)
	if err != nil {
		log.Error().Caller().Err(err).Msg("failed to update agent")
		return nil, status.Error(codes.Internal, "internal error")
	}

	log.Info().Caller().Msgf("'%s' allowed agent '%s' to update it's baseline", approver.Name, agent.Name)

	return &proto.Empty{}, nil
}
