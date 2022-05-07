package server

import (
	"github.com/Leantar/fimproto/proto"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) GetAgents(_ *proto.Empty, stream proto.Fim_GetAgentsServer) error {
	agents, err := s.repo.Endpoints().GetAgents(stream.Context())
	if err != nil {
		if s.repo.IsEmptyResultSetError(err) {
			return status.Error(codes.NotFound, "no agents were found")
		}
		log.Error().Caller().Err(err).Msg("failed to get agents")
		return status.Error(codes.Internal, "internal error")
	}

	for _, a := range agents {
		agent := proto.Agent{
			Name:              a.Name,
			HasBaseline:       a.HasBaseline,
			BaselineIsCurrent: a.BaselineIsCurrent,
			WatchedPaths:      a.WatchedPaths,
		}

		err := stream.Send(&agent)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) GetAlertsByAgent(endpointName *proto.EndpointName, stream proto.Fim_GetAlertsByAgentServer) error {
	agent, err := s.repo.Endpoints().GetByName(stream.Context(), endpointName.Name)
	if err != nil {
		if s.repo.IsEmptyResultSetError(err) {
			return status.Error(codes.NotFound, "no agent was found")
		}
		log.Error().Caller().Err(err).Msg("failed to get agent")
		return status.Error(codes.Internal, "internal error")
	}

	alerts, err := s.repo.Alerts().GetAllByAgent(stream.Context(), agent.ID)
	if err != nil {
		if s.repo.IsEmptyResultSetError(err) {
			return status.Error(codes.NotFound, "no alerts were found")
		}
		log.Error().Caller().Err(err).Msg("failed to get alerts")
		return status.Error(codes.Internal, "internal error")
	}

	for _, a := range alerts {
		alert := proto.Alert{
			Kind:       a.Kind,
			Difference: a.Difference,
			Path:       a.Path,
			IssuedAt:   a.IssuedAt,
		}

		err := stream.Send(&alert)
		if err != nil {
			return err
		}
	}

	return nil
}
