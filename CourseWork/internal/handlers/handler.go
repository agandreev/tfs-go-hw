package handlers

import (
	"encoding/json"
	"github.com/agandreev/tfs-go-hw/CourseWork/internal/domain"
	"github.com/agandreev/tfs-go-hw/CourseWork/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"io"
	"net/http"
	"time"
)

// Handler processes all http handlers and consists of service realization.
type Handler struct {
	Trader *service.AlgoTrader
}

// InitRoutes initializes routes with necessary middlewares.
func (handler *Handler) InitRoutes() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)
	r.Use(middleware.Timeout(10 * time.Second))

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", handler.addUserHandler)
		r.Post("/login", handler.loginHandler)
	})

	r.Group(func(r chi.Router) {
		r.Use(handler.authHandler)
		r.Route("/users", func(r chi.Router) {
			r.Post("/set_keys", handler.setKeys)
		})
		r.Route("/pair", func(r chi.Router) {
			r.Post("/start", handler.startPairHandler)
			r.Post("/stop", handler.stopPairHandler)
		})
	})

	return r
}

// addUserHandler handles adding user algorithm.
func (handler *Handler) addUserHandler(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		processError(w, http.StatusBadRequest, err)
		return
	}
	defer r.Body.Close()
	user := &domain.User{}
	if err = json.Unmarshal(data, user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err = handler.Trader.AddUser(*user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// setKeys handles setting keys algorithm.
func (handler *Handler) setKeys(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value(ctxKey("username")).(string)
	data, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	user := &domain.User{}
	if err = json.Unmarshal(data, user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err = handler.Trader.Users.SetKeys(username, user.PublicKey, user.PrivateKey); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// addUserHandler handles starting pair algorithm.
func (handler *Handler) startPairHandler(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value(ctxKey("username")).(string)
	data, err := io.ReadAll(r.Body)
	if err != nil {
		processError(w, http.StatusBadRequest, err)
		return
	}
	defer r.Body.Close()
	config := domain.Config{}
	if err = json.Unmarshal(data, &config); err != nil {
		processError(w, http.StatusBadRequest, err)
		return
	}
	if err = handler.Trader.AddPair(username, config); err != nil {
		processError(w, http.StatusBadRequest, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// addUserHandler handles stopping pair algorithm.
func (handler *Handler) stopPairHandler(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value(ctxKey("username")).(string)
	data, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	config := domain.Config{}
	if err = json.Unmarshal(data, &config); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err = handler.Trader.DeletePair(username, config); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// processError sends status code with error text.
func processError(w http.ResponseWriter, status int, err error) {
	w.WriteHeader(status)
	respBody, err := json.Marshal(domain.ErrorJSON{Message: err.Error()})
	if err != nil {
		return
	}
	_, _ = w.Write(respBody)
}
