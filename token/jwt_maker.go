package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const minSecretKeySize = 32

// JWTMaker is a JSON web token maker
type JWTMaker struct {
	secretKey string
}

// NewJWTMaker creates a new JWTMaker
func NewJWTMaker(secretKey string) (Maker, error) {
	if len(secretKey) < minSecretKeySize {
		return nil, fmt.Errorf("invalid key size: must be at least %d characters", minSecretKeySize)
	}

	return &JWTMaker{secretKey}, nil
}

// CreateToken creates a new token for a specific username and duration
func (maker *JWTMaker) CreateToken(username string, duration time.Duration) (string, error) {
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", err
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	return jwtToken.SignedString([]byte(maker.secretKey))
}

// VerifyToken checks if the token is valid or not
// If the token will be valid this fuction will return the Payload data stored inside the token
func (maker *JWTMaker) VerifyToken(token string) (*Payload, error) {
	// The function that receives the parsed but unverifed token, you should verify it is header to make sure that the signing algorithm
	// maches with what you normaly use to sign the token, then if header algo from request
	// will match with the algorithm which you normaly use to sign the token, you will return the Key
	// which will be than used to verify the token this will prevent the trivial atack mechanism
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, ErrInvalidToken
		}
		return []byte(maker.secretKey), nil
	}
	jwtToken, err := jwt.ParseWithClaims(token, &Payload{}, keyFunc)
	// if here the error is not equal to nil
	// than this can be from two reasons, first the token is invalid and the second one it is expired
	if err != nil {
		// jwt package pod povrchem už automaticky testuje jestli je token expired
		// ale výsledek skrývá uložením do svého vlastního objektu ValidationError
		// takže abych ověřil o který, ze dvou možných errorů se jedná musím objekt z err prvně vytáhnout
		// a přeuložit do proměnné verr
		// potom ověřím, že se mi ho opravu podařilo vytáhnout tzn. ok musí být true a zároven chci, aby vnitřní error
		// v tom objektu byl typu ErrExpiredToken, což ověřím pomocí funkce Is() z package errors.
		verr, ok := err.(*jwt.ValidationError)
		if ok && errors.Is(verr.Inner, ErrExpiredToken) {
			return nil, ErrExpiredToken
		}
		// pokud token není expired může být už jedině Invalit tzn vrátím ErrInvalidToken
		return nil, ErrInvalidToken
	}

	// pokud je vše v pohodě přečtu z ověřeného jwtTokenu Claims a z nich Payload uvnitř
	// převedu do struktury kterou vím, že mnou vytvořený Payload musí mít to znamená Payload struktura
	payload, ok := jwtToken.Claims.(*Payload)
	if !ok {
		// pokud se mi nepodaří dostat payload ověřeného tokenu do mé struktury pro Payload, bud to není můj token,
		// nebo je špatně něco jiného

		return nil, ErrInvalidToken
	}

	return payload, nil
}
