package googleservice

import (
	"context"
	"fmt"

	"github.com/saulo-duarte/chronos-lambda/internal/task"
	"golang.org/x/oauth2"
)

type GoogleEventHandler struct {
	oauthConfig *oauth2.Config
}

func NewGoogleEventHandler(oauthConfig *oauth2.Config) *GoogleEventHandler {
	return &GoogleEventHandler{
		oauthConfig: oauthConfig,
	}
}

func (h *GoogleEventHandler) HandleTaskEvent(ctx context.Context, t *task.Task, accessToken string) error {
	token := &oauth2.Token{AccessToken: accessToken, TokenType: "Bearer"}
	srv, err := NewGoogleCalendarService(ctx, token)
	if err != nil {
		return fmt.Errorf("falha ao criar serviço do Google Calendar: %w", err)
	}

	eventData := &TaskEventData{
		ID:          t.ID,
		Name:        t.Name,
		Description: t.Description,
		StartDate:   t.StartDate,
		DueDate:     t.DueDate,
		EventID:     t.GoogleCalendarEventId,
	}

	if t.Status == "DONE" {
		if eventData.EventID != "" {
			err := srv.DeleteEvent(ctx, eventData.EventID)
			if err != nil {
				return err
			}
		}
		t.GoogleCalendarEventId = ""
		return nil
	}

	if eventData.EventID != "" {
		return srv.UpdateEvent(ctx, eventData)
	}

	eventId, err := srv.CreateEvent(ctx, eventData)
	if err != nil {
		return err
	}
	t.GoogleCalendarEventId = eventId
	return nil
}
