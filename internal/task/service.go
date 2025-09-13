package task

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/saulo-duarte/chronos-lambda/internal/config"
)

var (
	ErrTaskNotFound = errors.New("task not found")
	ErrUnauthorized = errors.New("unauthorized")
)

type TaskService interface {
	CreateTask(ctx context.Context, t *Task) (*Task, error)
	GetTaskByID(ctx context.Context, id string) (*Task, error)
	ListTaskByUser(ctx context.Context, userID string) (*[]Task, error)
	ListTaskByProject(ctx context.Context, projectID string) (*[]Task, error)
	UpdateTask(ctx context.Context, t *Task) (*Task, error)
	DeleteTask(ctx context.Context, id string) error
}

type taskService struct {
	repo TaskRepository
}

func NewService(repo TaskRepository) TaskService {
	return &taskService{repo: repo}
}

func (s *taskService) CreateTask(ctx context.Context, t *Task) (*Task, error) {
	log := config.WithContext(ctx)

	t.ID = uuid.New()
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()

	if err := s.repo.Create(t); err != nil {
		log.WithError(err).Error("Falha ao criar task")
		return nil, err
	}

	log.WithField("task_id", t.ID).Info("Task criada com sucesso")
	return t, nil
}

func (s *taskService) GetTaskByID(ctx context.Context, id string) (*Task, error) {
	log := config.WithContext(ctx)

	taskID, err := uuid.Parse(id)
	if err != nil {
		log.WithError(err).Warn("ID inválido para busca da task")
		return nil, errors.New("invalid task id")
	}

	task, err := s.repo.GetByID(taskID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			log.WithField("task_id", id).Warn("Task não encontrada")
			return nil, ErrTaskNotFound
		}
		log.WithError(err).Error("Erro ao buscar task por ID")
		return nil, err
	}

	return task, nil
}

func (s *taskService) ListTaskByUser(ctx context.Context, userID string) (*[]Task, error) {
	log := config.WithContext(ctx)

	uid, err := uuid.Parse(userID)
	if err != nil {
		log.WithError(err).Warn("ID inválido para listar tasks por usuário")
		return nil, errors.New("invalid user id")
	}

	tasks, err := s.repo.ListByUser(uid)
	if err != nil {
		log.WithError(err).Error("Erro ao listar tasks por usuário")
		return nil, err
	}

	result := make([]Task, len(tasks))
	for i, t := range tasks {
		result[i] = *t
	}

	return &result, nil
}

func (s *taskService) ListTaskByProject(ctx context.Context, projectID string) (*[]Task, error) {
	log := config.WithContext(ctx)

	pid, err := uuid.Parse(projectID)
	if err != nil {
		log.WithError(err).Warn("ID inválido para listar tasks por projeto")
		return nil, errors.New("invalid project id")
	}

	tasks, err := s.repo.ListByProject(pid)
	if err != nil {
		log.WithError(err).Error("Erro ao listar tasks por projeto")
		return nil, err
	}

	result := make([]Task, len(tasks))
	for i, t := range tasks {
		result[i] = *t
	}

	return &result, nil
}

func (s *taskService) UpdateTask(ctx context.Context, t *Task) (*Task, error) {
	log := config.WithContext(ctx)

	existing, err := s.repo.GetByID(t.ID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			log.WithField("task_id", t.ID).Warn("Task não encontrada para atualização")
			return nil, ErrTaskNotFound
		}
		log.WithError(err).Error("Erro ao buscar task para atualização")
		return nil, err
	}

	existing.Name = t.Name
	existing.Description = t.Description
	existing.Status = t.Status
	existing.UpdatedAt = time.Now()

	if err := s.repo.Update(existing); err != nil {
		log.WithError(err).Error("Falha ao atualizar task")
		return nil, err
	}

	log.WithField("task_id", existing.ID).Info("Task atualizada com sucesso")
	return existing, nil
}

func (s *taskService) DeleteTask(ctx context.Context, id string) error {
	log := config.WithContext(ctx)

	taskID, err := uuid.Parse(id)
	if err != nil {
		log.WithError(err).Warn("ID inválido para deletar task")
		return errors.New("invalid task id")
	}

	_, err = s.repo.GetByID(taskID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			log.WithField("task_id", id).Warn("Task não encontrada para exclusão")
			return ErrTaskNotFound
		}
		log.WithError(err).Error("Erro ao buscar task para exclusão")
		return err
	}

	if err := s.repo.Delete(taskID); err != nil {
		log.WithError(err).Error("Falha ao excluir task")
		return err
	}

	log.WithField("task_id", id).Info("Task excluída com sucesso")
	return nil
}
