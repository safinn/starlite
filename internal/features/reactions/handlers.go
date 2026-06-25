package reactions

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"starlite/internal/features/reactions/models"
	"starlite/internal/features/reactions/pages"
	"starlite/internal/features/reactions/stream"
	"starlite/pkg/logger"

	"github.com/alexedwards/scs/v2"
	"github.com/delaneyj/toolbelt/id"
	"github.com/starfederation/datastar-go/datastar"
)

type Handlers struct {
	log            *slog.Logger
	reactionStream *stream.ReactionStream
	sessionManager *scs.SessionManager
	baseURL        string
	isDev          bool
}

func NewHandlers(log *slog.Logger, isDev bool, baseURL string, rs *stream.ReactionStream, sm *scs.SessionManager) *Handlers {
	return &Handlers{
		log:            log,
		reactionStream: rs,
		sessionManager: sm,
		baseURL:        baseURL,
		isDev:          isDev,
	}
}

func (h *Handlers) IndexPage(w http.ResponseWriter, r *http.Request) {
	counts, err := h.reactionStream.Snapshot(r.Context())
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if err := pages.Index(h.isDev, h.baseURL, counts).Render(r.Context(), w); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (h *Handlers) HandleReaction(w http.ResponseWriter, r *http.Request) {
	var req models.ReactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if !req.IsValid() {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	userID := h.getUserIDFromSession(r.Context())

	err := h.reactionStream.Add(r.Context(), userID, req.Reaction)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) ReactionsSSE(w http.ResponseWriter, r *http.Request) {
	sse := datastar.NewSSE(w, r)

	if err := h.reactionStream.Subscribe(r.Context(), func(counts map[string]int) {
		if err := sse.MarshalAndPatchSignals(counts); err != nil {
			h.log.Error("error marshaling and patching SSE signals", "error", err)
		}
	}); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (h *Handlers) getUserIDFromSession(ctx context.Context) string {
	userID := h.sessionManager.GetString(ctx, "userID")
	if userID == "" {
		h.log.Debug("Generating new userID")
		userID = id.NextEncodedID()
		h.sessionManager.Put(ctx, "userID", userID)
	}

	logger.Set(ctx, slog.String("userID", userID))

	return userID
}
