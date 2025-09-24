package task

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/saulo-duarte/chronos-lambda/internal/auth"
	"github.com/saulo-duarte/chronos-lambda/internal/config"
	"github.com/saulo-duarte/chronos-lambda/internal/project"
	studytopic "github.com/saulo-duarte/chronos-lambda/internal/study_topic"
	"github.com/saulo-duarte/chronos-lambda/internal/user"
	"github.com/sirupsen/logrus"
)

var (
	ErrTaskNotFound       = errors.New("task not found")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrProjectNotFound    = project.ErrProjectNotFound
	ErrStudyTopicNotFound = studytopic.ErrStudyTopicNotFound
)

type TaskService interface {
	CreateTask(ctx context.Context, t *Task) (*Task, error)
	FindAllByUser(ctx context.Context) ([]*Task, error)
	FindByID(ctx context.Context, id string) (*Task, error)
	DeleteByID(ctx context.Context, id string) error
	FindAllByProjectID(ctx context.Context, projectID string) ([]*Task, error)
	FindAllByTopicID(ctx context.Context, topicID string) ([]*Task, error)
	UpdateTask(ctx context.Context, t *Task) (*Task, error)
}

type taskService struct {
	repo           TaskRepository
	projectService project.ProjectService
	userRepo       user.UserRepository
	studyTopicRepo studytopic.StudyTopicRepository
	eventHandler   EventHandler
}

func NewService(repo TaskRepository, projectService project.ProjectService, userRepo user.UserRepository, studyTopicRepo studytopic.StudyTopicRepository, eventHandler EventHandler) TaskService {
	return &taskService{
		repo:           repo,
		projectService: projectService,
		userRepo:       userRepo,
		studyTopicRepo: studyTopicRepo,
		eventHandler:   eventHandler,
	}
}

func (s *taskService) CreateTask(ctx context.Context, t *Task) (*Task, error) {
	log := config.WithContext(ctx)
	claims, err := auth.GetUserClaimsFromContext(ctx)
	if err != nil {
		log.WithError(err).Warn("Attempt to create task without authentication")
		return nil, ErrUnauthorized
	}
	t.ID = uuid.New()
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()
	t.UserID = uuid.MustParse(claims.UserID)
	if t.Type == "PROJECT" && t.ProjectId == nil {
		return nil, errors.New("projectId is required for PROJECT tasks")
	}
	if t.ProjectId != nil {
		if _, err := s.projectService.GetProjectByID(ctx, t.ProjectId.String()); err != nil {
			log.WithError(err).WithFields(logrus.Fields{
				"project_id": t.ProjectId,
				"user_id":    t.UserID,
			}).Error("Project not found or does not belong to the user")
			return nil, ErrProjectNotFound
		}
	}
	if t.StudyTopicId != nil {
		if _, err := s.studyTopicRepo.GetByID(t.StudyTopicId.String()); err != nil {
			log.WithError(err).WithFields(logrus.Fields{
				"study_topic_id": *t.StudyTopicId,
				"user_id":        t.UserID,
			}).Error("Study topic not found or does not belong to the user")
			return nil, ErrStudyTopicNotFound
		}
	}
	if err := s.repo.Create(t); err != nil {
		log.WithError(err).Error("Failed to create task")
		return nil, err
	}
	if t.StartDate != nil || t.DueDate != nil {
		encryptedToken, err := s.userRepo.GetUserEncryptedGoogleCalendarAccessToken(claims.UserID)
		if err == nil && encryptedToken != "" {
			accessToken, decryptErr := config.Decrypt(encryptedToken)
			if decryptErr == nil {
				if err := s.eventHandler.HandleTaskEvent(ctx, t, accessToken); err != nil {
					log.WithError(err).WithField("task_id", t.ID).Error("Failed to create Google Calendar event")
				} else {
					if err := s.repo.Update(t); err != nil {
						log.WithError(err).WithField("task_id", t.ID).Error("Failed to update task with Google Calendar event ID")
					}
				}
			} else {
				log.WithError(decryptErr).WithField("user_id", claims.UserID).Error("Failed to decrypt Google Calendar access token")
			}
		}
	}
	log.WithField("task_id", t.ID).Info("Task created successfully")
	return t, nil
}

func (s *taskService) FindAllByUser(ctx context.Context) ([]*Task, error) {
	log := config.WithContext(ctx)
	claims, err := auth.GetUserClaimsFromContext(ctx)
	if err != nil {
		log.WithError(err).Warn("Attempt to list tasks without authentication")
		return nil, ErrUnauthorized
	}
	userID := uuid.MustParse(claims.UserID)
	tasks, err := s.repo.ListByUser(userID)
	if err != nil {
		log.WithError(err).Error("Failed to list tasks by user")
		return nil, err
	}
	return tasks, nil
}

func (s *taskService) FindByID(ctx context.Context, id string) (*Task, error) {
	log := config.WithContext(ctx)
	claims, err := auth.GetUserClaimsFromContext(ctx)
	if err != nil {
		log.WithError(err).Warn("Attempt to find task without authentication")
		return nil, ErrUnauthorized
	}
	taskID, err := uuid.Parse(id)
	if err != nil {
		log.WithError(err).Warn("Invalid task ID")
		return nil, errors.New("invalid task id")
	}
	userID := uuid.MustParse(claims.UserID)
	task, err := s.repo.FindByIdAndUserId(taskID, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			log.WithFields(logrus.Fields{
				"task_id": id,
				"user_id": userID,
			}).Warn("Task not found or does not belong to user")
			return nil, ErrTaskNotFound
		}
		log.WithError(err).Error("Error finding task by ID")
		return nil, err
	}
	return task, nil
}

func (s *taskService) DeleteByID(ctx context.Context, id string) error {
	log := config.WithContext(ctx)
	claims, err := auth.GetUserClaimsFromContext(ctx)
	if err != nil {
		log.WithError(err).Warn("Attempt to delete task without authentication")
		return ErrUnauthorized
	}
	taskID, err := uuid.Parse(id)
	if err != nil {
		log.WithError(err).Warn("Invalid task ID for deletion")
		return errors.New("invalid task id")
	}
	userID := uuid.MustParse(claims.UserID)
	task, err := s.repo.FindByIdAndUserId(taskID, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			log.WithFields(logrus.Fields{
				"task_id": id,
				"user_id": userID,
			}).Warn("Task not found or does not belong to user for deletion")
			return ErrTaskNotFound
		}
		log.WithError(err).Error("Error finding task before deletion")
		return err
	}
	if task.GoogleCalendarEventId != "" {
		encryptedToken, err := s.userRepo.GetUserEncryptedGoogleCalendarAccessToken(claims.UserID)
		if err == nil && encryptedToken != "" {
			accessToken, decryptErr := config.Decrypt(encryptedToken)
			if decryptErr == nil {
				err := s.eventHandler.HandleTaskEvent(ctx, task, accessToken)
				if err != nil {
					log.WithError(err).WithField("task_id", task.ID).Error("Failed to delete Google Calendar event")
				}
			} else {
				log.WithError(decryptErr).WithField("user_id", claims.UserID).Error("Failed to decrypt Google Calendar access token")
			}
		}
	}
	err = s.repo.Delete(taskID, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			log.WithFields(logrus.Fields{
				"task_id": id,
				"user_id": userID,
			}).Warn("Task not found or does not belong to user for deletion")
			return ErrTaskNotFound
		}
		log.WithError(err).Error("Failed to delete task")
		return err
	}
	log.WithField("task_id", id).Info("Task deleted successfully")
	return nil
}

func (s *taskService) FindAllByProjectID(ctx context.Context, projectID string) ([]*Task, error) {
	log := config.WithContext(ctx)
	claims, err := auth.GetUserClaimsFromContext(ctx)
	if err != nil {
		log.WithError(err).Warn("Attempt to list tasks by project without authentication")
		return nil, ErrUnauthorized
	}
	pid, err := uuid.Parse(projectID)
	if err != nil {
		log.WithError(err).Warn("Invalid project ID")
		return nil, errors.New("invalid project id")
	}
	userID := uuid.MustParse(claims.UserID)
	_, err = s.projectService.GetProjectByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, project.ErrProjectNotFound) {
			log.WithFields(logrus.Fields{
				"project_id": projectID,
				"user_id":    userID,
			}).Warn("Project not found or does not belong to user")
			return nil, ErrProjectNotFound
		}
		log.WithError(err).Error("Error finding project by ID")
		return nil, err
	}
	tasks, err := s.repo.ListByProjectAndUser(pid, userID)
	if err != nil {
		log.WithError(err).Error("Failed to list tasks by project")
		return nil, err
	}
	return tasks, nil
}

func (s *taskService) FindAllByTopicID(ctx context.Context, topicID string) ([]*Task, error) {
	log := config.WithContext(ctx)
	claims, err := auth.GetUserClaimsFromContext(ctx)
	if err != nil {
		log.WithError(err).Warn("Attempt to list tasks by topic without authentication")
		return nil, ErrUnauthorized
	}
	tid, err := uuid.Parse(topicID)
	if err != nil {
		log.WithError(err).Warn("Invalid study topic ID")
		return nil, errors.New("invalid study topic id")
	}
	userID := uuid.MustParse(claims.UserID)
	_, err = s.studyTopicRepo.GetByID(tid.String())
	if err != nil {
		if errors.Is(err, studytopic.ErrStudyTopicNotFound) {
			log.WithFields(logrus.Fields{
				"topic_id": topicID,
				"user_id":  userID,
			}).Warn("Study topic not found or does not belong to user")
			return nil, ErrStudyTopicNotFound
		}
		log.WithError(err).Error("Error finding study topic by ID")
		return nil, err
	}
	tasks, err := s.repo.ListByStudyTopicAndUser(tid, userID)
	if err != nil {
		log.WithError(err).Error("Failed to list tasks by study topic")
		return nil, err
	}
	return tasks, nil
}

func (s *taskService) UpdateTask(ctx context.Context, t *Task) (*Task, error) {
	log := config.WithContext(ctx)
	claims, err := auth.GetUserClaimsFromContext(ctx)
	if err != nil {
		log.WithError(err).Warn("Attempt to update task without authentication")
		return nil, ErrUnauthorized
	}
	userID := uuid.MustParse(claims.UserID)
	existing, err := s.repo.FindByIdAndUserId(t.ID, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			log.WithFields(logrus.Fields{
				"task_id": t.ID,
				"user_id": userID,
			}).Warn("Task not found for update")
			return nil, ErrTaskNotFound
		}
		log.WithError(err).Error("Error finding task for update")
		return nil, err
	}
	datesChanged := false
	if t.StartDate != nil && (existing.StartDate == nil || !t.StartDate.Equal(*existing.StartDate)) {
		existing.StartDate = t.StartDate
		datesChanged = true
	}
	if t.DueDate != nil && (existing.DueDate == nil || !t.DueDate.Equal(*existing.DueDate)) {
		existing.DueDate = t.DueDate
		datesChanged = true
	}
	if t.Name != "" {
		existing.Name = t.Name
	}
	if t.Description != "" {
		existing.Description = t.Description
	}
	if t.Status != "" {
		existing.Status = t.Status
	}
	if t.Priority != "" {
		existing.Priority = t.Priority
	}
	if !t.DoneAt.IsZero() {
		existing.DoneAt = t.DoneAt
	}
	existing.UpdatedAt = time.Now()
	if err := s.repo.Update(existing); err != nil {
		log.WithError(err).Error("Failed to update task")
		return nil, err
	}
	if datesChanged || existing.GoogleCalendarEventId == "" || existing.Status == "DONE" {
		encryptedToken, err := s.userRepo.GetUserEncryptedGoogleCalendarAccessToken(claims.UserID)
		if err == nil && encryptedToken != "" {
			accessToken, decryptErr := config.Decrypt(encryptedToken)
			if decryptErr == nil {
				err := s.eventHandler.HandleTaskEvent(ctx, existing, accessToken)
				if err != nil {
					log.WithError(err).WithField("task_id", existing.ID).Error("Failed to handle Google Calendar event")
				} else {
					if err := s.repo.Update(existing); err != nil {
						log.WithError(err).WithField("task_id", existing.ID).Error("Failed to update task with Google Calendar event ID")
					}
				}
			} else {
				log.WithError(decryptErr).WithField("user_id", claims.UserID).Error("Failed to decrypt Google Calendar access token")
			}
		}
	}
	log.WithField("task_id", existing.ID).Info("Task updated successfully")
	return existing, nil
}
