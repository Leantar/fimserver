package server

import (
	"context"
	"io"

	"github.com/Leantar/fimproto/proto"
	"github.com/Leantar/fimserver/models"
	"github.com/Leantar/fimserver/modules/alert"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	KindChange = "CHANGE"
	KindCreate = "CREATE"
)

func (s *Server) GetStartupInfo(ctx context.Context, _ *proto.Empty) (*proto.StartupInfo, error) {
	agent := ctx.Value(endpointKey("endpoint")).(models.Endpoint)

	return &proto.StartupInfo{
		CreateBaseline: !agent.HasBaseline,
		UpdateBaseline: !agent.BaselineIsCurrent,
		WatchedPaths:   agent.WatchedPaths,
	}, nil
}

func (s *Server) CreateBaseline(stream proto.Fim_CreateBaselineServer) error {
	agent := stream.Context().Value(endpointKey("endpoint")).(models.Endpoint)

	var objs []models.FsObject

	for {
		fsObject, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Error().Caller().Err(err).Msg("failed to read from stream")
			return err
		}

		objs = append(objs, models.FsObject{
			Path:     fsObject.Path,
			Hash:     fsObject.Hash,
			Created:  fsObject.Created,
			Modified: fsObject.Modified,
			Uid:      fsObject.Uid,
			Gid:      fsObject.Gid,
			Mode:     fsObject.Mode,
			AgentID:  agent.ID,
		})
	}

	if len(objs) == 0 {
		log.Warn().Caller().Msg("agent presented empty baseline")
		return status.Error(codes.InvalidArgument, "baseline can't be empty")
	}

	err := s.repo.BaselineFsObjects().CreateMany(stream.Context(), objs)
	if err != nil {
		log.Error().Caller().Err(err).Msg("failed to create fs object")
		return status.Error(codes.Internal, "internal error")
	}

	agent.HasBaseline = true
	agent.BaselineIsCurrent = true
	err = s.repo.Endpoints().Update(stream.Context(), agent)
	if err != nil {
		log.Error().Caller().Err(err).Msg("failed to update agent")
		return status.Error(codes.Internal, "internal error")
	}

	log.Info().Msgf("'%s' set its baseline", agent.Name)

	return stream.SendAndClose(&proto.Empty{})
}

func (s *Server) UpdateBaseline(stream proto.Fim_UpdateBaselineServer) error {
	agent := stream.Context().Value(endpointKey("endpoint")).(models.Endpoint)

	err := s.repo.BaselineFsObjects().DeleteBaselineForAgent(stream.Context(), agent.ID)
	if err != nil {
		log.Error().Caller().Err(err).Msg("failed to delete baseline")
		return status.Error(codes.Internal, "internal error")
	}

	err = s.repo.Alerts().DeleteAll(stream.Context(), agent.ID)
	if err != nil {
		log.Error().Caller().Err(err).Msg("failed to delete baseline")
		return status.Error(codes.Internal, "internal error")
	}

	var objs []models.FsObject

	for {
		fsObject, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Error().Caller().Err(err).Msg("failed to read from stream")
			return err
		}

		objs = append(objs, models.FsObject{
			Path:     fsObject.Path,
			Hash:     fsObject.Hash,
			Created:  fsObject.Created,
			Modified: fsObject.Modified,
			Uid:      fsObject.Uid,
			Gid:      fsObject.Gid,
			Mode:     fsObject.Mode,
			AgentID:  agent.ID,
		})
	}

	if len(objs) == 0 {
		log.Warn().Caller().Msg("agent presented empty baseline")
		return status.Error(codes.InvalidArgument, "baseline can't be empty")
	}

	err = s.repo.BaselineFsObjects().CreateMany(stream.Context(), objs)
	if err != nil {
		log.Error().Caller().Err(err).Msg("failed to create fs object")
		return status.Error(codes.Internal, "internal error")
	}

	agent.BaselineIsCurrent = true
	err = s.repo.Endpoints().Update(stream.Context(), agent)
	if err != nil {
		log.Error().Caller().Err(err).Msg("failed to update agent")
		return status.Error(codes.Internal, "internal error")
	}

	log.Info().Msgf("'%s' updated its baseline", agent.Name)

	return stream.SendAndClose(&proto.Empty{})
}

func (s *Server) ReportFsStatus(stream proto.Fim_ReportFsStatusServer) error {
	agent := stream.Context().Value(endpointKey("endpoint")).(models.Endpoint)

	baseline, err := s.repo.BaselineFsObjects().GetBaselineByAgent(stream.Context(), agent.ID)
	if err != nil {
		log.Error().Caller().Err(err).Msg("failed to get baseline")
		return status.Error(codes.Internal, "internal error")
	}

	m := alert.NewManager(baseline, agent.ID)

	for {
		fsObject, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Error().Caller().Err(err).Msg("failed to read from stream")
			return err
		}

		obj := models.FsObject{
			Path:     fsObject.Path,
			Hash:     fsObject.Hash,
			Created:  fsObject.Created,
			Modified: fsObject.Modified,
			Uid:      fsObject.Uid,
			Gid:      fsObject.Gid,
			Mode:     fsObject.Mode,
		}
		m.CheckForAlert(obj)
	}

	alerts := m.Result()

	for _, al := range alerts {
		err = createAlertIfNotDuplicate(stream.Context(), s.repo, al, agent.ID)
		if err != nil {
			log.Error().Caller().Err(err).Msg("failed to create alert")
			return status.Error(codes.Internal, "internal error")
		}
	}

	return stream.SendAndClose(&proto.Empty{})
}

func (s *Server) ReportFsEvent(ctx context.Context, event *proto.Event) (*proto.Empty, error) {
	agent := ctx.Value(endpointKey("endpoint")).(models.Endpoint)

	var baseObj models.FsObject
	var err error
	if event.Kind == KindChange {
		baseObj, err = s.repo.BaselineFsObjects().GetByPathAndAgentID(ctx, event.FsObject.Path, agent.ID)
		if err != nil {
			if s.repo.IsEmptyResultSetError(err) {
				// Fanotify sometimes returns a CHANGE instead of a CREATE for newly created files.
				// We need to handle this case by overwriting the event kind
				event.Kind = KindCreate
			} else {
				log.Error().Caller().Err(err).Msg("failed to get fs object")
				return nil, status.Error(codes.Internal, "internal error")
			}
		}
	}

	evtObject := models.FsObject{
		Path:     event.FsObject.Path,
		Hash:     event.FsObject.Hash,
		Created:  event.FsObject.Created,
		Modified: event.FsObject.Modified,
		Uid:      event.FsObject.Uid,
		Gid:      event.FsObject.Gid,
		Mode:     event.FsObject.Mode,
	}

	al := models.Alert{
		Kind:     event.Kind,
		IssuedAt: event.IssuedAt,
		Path:     event.FsObject.Path,
		Modified: event.FsObject.Modified,
		AgentID:  agent.ID,
	}

	if al.Kind == KindChange {
		al.Difference = alert.GetDifference(baseObj, evtObject)
	}

	err = createAlertIfNotDuplicate(ctx, s.repo, al, agent.ID)
	if err != nil {
		log.Error().Caller().Err(err).Msg("failed to create alert")
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &proto.Empty{}, nil
}

func createAlertIfNotDuplicate(ctx context.Context, repo Repository, al models.Alert, agentID uint64) error {
	latest, err := repo.Alerts().GetLatestByPathAndAgent(ctx, al.Path, agentID)
	if repo.IsEmptyResultSetError(err) {
		// no previous alert exists for this path
		return repo.Alerts().Create(ctx, al)
	}
	if err != nil {
		return err
	}

	if al.Kind != latest.Kind {
		return repo.Alerts().Create(ctx, al)
	} else {
		if al.Kind == KindCreate && al.Modified > latest.Modified {
			return repo.Alerts().Create(ctx, al)
		} else if al.Kind == KindChange && al.Difference != latest.Difference {
			return repo.Alerts().Create(ctx, al)
		}
	}

	return nil
}
