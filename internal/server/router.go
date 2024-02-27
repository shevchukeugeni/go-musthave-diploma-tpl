package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth"
	"go.uber.org/zap"

	"github.com/shevchukeugeni/gofermart/internal/types"
)

var tokenAuth *jwtauth.JWTAuth

const Secret = "secret-gofermart"

func init() {
	tokenAuth = jwtauth.New("HS256", []byte(Secret), nil)
}

type router struct {
	logger *zap.Logger
}

func SetupRouter(logger *zap.Logger) http.Handler {
	ro := &router{
		logger: logger,
	}
	return ro.Handler()
}

func (ro *router) Handler() http.Handler {
	rtr := chi.NewRouter()
	rtr.Use(middleware.Logger)
	rtr.Post("/api/user/register", ro.register)
	rtr.Post("/api/user/login", ro.auth)
	rtr.Route("/api/user", func(r chi.Router) {
		r.Use(jwtauth.Verifier(tokenAuth))
		r.Use(jwtauth.Authenticator)
		r.Post("/orders", ro.newOrder)
		r.Get("/orders", ro.orders)
		r.Get("/balance", ro.balance)
		r.Post("/balance/withdraw", ro.withdraw)
		r.Get("/withdrawals", ro.withdrawalsList)
	})
	return rtr
}

func (ro *router) register(w http.ResponseWriter, r *http.Request) {
	panic("implement me")
}

func (ro *router) auth(w http.ResponseWriter, r *http.Request) {
	var req types.UserLoginRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Unable to decode json: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Login == "" || req.Password == "" {
		http.Error(w, "Missing username or password.", http.StatusBadRequest)
		return
	}

	_, tokenString, err := tokenAuth.Encode(map[string]interface{}{"username": req.Login})
	if err != nil {
		http.Error(w, "Unable to encode: "+err.Error(), http.StatusUnprocessableEntity)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(map[string]string{"access_token": tokenString})
	if err != nil {
		http.Error(w, "Can't marshal data: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (ro *router) newOrder(w http.ResponseWriter, r *http.Request) {
	panic("implement me")
}

func (ro *router) orders(w http.ResponseWriter, r *http.Request) {
	//panic("implement me")
	fmt.Println("1")
}

func (ro *router) balance(w http.ResponseWriter, r *http.Request) {
	panic("implement me")
}

func (ro *router) withdraw(w http.ResponseWriter, r *http.Request) {
	panic("implement me")
}

func (ro *router) withdrawalsList(w http.ResponseWriter, r *http.Request) {
	panic("implement me")
}
