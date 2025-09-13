package task

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("task not found")

type TaskRepository interface {
	Create(t *Task) error
	GetByID(id uuid.UUID) (*Task, error)
	ListByUser(userID uuid.UUID) ([]*Task, error)
	ListByProject(projectID uuid.UUID) ([]*Task, error)
	Update(t *Task) error
	Delete(id uuid.UUID) error
}

type taskRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) Create(t *Task) error {
	return r.db.Create(t).Error
}

func (r *taskRepository) GetByID(id uuid.UUID) (*Task, error) {
	var t Task
	if err := r.db.First(&t, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (r *taskRepository) ListByUser(userID uuid.UUID) ([]*Task, error) {
	var tasks []*Task
	if err := r.db.Where("user_id = ?", userID).Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *taskRepository) ListByProject(projectID uuid.UUID) ([]*Task, error) {
	var tasks []*Task
	if err := r.db.Where("project_id = ?", projectID).Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *taskRepository) Update(t *Task) error {
	return r.db.Model(&Task{}).Where("id = ?", t.ID).Updates(t).Error
}

func (r *taskRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&Task{}, "id = ?", id).Error
}
