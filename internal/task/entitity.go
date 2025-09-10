package task

import (
	"time"

	"github.com/google/uuid"
	"github.com/saulo-duarte/chronos-lambda/internal/project"
	"github.com/saulo-duarte/chronos-lambda/internal/user"
)

type Task struct {
	ID                    uuid.UUID       `gorm:"type:uuid;default:uuid_generate_v4()" json:"id"`
	GoogleCalendarEventId string          `json:"google_calendar_event_id"`
	Name                  string          `json:"name"`
	Description           string          `json:"description"`
	Status                TaskStatus      `json:"status"`
	Type                  TaskType        `json:"type"`
	Priority              TaskPriority    `json:"priority"`
	StartDate             *time.Time      `json:"start_date"`
	DueDate               *time.Time      `json:"due_date"`
	ProjectId             uuid.UUID       `json:"project_id"`
	Project               project.Project `gorm:"foreignKey:ProjectId" json:"project"`
	UserID                uuid.UUID       `gorm:"column:user_id;not null" json:"user_id"`
	User                  user.User       `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
	DoneAt                *time.Time      `json:"done_at"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
}
