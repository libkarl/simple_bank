package token

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Different types of error returned by the VerifyToken function
var (
	ErrInvalidToken = errors.New("token is invalid") //
	ErrExpiredToken = errors.New("token has expired")
)

// Payload contains the payload data of the token
type Payload struct {
	// This field is used to invalidate specific token in case it is leaked
	// this field will be populate with specific id for each token
	ID uuid.UUID `json:"id"`
	// This field will be used to identify token owner
	Username string `json:"username"`
	// This filed is time when the token was created.
	IssuedAt time.Time `json:"issued_at"`
	// It is absolutely neccessary to create just tokens with sort expiration
	// to avoid problems-> ExpiredAt is field with record where is expired time for each token saved
	ExpiredAt time.Time `json:"expired_at"`
}

// NewPayload creates a new token payload with specific username and duration

func NewPayload(username string, duration time.Duration) (*Payload, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	payload := &Payload{
		ID:        tokenID,
		Username:  username,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(duration),
	}

	return payload, nil
}

// Valid checks if the token payload is valid or not
func (payload *Payload) Valid() error {
	if time.Now().After(payload.ExpiredAt) {
		return ErrExpiredToken
	}

	return nil
}
