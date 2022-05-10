package server

import (
	"context"
	"errors"
	"strings"

	"github.com/Leantar/fimserver/models"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type checkableEndpoint struct {
	Kind              string
	Roles             []interface{}
	HasBaseline       bool
	BaselineIsCurrent bool
}

func (s *Server) checkAuthorization(endpoint models.Endpoint, fullMethod string) error {
	method := strings.TrimPrefix(fullMethod, "/fim.Fim/")

	ep := checkableEndpoint{
		Kind:              endpoint.Kind,
		Roles:             make([]interface{}, 0),
		HasBaseline:       endpoint.HasBaseline,
		BaselineIsCurrent: endpoint.BaselineIsCurrent,
	}

	for _, role := range endpoint.Roles {
		ep.Roles = append(ep.Roles, role)
	}

	ok, err := s.enforcer.Enforce(ep, method)
	if err != nil {
		log.Error().Caller().Err(err).Msg("failed to check authz")
		return errors.New("unauthorized")
	}
	if !ok {
		return errors.New("unauthorized")
	}

	log.Info().Msgf("'%s' accessed '%s'", endpoint.Name, method)

	return nil
}

func (s *Server) UnaryAuthorizationInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	endpoint := ctx.Value(endpointKey("endpoint")).(models.Endpoint)

	err = s.checkAuthorization(endpoint, info.FullMethod)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, "unauthorized")
	}

	return handler(ctx, req)
}

func (s *Server) StreamAuthorizationInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	endpoint := stream.Context().Value(endpointKey("endpoint")).(models.Endpoint)

	err := s.checkAuthorization(endpoint, info.FullMethod)
	if err != nil {
		return status.Error(codes.PermissionDenied, "unauthorized")
	}

	return handler(srv, stream)
}
