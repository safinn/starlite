package index

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"starlite/internal/features/index/models"
	"starlite/internal/features/index/pages"
	"starlite/internal/features/index/services"
	"starlite/pkg/logctx"

	"github.com/alexedwards/scs/v2"
	"github.com/delaneyj/toolbelt/id"
	"github.com/starfederation/datastar-go/datastar"
)

type Handlers struct {
	log              *slog.Logger
	reactionsService *services.ReactionsService
	sessionManager   *scs.SessionManager
}

func NewHandlers(log *slog.Logger, rs *services.ReactionsService, sm *scs.SessionManager) *Handlers {
	return &Handlers{
		log:              log,
		reactionsService: rs,
		sessionManager:   sm,
	}
}

func (h *Handlers) IndexPage(w http.ResponseWriter, r *http.Request) {
	counts, err := h.reactionsService.GetReactions(r.Context())
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if err := pages.Index(counts).Render(r.Context(), w); err != nil {
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

	err := h.reactionsService.AddReaction(r.Context(), userID, req.Reaction)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) ReactionsSSE(w http.ResponseWriter, r *http.Request) {
	sse := datastar.NewSSE(w, r)

	if err := h.reactionsService.HandleReactionSSE(r.Context(), sse); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (h *Handlers) getUserIDFromSession(ctx context.Context) string {
	userID := h.sessionManager.GetString(ctx, "userID")
	if userID == "" {
		h.log.Debug("Generating new userID")
		userID = id.NextEncodedID()
		h.sessionManager.Put(ctx, "userID", userID)
	}

	logctx.Set(ctx, slog.String("userID", userID))

	return userID
}
