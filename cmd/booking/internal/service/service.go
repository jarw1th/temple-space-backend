package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"templespace/cmd/booking/internal/domain"
)

type TokenVerifier interface {
	Verify(ctx context.Context, token string) (userID string, err error)
}

type PaymentGateway interface {
	Charge(ctx context.Context, bookingID string) error
}

type Service struct {
	repo      domain.BookingRepository
	readModel domain.ReadModel
	events    domain.EventPublisher
	auth      TokenVerifier
	payment   PaymentGateway
}

func New(repo domain.BookingRepository, readModel domain.ReadModel, events domain.EventPublisher, auth TokenVerifier, payment PaymentGateway) *Service {
	return &Service{repo: repo, readModel: readModel, events: events, auth: auth, payment: payment}
}

func (s *Service) CreateBooking(ctx context.Context, accessToken, spaceID, userID string, start, end time.Time) (*domain.Booking, error) {
	if _, err := s.auth.Verify(ctx, accessToken); err != nil {
		return nil, err
	}
	available, err := s.repo.IsAvailable(spaceID, start, end)
	if err != nil {
		return nil, err
	}
	if !available {
		return nil, errors.New("slot not available")
	}
	b := &domain.Booking{
		ID:        generateID(),
		SpaceID:   spaceID,
		UserID:    userID,
		SlotStart: start,
		SlotEnd:   end,
		Status:    domain.StatusPending,
		Version:   1,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if err := s.repo.Create(b); err != nil {
		return nil, err
	}
	_ = s.readModel.CacheAvailability(spaceID, start, end, false)
	payload, _ := json.Marshal(b)
	_ = s.events.Publish("booking_created", b.ID, payload)
	return b, nil
}

func (s *Service) ConfirmPayment(ctx context.Context, accessToken, bookingID string) (*domain.Booking, error) {
	if _, err := s.auth.Verify(ctx, accessToken); err != nil {
		return nil, err
	}
	b, err := s.repo.GetByID(bookingID)
	if err != nil {
		return nil, err
	}
	if b.Status == domain.StatusCancelled {
		return nil, errors.New("booking cancelled")
	}
	if err := s.payment.Charge(ctx, bookingID); err != nil {
		return nil, err
	}
	b.Status = domain.StatusPaid
	b.Version++
	b.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(b); err != nil {
		return nil, err
	}
	payload, _ := json.Marshal(b)
	_ = s.events.Publish("booking_paid", b.ID, payload)
	return b, nil
}

func (s *Service) CancelBooking(ctx context.Context, accessToken, bookingID string) (*domain.Booking, error) {
	if _, err := s.auth.Verify(ctx, accessToken); err != nil {
		return nil, err
	}
	b, err := s.repo.GetByID(bookingID)
	if err != nil {
		return nil, err
	}
	if b.Status == domain.StatusPaid {
		return nil, errors.New("cannot cancel paid booking")
	}
	b.Status = domain.StatusCancelled
	b.Version++
	b.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(b); err != nil {
		return nil, err
	}
	payload, _ := json.Marshal(b)
	_ = s.events.Publish("booking_cancelled", b.ID, payload)
	_ = s.readModel.CacheAvailability(b.SpaceID, b.SlotStart, b.SlotEnd, true)
	return b, nil
}

func generateID() string {
	// simple placeholder; replace with proper UUID
	return time.Now().UTC().Format("20060102150405.000000000")
}
