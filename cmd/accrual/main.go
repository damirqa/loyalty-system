package accrual

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type Response struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}

func main() {
	http.HandleFunc("/api/orders/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) < 4 {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		orderNumber := parts[3]
		resp := Response{
			Order:   orderNumber,
			Status:  "PROCESSED",
			Accrual: 100.0,
		}
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	log.Println("Accrual system running on :8081")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatal(err)
	}
}
