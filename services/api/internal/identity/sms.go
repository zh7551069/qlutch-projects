package identity

import (
	"context"
	"sync"
)

// SMS is one mocked text message. The code is not logged or returned by the API.
type SMS struct {
	RegionCode    string
	PhoneNational string
	Purpose       string
	Code          string
	Locale        string
}

// Sender delivers a code. Implementations must not call back into Service.
type Sender interface {
	Send(ctx context.Context, message SMS) error
}

// DiscardSender drops the code. It is the production mock.
type DiscardSender struct{}

// Send implements Sender.
func (DiscardSender) Send(context.Context, SMS) error { return nil }

// CaptureSender stores messages for tests in this process. It is not an HTTP route.
type CaptureSender struct {
	mu       sync.Mutex
	Messages []SMS
}

// Send implements Sender.
func (s *CaptureSender) Send(_ context.Context, message SMS) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Messages = append(s.Messages, message)
	return nil
}

// Codes returns the captured codes.
func (s *CaptureSender) Codes() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]string, len(s.Messages))
	for i, message := range s.Messages {
		out[i] = message.Code
	}
	return out
}
