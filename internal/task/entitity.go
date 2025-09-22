package task

import (
	"time"

	"github.com/google/uuid"
	"github.com/saulo-duarte/chronos-lambda/internal/project"
	studytopic "github.com/saulo-duarte/chronos-lambda/internal/study_topic"
	"github.com/saulo-duarte/chronos-lambda/internal/user"
	"github.com/saulo-duarte/chronos-lambda/internal/util"
)

type Task struct {
	ID                    uuid.UUID             `gorm:"type:uuid;default:uuid_generate_v4()" json:"id"`
	GoogleCalendarEventId string                `json:"googleCalendarEventId"`
	Name                  string                `json:"name"`
	Description           string                `json:"description"`
	Status                TaskStatus            `json:"status"`
	Type                  TaskType              `json:"type"`
	Priority              TaskPriority          `json:"priority"`
	StartDate             *util.LocalDateTime   `json:"startDate"`
	DueDate               *util.LocalDateTime   `json:"dueDate"`
	ProjectId             *uuid.UUID            `json:"projectId"`
	Project               project.Project       `gorm:"foreignKey:ProjectId" json:"project"`
	StudyTopicId          *uuid.UUID            `json:"studyTopicId"`
	StudyTopic            studytopic.StudyTopic `gorm:"foreignKey:StudyTopicId" json:"studyTopic"`
	UserID                uuid.UUID             `gorm:"column:user_id;not null" json:"userId"`
	User                  user.User             `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
	DoneAt                *util.LocalDateTime   `json:"doneAt"`
	CreatedAt             time.Time             `json:"createdAt"`
	UpdatedAt             time.Time             `json:"updatedAt"`
}
