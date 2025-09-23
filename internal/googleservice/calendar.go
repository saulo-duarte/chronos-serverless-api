package googleservice

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/saulo-duarte/chronos-lambda/internal/config"
	"github.com/saulo-duarte/chronos-lambda/internal/util"
	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

const timeZone = "America/Sao_Paulo"

type TaskEventData struct {
	ID          uuid.UUID
	Name        string
	Description string
	StartDate   *util.LocalDateTime
	DueDate     *util.LocalDateTime
	EventID     string
}

type GoogleCalendarService struct {
	CalendarService *calendar.Service
}

func NewGoogleCalendarService(ctx context.Context, token *oauth2.Token) (*GoogleCalendarService, error) {
	log := config.WithContext(ctx)
	log.WithFields(map[string]interface{}{
		"token_type":         token.TokenType,
		"token_value_length": len(token.AccessToken),
	}).Info("Criando novo cliente do Google Calendar")

	tokenSource := oauth2.StaticTokenSource(token)
	client := oauth2.NewClient(ctx, tokenSource)
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.WithError(err).Error("Falha ao criar o cliente do Google Calendar")
		return nil, fmt.Errorf("falha ao criar o cliente do Google Calendar: %w", err)
	}
	log.Info("Cliente do Google Calendar criado com sucesso")
	return &GoogleCalendarService{
		CalendarService: srv,
	}, nil
}

func (s *GoogleCalendarService) CreateEvent(ctx context.Context, t *TaskEventData) (string, error) {
	log := config.WithContext(ctx)
	log.WithField("task_data", t).Info("Tentando criar evento no Google Calendar com os seguintes dados da tarefa")

	event := s.createCalendarEventFromTaskData(t)
	log.WithField("calendar_event_data", event).Info("Evento do Google Calendar preparado para inserção")

	newEvent, err := s.CalendarService.Events.Insert("primary", event).Do()
	if err != nil {
		log.WithError(err).WithField("task_id", t.ID).Error("Falha ao criar evento")
		return "", fmt.Errorf("falha ao inserir evento: %w", err)
	}
	log.WithFields(map[string]interface{}{
		"task_id":         t.ID,
		"event_id":        newEvent.Id,
		"event_html_link": newEvent.HtmlLink,
	}).Info("Evento criado com sucesso no Google Calendar")
	return newEvent.Id, nil
}

func (s *GoogleCalendarService) UpdateEvent(ctx context.Context, t *TaskEventData) error {
	log := config.WithContext(ctx)
	log.WithField("task_data", t).Info("Tentando atualizar evento no Google Calendar com os seguintes dados da tarefa")

	event := s.createCalendarEventFromTaskData(t)
	log.WithField("calendar_event_data", event).Info("Evento do Google Calendar preparado para atualização")

	_, err := s.CalendarService.Events.Update("primary", t.EventID, event).Do()
	if err != nil {
		log.WithError(err).WithFields(map[string]interface{}{
			"task_id":  t.ID,
			"event_id": t.EventID,
		}).Error("Falha ao atualizar evento")
		return fmt.Errorf("falha ao atualizar evento: %w", err)
	}
	log.WithFields(map[string]interface{}{
		"task_id":  t.ID,
		"event_id": t.EventID,
	}).Info("Evento atualizado com sucesso no Google Calendar")
	return nil
}

func (s *GoogleCalendarService) DeleteEvent(ctx context.Context, eventId string) error {
	log := config.WithContext(ctx)
	log.WithField("event_id", eventId).Info("Tentando deletar evento do Google Calendar")

	err := s.CalendarService.Events.Delete("primary", eventId).Do()
	if err != nil {
		if e, ok := err.(*googleapi.Error); ok && e.Code == 404 {
			log.WithField("event_id", eventId).Warn("Evento não encontrado, mas a deleção é considerada bem-sucedida")
			return nil
		}
		log.WithError(err).WithField("event_id", eventId).Error("Falha ao deletar evento")
		return fmt.Errorf("falha ao deletar evento: %w", err)
	}
	log.WithField("event_id", eventId).Info("Evento deletado com sucesso do Google Calendar")
	return nil
}

func (s *GoogleCalendarService) createCalendarEventFromTaskData(t *TaskEventData) *calendar.Event {
	var start, end *calendar.EventDateTime
	if t.StartDate != nil || t.DueDate != nil {
		var startTime, endTime time.Time
		if t.StartDate != nil {
			startTime = t.StartDate.Time
		} else {
			startTime = t.DueDate.Time
		}
		if t.DueDate != nil {
			endTime = t.DueDate.Time
		} else {
			endTime = startTime.Add(1 * time.Hour)
		}
		start = s.toGoogleDateTime(startTime)
		end = s.toGoogleDateTime(endTime)
	}
	return &calendar.Event{
		Summary:     t.Name,
		Description: t.Description,
		Start:       start,
		End:         end,
	}
}

func (s *GoogleCalendarService) toGoogleDateTime(dateTime time.Time) *calendar.EventDateTime {
	return &calendar.EventDateTime{
		DateTime: dateTime.Format(time.RFC3339),
		TimeZone: timeZone,
	}
}
