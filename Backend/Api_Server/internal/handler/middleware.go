package internal

import (
	"context"
	"net/http"
)

type contextKey string

const (
	UserIDContextKey contextKey = "user_id"
)

func (h *PartHandler) MiddlAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("auth_session")
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		authSession, err := h.service.GetAuthSession(r.Context(), cookie.Value)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDContextKey, authSession.ID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
