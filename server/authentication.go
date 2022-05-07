package server

import (
	"context"
	"errors"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type endpointKey string

func (s *Server) checkAuthentication(ctx context.Context) (context.Context, error) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return ctx, errors.New("couldn't get peer from ctx")
	}

	tlsAuth, ok := p.AuthInfo.(credentials.TLSInfo)
	if !ok {
		return ctx, errors.New("invalid credential type")
	}

	name := tlsAuth.State.PeerCertificates[0].Subject.CommonName

	endpoint, err := s.repo.Endpoints().GetByName(ctx, name)
	if err != nil {
		log.Warn().Caller().Err(err).Msg("failed to get endpoint by name")
		return ctx, err
	}

	return context.WithValue(ctx, endpointKey("endpoint"), endpoint), nil
}

func (s *Server) StreamAuthenticationInterceptor(srv interface{}, stream grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	newCtx, err := s.checkAuthentication(stream.Context())
	if err != nil {
		return status.Error(codes.Unauthenticated, "unauthenticated")
	}

	wrapped := grpc_middleware.WrapServerStream(stream)
	wrapped.WrappedContext = newCtx

	return handler(srv, wrapped)
}

func (s *Server) UnaryAuthenticationInterceptor(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	newCtx, err := s.checkAuthentication(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	return handler(newCtx, req)
}
