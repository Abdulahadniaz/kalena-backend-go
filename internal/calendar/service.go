package calendar

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type Service struct {
	service *calendar.Service
}

func NewCalendarService(client *http.Client) (*Service, error) {
	ctx := context.Background()
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Calendar client: %v", err)
	}
	return &Service{service: srv}, nil
}

func (s *Service) GetUpcomingEvents(maxResults int) ([]Event, error) {
	t := time.Now().Format(time.RFC3339)
	events, err := s.service.Events.List("primary").
		ShowDeleted(false).
		SingleEvents(true).
		TimeMin(t).
		MaxResults(int64(maxResults)).
		OrderBy("startTime").
		Do()

	if err != nil {
		return nil, fmt.Errorf("unable to retrieve events: %v", err)
	}

	var calEvents []Event
	for _, item := range events.Items {
		start := item.Start.DateTime
		if start == "" {
			start = item.Start.Date
		}

		end := item.End.DateTime
		if end == "" {
			end = item.End.Date
		}

		calEvents = append(calEvents, Event{
			ID:       item.Id,
			Summary:  item.Summary,
			Start:    start,
			End:      end,
			Location: item.Location,
		})
	}

	return calEvents, nil
}
