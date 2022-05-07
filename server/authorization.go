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

func (s *Server) removeOneTimeRoles(endpointName, fullMethod string) (err error) {
	method := strings.TrimPrefix(fullMethod, "/fim.Fim/")

	switch method {
	case "CreateBaseline":
		_, err = s.enforcer.DeleteRoleForUser(endpointName, "baseline")
	case "UpdateBaseline":
		_, err = s.enforcer.DeleteRoleForUser(endpointName, "updater")
	}

	return
}

func (s *Server) UnaryOneTimeRoleRemoveInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	endpoint := ctx.Value(endpointKey("endpoint")).(models.Endpoint)

	resp, err := handler(ctx, req)
	if err != nil {
		return nil, err
	}

	err = s.removeOneTimeRoles(endpoint.Name, info.FullMethod)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	return resp, nil
}

func (s *Server) StreamOneTimeRoleRemoveInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	endpoint := stream.Context().Value(endpointKey("endpoint")).(models.Endpoint)

	err := handler(srv, stream)
	if err != nil {
		return err
	}

	err = s.removeOneTimeRoles(endpoint.Name, info.FullMethod)
	if err != nil {
		return status.Error(codes.Internal, "internal error")
	}

	return nil
}

func (s *Server) checkAuthorization(endpointName, fullMethod string) error {
	method := strings.TrimPrefix(fullMethod, "/fim.Fim/")

	ok, err := s.enforcer.Enforce(endpointName, method)
	if err != nil {
		log.Error().Caller().Err(err).Msg("failed to check authz")
		return errors.New("unauthorized")
	}
	if !ok {
		return errors.New("unauthorized")
	}

	log.Info().Msgf("'%s' accessed '%s'", endpointName, method)

	return nil
}

func (s *Server) UnaryAuthorizationInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	endpoint := ctx.Value(endpointKey("endpoint")).(models.Endpoint)

	err = s.checkAuthorization(endpoint.Name, info.FullMethod)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, "unauthorized")
	}

	return handler(ctx, req)
}

func (s *Server) StreamAuthorizationInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	endpoint := stream.Context().Value(endpointKey("endpoint")).(models.Endpoint)

	err := s.checkAuthorization(endpoint.Name, info.FullMethod)
	if err != nil {
		return status.Error(codes.PermissionDenied, "unauthorized")
	}

	return handler(srv, stream)
}
