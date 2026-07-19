package events

import "context"

type EventType string

const (
	EventBookCreated EventType = "book.created"
	EventBookUpdated EventType = "book.updated"
	EventBookDeleted EventType = "book.deleted"
	EventLoanIssued  EventType = "loan.issued"
	EventLoanReturned EventType = "loan.returned"
)

type Event struct {
	Type      EventType
	Payload   any
	Timestamp int64
}

type Publisher interface {
	Publish(ctx context.Context, event Event) error
}

type NoopPublisher struct{}

func NewNoopPublisher() *NoopPublisher {
	return &NoopPublisher{}
}

func (p *NoopPublisher) Publish(ctx context.Context, event Event) error {
	_ = ctx
	_ = event
	return nil
}
