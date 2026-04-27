package handlers

import (
	"encoding/json"
	"net/http"
	"watchdog/internal/servise"
	"watchdog/internal/storage"
)

type SitesRequest struct {
	URL string `json:"url"`
}

type Handler struct {
	storage       *storage.Storage
	servicseCheck *servise.Service
}

func NewHandlers(storage *storage.Storage, service *servise.Service) *Handler {
	return &Handler{
		storage:       storage,
		servicseCheck: service,
	}
}

func (h *Handler) Sites(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		sites, err := h.storage.GetAllSites()
		if err != nil {
			http.Error(w, `{"error": "error database"}`, http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"sites": sites})
	case http.MethodPost:
		var req SitesRequest

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "Неверный формат json", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		if req.URL == "" {
			http.Error(w, "url должен быть обязательным", http.StatusBadRequest)
			return
		}

		err = h.storage.SaveSite(req.URL)
		if err != nil {
			http.Error(w, `{"error": "ошибка при сохранении в БД"}`, http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "Сайт успешно добавлен"})
	default:
		http.Error(w, `{"error": "Метод не поддерживается"}`, http.StatusMethodNotAllowed)
		return
	}
}

func (h *Handler) CheckStatiscs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		paramUrl := r.URL.Query().Get("url")
		statistic, err := h.servicseCheck.GetCheckStatistic(paramUrl)
		if err != nil {
			http.Error(w, `{"error": "Ошибка обработки статистики"}`, http.StatusBadRequest)
		}

		json.NewEncoder(w).Encode(map[string]any{"statistic": statistic})

	default:
		http.Error(w, `{"error": "Метод не поддерживается"}`, http.StatusMethodNotAllowed)
	}
}
