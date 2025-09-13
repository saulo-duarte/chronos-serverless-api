package task

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/saulo-duarte/chronos-lambda/internal/config"
)

type Handler struct {
	service TaskService
}

func NewHandler(s TaskService) *Handler {
	return &Handler{service: s}
}

func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	log := config.WithContext(r.Context())

	var payload Task
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.WithError(err).Error("Corpo da requisição inválido")
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	task, err := h.service.CreateTask(r.Context(), &payload)
	if err != nil {
		log.WithError(err).Error("Falha ao criar task")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	config.JSON(w, http.StatusCreated, task)
}

func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	log := config.WithContext(r.Context())

	var payload struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.ID == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	task, err := h.service.GetTaskByID(r.Context(), payload.ID)
	if err != nil {
		if err == ErrTaskNotFound {
			http.Error(w, "task not found", http.StatusNotFound)
			return
		}
		log.WithError(err).Error("Erro ao buscar task")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	config.JSON(w, http.StatusOK, task)
}

func (h *Handler) ListTasksByUser(w http.ResponseWriter, r *http.Request) {
	log := config.WithContext(r.Context())

	var payload struct {
		UserID string `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.UserID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	tasks, err := h.service.ListTaskByUser(r.Context(), payload.UserID)
	if err != nil {
		log.WithError(err).Error("Erro ao listar tasks por usuário")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	config.JSON(w, http.StatusOK, tasks)
}

func (h *Handler) ListTasksByProject(w http.ResponseWriter, r *http.Request) {
	log := config.WithContext(r.Context())

	var payload struct {
		ProjectID string `json:"project_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.ProjectID == "" {
		http.Error(w, "project_id is required", http.StatusBadRequest)
		return
	}

	tasks, err := h.service.ListTaskByProject(r.Context(), payload.ProjectID)
	if err != nil {
		log.WithError(err).Error("Erro ao listar tasks por projeto")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	config.JSON(w, http.StatusOK, tasks)
}

func (h *Handler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	log := config.WithContext(r.Context())

	var payload Task
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.ID == uuid.Nil {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	task, err := h.service.UpdateTask(r.Context(), &payload)
	if err != nil {
		if err == ErrTaskNotFound {
			http.Error(w, "task not found", http.StatusNotFound)
			return
		}
		log.WithError(err).Error("Erro ao atualizar task")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	config.JSON(w, http.StatusOK, task)
}

func (h *Handler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	log := config.WithContext(r.Context())

	var payload struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.ID == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteTask(r.Context(), payload.ID); err != nil {
		if err == ErrTaskNotFound {
			http.Error(w, "task not found", http.StatusNotFound)
			return
		}
		log.WithError(err).Error("Erro ao excluir task")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	config.JSON(w, http.StatusOK, map[string]string{
		"message": "task deleted successfully",
	})
}
