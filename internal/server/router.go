package server

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/shevchukeugeni/gofermart/internal/auth"
	"github.com/shevchukeugeni/gofermart/internal/store"
	"golang.org/x/crypto/sha3"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth"
	"go.uber.org/zap"

	"github.com/shevchukeugeni/gofermart/internal/types"
)

type router struct {
	logger         *zap.Logger
	userRepo       store.User
	orderRepo      store.Order
	withdrawalRepo store.Withdrawal
}

func SetupRouter(logger *zap.Logger, user store.User, order store.Order, wtd store.Withdrawal) http.Handler {
	ro := &router{
		logger:         logger,
		userRepo:       user,
		orderRepo:      order,
		withdrawalRepo: wtd,
	}
	return ro.Handler()
}

func (ro *router) Handler() http.Handler {
	rtr := chi.NewRouter()
	rtr.Use(middleware.Logger)
	rtr.Post("/api/user/register", ro.register)
	rtr.Post("/api/user/login", ro.auth)
	rtr.Route("/api/user", func(r chi.Router) {
		r.Use(jwtauth.Verifier(auth.TokenAuth))
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
	var req types.UserLoginRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Unable to decode json: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Login == "" || req.Password == "" {
		http.Error(w, "Missing login or password.", http.StatusBadRequest)
		return
	}

	usr := req.User().ToDB()

	err = ro.userRepo.CreateUser(r.Context(), usr)
	if err != nil {
		if errors.Is(err, types.ErrUserAlreadyExists) {
			http.Error(w, "Unable to create user: "+err.Error(), http.StatusConflict)
		} else {
			http.Error(w, "Unable to create user: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	tokenString, err := auth.GenerateToken(usr.ID)
	if err != nil {
		http.Error(w, "Unable to generate token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", tokenString))
	w.WriteHeader(http.StatusOK)
}

func (ro *router) auth(w http.ResponseWriter, r *http.Request) {
	var req types.UserLoginRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Unable to decode json: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Login == "" || req.Password == "" {
		http.Error(w, "Missing login or password.", http.StatusBadRequest)
		return
	}

	usr, err := ro.userRepo.GetByLogin(r.Context(), req.Login)
	if err != nil {
		http.Error(w, "Unable to find user: "+err.Error(), http.StatusBadRequest)
		return
	}

	h := sha3.New512()
	h.Write([]byte(req.Password))

	if base64.StdEncoding.EncodeToString(h.Sum(nil)) != usr.Password {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	tokenString, err := auth.GenerateToken(usr.ID)
	if err != nil {
		http.Error(w, "Unable to generate token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", tokenString))
	w.WriteHeader(http.StatusOK)
}

func (ro *router) newOrder(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "text/plain" {
		http.Error(w, "incorrect request format", http.StatusBadRequest)
		return
	}

	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	numberB, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read body: "+err.Error(), http.StatusBadRequest)
		return
	}
	r.Body.Close()

	number := string(numberB)

	err = types.ValidateOrder(number)
	if err != nil {
		http.Error(w, "Order number validation failed", http.StatusUnprocessableEntity)
		return
	}

	err = ro.orderRepo.CreateOrder(r.Context(), number, userID)
	switch {
	case errors.Is(err, types.ErrOrderAlreadyCreatedByUser):
		w.WriteHeader(200)
		return
	case errors.Is(err, types.ErrOrderAlreadyCreatedByAnother):
		http.Error(w, err.Error(), http.StatusConflict)
		return
	case err == nil:
		w.WriteHeader(http.StatusAccepted)
		return
	default:
		http.Error(w, "Unable to create order: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (ro *router) orders(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	ret, err := ro.orderRepo.GetOrdersByUser(r.Context(), userID)
	if err != nil {
		http.Error(w, "Unable to get orders: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if len(ret) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		http.Error(w, "Can't marshal data: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (ro *router) balance(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	balance, err := ro.withdrawalRepo.GetBalance(r.Context(), userID)
	if err != nil {
		http.Error(w, "Unable to get balance: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(balance)
	if err != nil {
		http.Error(w, "Can't marshal data: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (ro *router) withdraw(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	var req types.WithdrawalRequest

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Unable to decode json: "+err.Error(), http.StatusBadRequest)
		return
	}

	err = types.ValidateOrder(req.Order)
	if err != nil || req.Sum == 0 {
		http.Error(w, "Order number validation failed", http.StatusUnprocessableEntity)
		return
	}

	err = ro.withdrawalRepo.CreateWithdrawal(r.Context(), req.Order, userID, req.Sum)
	switch {
	case errors.Is(err, types.ErrInsufficientBalance):
		w.WriteHeader(402)
		return
	case err == nil:
		w.WriteHeader(http.StatusOK)
		return
	default:
		http.Error(w, "Unable to create order: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (ro *router) withdrawalsList(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	ret, err := ro.withdrawalRepo.GetWithdrawalsByUser(r.Context(), userID)
	if err != nil {
		http.Error(w, "Unable to get orders: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if len(ret) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		http.Error(w, "Can't marshal data: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
