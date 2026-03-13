package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"starlite/internal/db/zz"
	"starlite/internal/features/index/models"

	toolbeltdb "github.com/delaneyj/toolbelt/db"
	"github.com/delaneyj/toolbelt/embeddednats"
	"github.com/nats-io/nats.go"
	"github.com/starfederation/datastar-go/datastar"
	"zombiezen.com/go/sqlite"
)

type ReactionsService struct {
	log *slog.Logger
	db  *toolbeltdb.Database
	nc  *nats.Conn
}

func NewReactionsService(log *slog.Logger, db *toolbeltdb.Database, ns *embeddednats.Server) (*ReactionsService, error) {
	nc, err := ns.Client()
	if err != nil {
		return nil, fmt.Errorf("error creating nats client: %w", err)
	}

	return &ReactionsService{
		log: log,
		db:  db,
		nc:  nc,
	}, nil
}

// Start begins listening for relevant NATS events and updates the reaction counts accordingly.
func (s *ReactionsService) Start(ctx context.Context) error {
	_, err := s.nc.Subscribe("reaction.added", func(msg *nats.Msg) {
		var req models.ReactionRequest
		if err := json.Unmarshal(msg.Data, &req); err != nil {
			s.log.Error("error unmarshaling reaction event data", "error", err)
			return
		}

		err := s.db.WriteWithoutTx(ctx, func(tx *sqlite.Conn) error {
			return zz.OnceIncrementReactionCount(tx, req.Reaction)
		})
		if err != nil {
			s.log.Error("error incrementing reaction count", "error", err)
		}

		err = s.nc.Publish("reaction.updated", nil)
		if err != nil {
			s.log.Error("error publishing reaction.updated event", "error", err)
		}
	})
	if err != nil {
		return fmt.Errorf("error subscribing to reaction.added: %w", err)
	}
	return nil
}

func (s *ReactionsService) GetReactions(ctx context.Context) (map[string]int, error) {
	var counts []*zz.ReactionCountModel
	err := s.db.ReadTX(ctx, func(tx *sqlite.Conn) error {
		var err error
		counts, err = zz.OnceGetReactionCounts(tx)
		if err != nil {
			return fmt.Errorf("error getting reaction counts: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error reading reactions from database: %w", err)
	}

	countsMap := make(map[string]int)
	for _, c := range counts {
		countsMap[c.Reaction] = int(c.Count)
	}

	return countsMap, nil
}

func (s *ReactionsService) AddReaction(ctx context.Context, userID, reaction string) error {
	err := s.db.WriteWithoutTx(ctx, func(tx *sqlite.Conn) error {

		err := zz.OnceAddReaction(tx, zz.AddReactionParams{
			UserId:   userID,
			Reaction: reaction,
		})
		if err != nil {
			return fmt.Errorf("error adding reaction: %w", err)
		}

		msgData, err := json.Marshal(models.ReactionRequest{
			Reaction: reaction,
		})
		if err != nil {
			return fmt.Errorf("error marshaling reaction event data: %w", err)
		}

		if err := s.nc.PublishMsg(&nats.Msg{
			Subject: "reaction.added",
			Data:    msgData,
		}); err != nil {
			return fmt.Errorf("error publishing reaction event: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("error writing reaction to database: %w", err)
	}

	return nil
}

func (s *ReactionsService) HandleReactionSSE(ctx context.Context, sse *datastar.ServerSentEventGenerator) error {
	sub, err := s.nc.Subscribe("reaction.updated", func(msg *nats.Msg) {
		counts, err := s.GetReactions(ctx)
		if err != nil {
			s.log.Error("error getting reaction counts for SSE update", "error", err)
			return
		}

		if err := sse.MarshalAndPatchSignals(counts); err != nil {
			s.log.Error("error marshaling and patching SSE signals", "error", err)
		}
	})
	if err != nil {
		return fmt.Errorf("error subscribing to reaction.updated: %w", err)
	}

	<-ctx.Done()
	return sub.Unsubscribe()
}
