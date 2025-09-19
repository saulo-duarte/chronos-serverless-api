package googleservice

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/saulo-duarte/chronos-lambda/internal/task"
	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

const timeZone = "America/Sao_Paulo"

type GoogleCalendarService struct {
	CalendarService *calendar.Service
}

func NewGoogleCalendarService(token *oauth2.Token) (*GoogleCalendarService, error) {
	ctx := context.Background()
	tokenSource := oauth2.StaticTokenSource(token)
	client := oauth2.NewClient(ctx, tokenSource)

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("falha ao criar o cliente do Google Calendar: %w", err)
	}

	return &GoogleCalendarService{
		CalendarService: srv,
	}, nil
}

func (s *GoogleCalendarService) CreateEvent(t *task.Task) (string, error) {
	event := s.createCalendarEventFromTask(t)
	newEvent, err := s.CalendarService.Events.Insert("primary", event).Do()
	if err != nil {
		slog.Error("Falha ao criar evento", "task_id", t.ID, "error", err)
		return "", fmt.Errorf("falha ao inserir evento: %w", err)
	}
	slog.Info("Evento criado com sucesso no Google Calendar", "task_id", t.ID, "event_id", newEvent.Id)
	return newEvent.Id, nil
}

func (s *GoogleCalendarService) UpdateOrCreateEvent(t *task.Task) (string, error) {
	if t.GoogleCalendarEventId == "" {
		slog.Warn("ID de evento do Google Calendar não encontrado, criando um novo evento", "task_id", t.ID)
		return s.CreateEvent(t)
	}

	event := s.createCalendarEventFromTask(t)
	_, err := s.CalendarService.Events.Update("primary", t.GoogleCalendarEventId, event).Do()

	if err != nil {
		if e, ok := err.(*googleapi.Error); ok && e.Code == 404 {
			slog.Warn("Evento não encontrado no Google Calendar, criando um novo evento", "task_id", t.ID, "event_id", t.GoogleCalendarEventId)
			return s.CreateEvent(t)
		}
		slog.Error("Falha ao atualizar evento", "task_id", t.ID, "event_id", t.GoogleCalendarEventId, "error", err)
		return "", fmt.Errorf("falha ao atualizar evento: %w", err)
	}

	slog.Info("Evento atualizado com sucesso no Google Calendar", "task_id", t.ID, "event_id", t.GoogleCalendarEventId)
	return t.GoogleCalendarEventId, nil
}

func (s *GoogleCalendarService) DeleteEvent(eventId string) error {
	err := s.CalendarService.Events.Delete("primary", eventId).Do()
	if err != nil {
		if e, ok := err.(*googleapi.Error); ok && e.Code == 404 {
			slog.Warn("Evento não encontrado, mas a deleção é considerada bem-sucedida", "event_id", eventId)
			return nil
		}
		slog.Error("Falha ao deletar evento", "event_id", eventId, "error", err)
		return fmt.Errorf("falha ao deletar evento: %w", err)
	}
	slog.Info("Evento deletado com sucesso no Google Calendar", "event_id", eventId)
	return nil
}

func (s *GoogleCalendarService) createCalendarEventFromTask(t *task.Task) *calendar.Event {
	var start, end *calendar.EventDateTime

	if t.StartDate != nil || t.DueDate != nil {
		var startTime, endTime time.Time
		if t.StartDate != nil {
			startTime = *t.StartDate
		} else {
			startTime = *t.DueDate
		}

		if t.DueDate != nil {
			endTime = *t.DueDate
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

func HandleTaskEvent(ctx context.Context, t *task.Task, accessToken string) (uuid.UUID, error) {
	token := &oauth2.Token{AccessToken: accessToken}

	srv, err := NewGoogleCalendarService(token)
	if err != nil {
		return uuid.Nil, fmt.Errorf("falha ao criar o serviço do Google Calendar: %w", err)
	}

	if t.Status == task.DONE {
		if t.GoogleCalendarEventId != "" {
			err := srv.DeleteEvent(t.GoogleCalendarEventId)
			if err != nil {
				return uuid.Nil, fmt.Errorf("falha ao deletar evento do Google Calendar: %w", err)
			}
		}
		return t.ID, nil
	}

	eventId, err := srv.UpdateOrCreateEvent(t)
	if err != nil {
		return uuid.Nil, fmt.Errorf("falha ao criar/atualizar evento do Google Calendar: %w", err)
	}

	t.GoogleCalendarEventId = eventId
	return t.ID, nil
}
