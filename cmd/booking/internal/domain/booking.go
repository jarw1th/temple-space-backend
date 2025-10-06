package domain

import "time"

type BookingStatus string

const (
	StatusPending   BookingStatus = "pending"
	StatusConfirmed BookingStatus = "confirmed"
	StatusPaid      BookingStatus = "paid"
	StatusCancelled BookingStatus = "cancelled"
)

type Booking struct {
	ID        string
	SpaceID   string
	UserID    string
	SlotStart time.Time
	SlotEnd   time.Time
	Status    BookingStatus
	Version   int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type BookingRepository interface {
	Create(b *Booking) error
	Update(b *Booking) error
	GetByID(id string) (*Booking, error)
	IsAvailable(spaceID string, start, end time.Time) (bool, error)
}

type ReadModel interface {
	CacheAvailability(spaceID string, start, end time.Time, available bool) error
}

type EventPublisher interface {
	Publish(topic string, key string, payload []byte) error
}
