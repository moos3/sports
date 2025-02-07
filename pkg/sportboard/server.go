package sportboard

import (
	"context"
	"net/http"

	"github.com/twitchtv/twirp"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/robbydyer/sports/internal/proto/sportboard"
)

// Server ...
type Server struct {
	board *SportBoard
}

// GetRPCHandler ...
func (s *SportBoard) GetRPCHandler() (string, http.Handler) {
	return s.rpcServer.PathPrefix(), s.rpcServer
}

// SetStatus ...
func (s *Server) SetStatus(ctx context.Context, req *pb.SetStatusReq) (*emptypb.Empty, error) {
	cancelBoard := false
	clearDrawCache := false

	if req.Status == nil {
		return &emptypb.Empty{}, twirp.NewError(twirp.InvalidArgument, "nil status sent")
	}

	if s.board.config.HideFavoriteScore.CAS(!req.Status.FavoriteHidden, req.Status.FavoriteHidden) {
		cancelBoard = true
	}
	if req.Status.Enabled {
		if s.board.Enable() {
			cancelBoard = true
		}
	} else {
		if s.board.Disable() {
			cancelBoard = true
		}
	}
	if s.board.config.FavoriteSticky.CAS(!req.Status.FavoriteSticky, req.Status.FavoriteSticky) {
		cancelBoard = true
	}
	if s.board.config.GamblingSpread.CAS(!req.Status.OddsEnabled, req.Status.OddsEnabled) {
		cancelBoard = true
		clearDrawCache = true
	}
	if s.board.config.ScrollMode.CAS(!req.Status.ScrollEnabled, req.Status.ScrollEnabled) {
		cancelBoard = true
		clearDrawCache = true
	}
	if s.board.config.TightScroll.CAS(!req.Status.TightScrollEnabled, req.Status.TightScrollEnabled) {
		cancelBoard = true
		clearDrawCache = true
	}
	if s.board.config.ShowRecord.CAS(!req.Status.RecordRankEnabled, req.Status.RecordRankEnabled) {
		cancelBoard = true
		clearDrawCache = true
	}
	if s.board.config.UseGradient.CAS(!req.Status.UseGradient, req.Status.UseGradient) {
		cancelBoard = true
		clearDrawCache = true
	}
	if s.board.config.LiveOnly.CAS(!req.Status.LiveOnly, req.Status.LiveOnly) {
		cancelBoard = true
	}

	if clearDrawCache {
		s.board.clearDrawCache()
	}

	if cancelBoard {
		s.board.callCancelBoard()
	}

	return &emptypb.Empty{}, nil
}

// GetStatus ...
func (s *Server) GetStatus(ctx context.Context, req *emptypb.Empty) (*pb.StatusResp, error) {
	return &pb.StatusResp{
		Status: &pb.Status{
			Enabled:            s.board.config.Enabled.Load(),
			FavoriteHidden:     s.board.config.HideFavoriteScore.Load(),
			FavoriteSticky:     s.board.config.FavoriteSticky.Load(),
			ScrollEnabled:      s.board.config.ScrollMode.Load(),
			TightScrollEnabled: s.board.config.TightScroll.Load(),
			RecordRankEnabled:  s.board.config.ShowRecord.Load(),
			OddsEnabled:        s.board.config.GamblingSpread.Load(),
			UseGradient:        s.board.config.UseGradient.Load(),
			LiveOnly:           s.board.config.LiveOnly.Load(),
		},
	}, nil
}
