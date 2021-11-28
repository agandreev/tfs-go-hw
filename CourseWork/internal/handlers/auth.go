package handlers

import (
	"context"
	"encoding/json"
	"github.com/agandreev/tfs-go-hw/CourseWork/internal/domain"
	"io"
	"net/http"
	"strings"
)

// ctxKey is necessary for context value transfer.
type ctxKey string

// loginHandler handles login algorithm.
func (handler *Handler) loginHandler(w http.ResponseWriter, r *http.Request) {
	var input domain.User

	d, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	if err = json.Unmarshal(d, &input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	token, err := handler.Trader.Users.GenerateJWT(input)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	data, err := json.Marshal(map[string]interface{}{"token": token})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if _, err = w.Write(data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// authHandler handles auth algorithm.
func (handler *Handler) authHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		headerParts := strings.Split(header, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if len(headerParts[1]) == 0 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		username, err := handler.Trader.Users.ParseToken(headerParts[1])
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, ctxKey("username"), username)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}
