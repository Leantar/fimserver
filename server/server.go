package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"github.com/Leantar/fimproto/proto"
	"github.com/Leantar/fimserver/models"
	casbinadapter "github.com/Leantar/fimserver/modules/casbin"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcValidator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io/ioutil"
	"net"
	"path/filepath"
	"strconv"
)

type EndpointRepository interface {
	Create(ctx context.Context, ep models.Endpoint) error
	GetByName(ctx context.Context, name string) (models.Endpoint, error)
	GetAgents(ctx context.Context) ([]models.Endpoint, error)
	Update(ctx context.Context, ep models.Endpoint) error
	Delete(ctx context.Context, name string) error
}

type BaselineFsObjectRepository interface {
	CreateMany(ctx context.Context, wfs []models.FsObject) error
	GetByPathAndAgentID(ctx context.Context, path string, agentID uint64) (models.FsObject, error)
	GetBaselineByAgent(ctx context.Context, agentID uint64) ([]models.FsObject, error)
	DeleteBaselineForAgent(ctx context.Context, agentID uint64) error
}

type AlertRepository interface {
	Create(ctx context.Context, alert models.Alert) error
	GetAllByAgent(ctx context.Context, agentID uint64) ([]models.Alert, error)
	GetLatestByPathAndAgent(ctx context.Context, path string, agentID uint64) (models.Alert, error)
	DeleteAll(ctx context.Context, agentID uint64) error
}

type Repository interface {
	IsEmptyResultSetError(err error) bool
	Endpoints() EndpointRepository
	BaselineFsObjects() BaselineFsObjectRepository
	Alerts() AlertRepository
	Rules() casbinadapter.RuleRepository
}

type Config struct {
	Host        string `yaml:"host"`
	Port        int64  `yaml:"port"`
	CertFile    string `yaml:"cert_file"`
	CertKeyFile string `yaml:"cert_key_file"`
	CaFile      string `yaml:"ca_file"`
}

type Server struct {
	proto.UnimplementedFimServer
	srv      *grpc.Server
	repo     Repository
	enforcer *casbin.Enforcer
	conf     Config
}

func New(repo Repository, config Config) *Server {
	a := casbinadapter.NewAdapter(repo.Rules())
	m, err := model.NewModelFromString(`[request_definition]
	r = sub, obj
	
	[policy_definition]
	p = sub, obj
	
	[role_definition]
	g = _, _
	
	[policy_effect]
	e = some(where (p.eft == allow))
	
	[matchers]
	m = g(r.sub, p.sub) && r.obj == p.obj
	`)
	if err != nil {
		log.Fatal().Caller().Err(err).Msg("failed to create casbin model")
	}

	e, err := casbin.NewEnforcer(m, a)
	if err != nil {
		log.Fatal().Caller().Err(err).Msg("failed to create casbin enforcer")
	}

	return &Server{
		repo:     repo,
		enforcer: e,
		conf:     config,
	}
}

func (s *Server) Run() error {
	address := net.JoinHostPort(s.conf.Host, strconv.FormatInt(s.conf.Port, 10))

	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	creds, err := createGrpcCredentials(s.conf.CertFile, s.conf.CertKeyFile, s.conf.CaFile)
	if err != nil {
		return err
	}

	srv := grpc.NewServer(
		grpc.StreamInterceptor(
			middleware.ChainStreamServer(
				s.StreamAuthenticationInterceptor,
				s.StreamAuthorizationInterceptor,
				s.StreamOneTimeRoleRemoveInterceptor,
				grpcValidator.StreamServerInterceptor(),
			),
		),
		grpc.UnaryInterceptor(
			middleware.ChainUnaryServer(
				s.UnaryAuthenticationInterceptor,
				s.UnaryAuthorizationInterceptor,
				s.UnaryOneTimeRoleRemoveInterceptor,
				grpcValidator.UnaryServerInterceptor(),
			),
		),
		grpc.Creds(creds),
	)

	s.srv = srv

	proto.RegisterFimServer(srv, s)

	log.Info().Msgf("starting to listen on: %s", address)
	return srv.Serve(listener)
}

func (s *Server) Stop() {
	log.Info().Msg("shutting down")
	s.srv.GracefulStop()
}

func createGrpcCredentials(certPath, keyPath, caPath string) (credentials.TransportCredentials, error) {
	certPath, err := filepath.Abs(certPath)
	if err != nil {
		return nil, err
	}

	keyPath, err = filepath.Abs(keyPath)
	if err != nil {
		return nil, err
	}

	caPath, err = filepath.Abs(caPath)
	if err != nil {
		return nil, err
	}

	srvCert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, err
	}

	caBytes, err := ioutil.ReadFile(caPath)
	if err != nil {
		return nil, err
	}

	pool := x509.NewCertPool()
	ok := pool.AppendCertsFromPEM(caBytes)
	if !ok {
		return nil, errors.New("couldn't parse ca.cert")
	}

	return credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{srvCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    pool,
		MinVersion:   tls.VersionTLS13,
	}), nil
}
