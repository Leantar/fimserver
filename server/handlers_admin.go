package server

import (
	"context"

	"github.com/Leantar/fimproto/proto"
	"github.com/Leantar/fimserver/models"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) CreateAgentEndpoint(ctx context.Context, endpoint *proto.AgentEndpoint) (*proto.Empty, error) {
	admin := ctx.Value(endpointKey("endpoint")).(models.Endpoint)

	_, err := s.repo.Endpoints().GetByName(ctx, endpoint.Name)
	if err == nil {
		return nil, status.Error(codes.AlreadyExists, "endpoint already exists")
	}

	e := models.Endpoint{
		Name:              endpoint.Name,
		Kind:              "agent",
		Roles:             []string{},
		HasBaseline:       false,
		BaselineIsCurrent: false,
		WatchedPaths:      endpoint.WatchedPaths,
	}

	err = s.repo.Endpoints().Create(ctx, e)
	if err != nil {
		log.Error().Caller().Err(err).Msg("failed to create endpoint")
		return nil, status.Error(codes.Internal, "internal error")
	}

	log.Info().Msgf("'%s' created agent '%s", admin.Name, endpoint.Name)

	return &proto.Empty{}, nil
}

func (s *Server) CreateClientEndpoint(ctx context.Context, endpoint *proto.ClientEndpoint) (*proto.Empty, error) {
	admin := ctx.Value(endpointKey("endpoint")).(models.Endpoint)

	_, err := s.repo.Endpoints().GetByName(ctx, endpoint.Name)
	if err == nil {
		return nil, status.Error(codes.AlreadyExists, "endpoint already exists")
	}

	e := models.Endpoint{
		Name:         endpoint.Name,
		Kind:         "client",
		Roles:        endpoint.Roles,
		WatchedPaths: []string{},
	}

	err = s.repo.Endpoints().Create(ctx, e)
	if err != nil {
		log.Error().Caller().Err(err).Msg("failed to create endpoint")
		return nil, status.Error(codes.Internal, "internal error")
	}

	log.Info().Msgf("'%s' created client '%s' with roles '%v'", admin.Name, endpoint.Name, endpoint.Roles)

	return &proto.Empty{}, nil
}

func (s *Server) DeleteEndpoint(ctx context.Context, endpointName *proto.EndpointName) (*proto.Empty, error) {
	admin := ctx.Value(endpointKey("endpoint")).(models.Endpoint)

	_, err := s.repo.Endpoints().GetByName(ctx, endpointName.Name)
	if err != nil {
		if s.repo.IsEmptyResultSetError(err) {
			return nil, status.Error(codes.NotFound, "endpoint not found")
		}
		log.Error().Caller().Err(err).Msg("failed to get endpoint")
		return nil, status.Error(codes.Internal, "internal error")
	}

	err = s.repo.Endpoints().Delete(ctx, endpointName.Name)
	if err != nil {
		log.Error().Caller().Err(err).Msg("failed to delete endpoint")
		return nil, status.Error(codes.Internal, "internal error")
	}

	log.Info().Msgf("'%s' deleted endpoint '%s'", admin.Name, endpointName.Name)

	return &proto.Empty{}, nil
}

func (s *Server) UpdateEndpointWatchedPaths(ctx context.Context, obj *proto.AgentEndpoint) (*proto.Empty, error) {
	admin := ctx.Value(endpointKey("endpoint")).(models.Endpoint)

	agent, err := s.repo.Endpoints().GetByName(ctx, obj.Name)
	if err != nil {
		if s.repo.IsEmptyResultSetError(err) {
			return nil, status.Error(codes.NotFound, "agent not found")
		}
		log.Error().Caller().Err(err).Msg("failed to get agent")
		return nil, status.Error(codes.Internal, "internal error")
	}

	agent.WatchedPaths = obj.WatchedPaths

	err = s.repo.Endpoints().Update(ctx, agent)
	if err != nil {
		log.Error().Caller().Err(err).Msg("failed to update endpoint")
		return nil, status.Error(codes.Internal, "internal error")
	}

	log.Info().Msgf("'%s' changed watched paths for '%s'", admin.Name, agent.Name)

	return &proto.Empty{}, nil
}
