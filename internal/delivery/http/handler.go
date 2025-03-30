package http

import (
	"damirqa/loyalty-system/internal/service"
	"encoding/json"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
	"io"
	"net/http"
)

type APIHandler struct {
	authService    service.AuthService
	orderService   service.OrderService
	balanceService service.BalanceService
	accrualClient  service.AccrualClient
	logger         *zap.Logger
}

func NewAPIHandler(auth service.AuthService, order service.OrderService, balance service.BalanceService, accrual service.AccrualClient, logger *zap.Logger) *APIHandler {
	return &APIHandler{
		authService:    auth,
		orderService:   order,
		balanceService: balance,
		accrualClient:  accrual,
		logger:         logger,
	}
}

func (h *APIHandler) RegisterRoutes(r *chi.Mux) {
	r.Post("/api/user/register", h.Register)
	r.Post("/api/user/login", h.Login)
	r.Post("/api/user/orders", h.SubmitOrder)
	r.Get("/api/user/orders", h.ListOrders)
	r.Get("/api/user/balance", h.GetBalance)
	r.Post("/api/user/balance/withdraw", h.Withdraw)
	r.Get("/api/user/withdrawals", h.ListWithdrawals)
}

func (h *APIHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.authService.Register(r.Context(), req.Login, req.Password)
	if err != nil {
		if err.Error() == "user already exists" {
			http.Error(w, "User already exists", http.StatusConflict)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *APIHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.authService.Login(r.Context(), req.Login, req.Password)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *APIHandler) SubmitOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	orderNumber := string(body)
	err = h.orderService.SubmitOrder(r.Context(), userID, orderNumber)
	if err != nil {
		switch err.Error() {
		case "order already submitted by user":
			w.WriteHeader(http.StatusOK)
		case "order already submitted by another user":
			http.Error(w, "order already submitted by another user", http.StatusConflict)
		case "invalid order number":
			http.Error(w, "invalid order number", http.StatusUnprocessableEntity)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (h *APIHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	orders, err := h.orderService.ListOrders(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	err = json.NewEncoder(w).Encode(orders)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *APIHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	balance, err := h.balanceService.GetBalance(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(balance)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *APIHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Order string  `json:"order"`
		Sum   float64 `json:"sum"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := h.balanceService.Withdraw(r.Context(), userID, req.Order, req.Sum)
	if err != nil {
		if err.Error() == "insufficient funds" {
			http.Error(w, "insufficient funds", http.StatusPaymentRequired)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}

func (h *APIHandler) ListWithdrawals(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	withdrawals, err := h.balanceService.ListWithdrawals(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	err = json.NewEncoder(w).Encode(withdrawals)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
