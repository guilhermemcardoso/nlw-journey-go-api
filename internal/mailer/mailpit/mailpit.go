package mailpit

import (
	"context"
	"fmt"
	"journey/internal/pgstore"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/wneessen/go-mail"
)

type store interface {
	GetTrip(context.Context, uuid.UUID) (pgstore.Trip, error)
}

type Mailpit struct {
	store store
}

func NewMailpit(pool *pgxpool.Pool) Mailpit {
	return Mailpit{pgstore.New(pool)}
}

func (m Mailpit) SendCOnfirmTripEmailToTripOwner(tripID uuid.UUID) error {
	ctx := context.Background()
	trip, err := m.store.GetTrip(ctx, tripID)

	if err != nil {
		return fmt.Errorf("mailpit: failed to get trip for SendCOnfirmTripEmailToTripOwner: %w", err)
	}

	msg := mail.NewMsg()
	if err := msg.From("mailpit@journey.com"); err != nil {
		return fmt.Errorf("mailpit: failed to set From address for SendCOnfirmTripEmailToTripOwner: %w", err)
	}

	if err := msg.To(trip.OwnerEmail); err != nil {
		return fmt.Errorf("mailpit: failed to set To address for SendCOnfirmTripEmailToTripOwner: %w", err)
	}

	msg.Subject("Trip confirmation")
	msg.SetBodyString(mail.TypeTextPlain, fmt.Sprintf(`
        Olá %s!

        A sua viagem para %s que começa no dia %s precisa ser confirmada.

        Clique no botão abaixo para confirmar.
        `,
		trip.OwnerName, trip.Destination, trip.StartsAt.Time.Format(time.DateOnly),
	))

	client, err := mail.NewClient("mailpit", mail.WithTLSPortPolicy(mail.NoTLS), mail.WithPort(1025))

	if err != nil {
		return fmt.Errorf("mailpit: failed to create client for SendCOnfirmTripEmailToTripOwner: %w", err)
	}

	if err := client.DialAndSend(msg); err != nil {
		return fmt.Errorf("mailpit: failed to send email for SendCOnfirmTripEmailToTripOwner: %w", err)
	}

	return nil
}
