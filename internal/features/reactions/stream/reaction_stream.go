package stream

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"starlite/internal/db/zz"
	"starlite/internal/features/reactions/models"

	toolbeltdb "github.com/delaneyj/toolbelt/db"
	"github.com/nats-io/nats.go"
	"zombiezen.com/go/sqlite"
)

type ReactionStream struct {
	log *slog.Logger
	db  *toolbeltdb.Database
	nc  *nats.Conn
}

func New(ctx context.Context, log *slog.Logger, db *toolbeltdb.Database, nc *nats.Conn) (*ReactionStream, error) {
	s := &ReactionStream{log: log, db: db, nc: nc}
	if err := s.start(ctx); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *ReactionStream) start(ctx context.Context) error {
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

func (s *ReactionStream) Add(ctx context.Context, userID, reaction string) error {
	err := s.db.WriteWithoutTx(ctx, func(tx *sqlite.Conn) error {
		err := zz.OnceAddReaction(tx, zz.AddReactionParams{
			UserId:   userID,
			Reaction: reaction,
		})
		if err != nil {
			return fmt.Errorf("error adding reaction: %w", err)
		}

		msgData, err := json.Marshal(models.ReactionRequest{Reaction: reaction})
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

func (s *ReactionStream) Snapshot(ctx context.Context) (map[string]int, error) {
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

	countsMap := make(map[string]int, len(counts))
	for _, c := range counts {
		countsMap[c.Reaction] = int(c.Count)
	}
	return countsMap, nil
}

func (s *ReactionStream) Subscribe(ctx context.Context, handler func(map[string]int)) error {
	sub, err := s.nc.Subscribe("reaction.updated", func(msg *nats.Msg) {
		counts, err := s.Snapshot(ctx)
		if err != nil {
			s.log.Error("error getting reaction counts for SSE update", "error", err)
			return
		}
		handler(counts)
	})
	if err != nil {
		return fmt.Errorf("error subscribing to reaction.updated: %w", err)
	}

	<-ctx.Done()
	return sub.Unsubscribe()
}
