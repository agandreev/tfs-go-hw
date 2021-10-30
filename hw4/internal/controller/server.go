package controller

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/agandreev/tfs-go-hw/hw4/internal/domain"
	"github.com/agandreev/tfs-go-hw/hw4/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt"
)

type ctxKey string

type Server struct {
	service.GeneralChat
}

type UserCredentials struct {
	Password string `json:"password"`
	Username string `json:"username"`
}

type MessageCredentials struct {
	Message string `json:"message"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func (server *Server) Run() {
	r := server.registerRoutes()
	srv := http.Server{Addr: "0.0.0.0:5000", Handler: r}
	sig := make(chan os.Signal, 1)
	stop := make(chan struct{})
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig
		err := srv.Shutdown(context.Background())
		if err != nil {
			log.Fatal(err)
		}
		stop<- struct{}{}
	}()

	err := srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	<-stop
}

func (server *Server) registerRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)
	r.Use(middleware.Timeout(10 * time.Second))

	r.Group(func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/sign_up", server.registerHandler)
			r.Post("/sign_in", server.loginHandler)
		})
	})

	r.Group(func(r chi.Router) {
		r.Use(server.authHandler)
		r.Route("/users", func(r chi.Router) {
			r.Get("/me/messages", server.getDirectMessagesHandler)
			r.Post("/{id}/messages", server.postDirectMessagesHandler)
		})

		r.Route("/messages", func(r chi.Router) {
			r.Get("/", server.getGeneralMessagesHandler)
			r.Post("/", server.postGeneralMessagesHandler)
		})
	})

	return r
}

func (server *Server) registerHandler(w http.ResponseWriter, r *http.Request) {
	var input UserCredentials

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
	id, err := server.Users.CreateUser(domain.User{Username: input.Username,
		Password: input.Password})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	data, err := json.Marshal(map[string]interface{}{"id": id})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if _, err = w.Write(data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (server *Server) loginHandler(w http.ResponseWriter, r *http.Request) {
	var input UserCredentials

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
	token, err := server.Users.GenerateJWT(domain.User{Username: input.Username,
		Password: input.Password})
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

func (server *Server) authHandler(next http.Handler) http.Handler {
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

		userID, err := server.Users.ParseToken(headerParts[1])
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, ctxKey("id"), userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

func (server *Server) getDirectMessagesHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("id").(uint64)
	messageIDString := r.URL.Query().Get("messageID")
	offsetString := r.URL.Query().Get("offset")
	if offsetString == "" || messageIDString == "" {
		server.getDirectMessages(w, userID)
	} else {
		messageID, err := strconv.ParseUint(messageIDString, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		offset, err := strconv.ParseUint(offsetString, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		direct, err := server.Messages.ListDirect(userID, offset, messageID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		data, err := json.Marshal(direct)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, err = w.Write(data)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func (server *Server) getDirectMessages(w http.ResponseWriter, userID uint64) {
	direct, err := server.Messages.GetDirect(userID)
	if err != nil {
		_, err = w.Write([]byte("no messages are stored"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
	}
	data, err := json.Marshal(direct)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, err = w.Write(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (server *Server) getGeneralMessages(w http.ResponseWriter, userID uint64) {
	direct, err := server.Messages.GetGeneral()
	if err != nil {
		_, err = w.Write([]byte("no messages are stored"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
	}
	data, err := json.Marshal(direct)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, err = w.Write(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (server *Server) postDirectMessagesHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("id").(uint64)
	companionID, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if userID == companionID {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var input MessageCredentials

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
	message := domain.DirectMessage{
		GeneralMessage: domain.GeneralMessage{SenderID: userID, Content: input.Message},
		ReceiverID:     companionID}
	err = server.SendDirectMessage(message)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (server *Server) getGeneralMessagesHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("id").(uint64)
	messageIDString := r.URL.Query().Get("messageID")
	offsetString := r.URL.Query().Get("offset")
	if offsetString == "" || messageIDString == "" {
		server.getGeneralMessages(w, userID)
	} else {
		messageID, err := strconv.ParseUint(messageIDString, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		offset, err := strconv.ParseUint(offsetString, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		direct, err := server.Messages.ListGeneral(offset, messageID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		data, err := json.Marshal(direct)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, err = w.Write(data)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func (server *Server) postGeneralMessagesHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("id").(uint64)

	var input MessageCredentials

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
	message := domain.GeneralMessage{SenderID: userID, Content: input.Message}
	server.Messages.AddGeneral(message)
	w.WriteHeader(http.StatusCreated)
}
