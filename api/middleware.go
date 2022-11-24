package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/karlib/simple_bank/token"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
	// použité pro uložení zíkaného payloadu do contextu před zavoláním Next() funkce
	// jelikož v kontextu se hodnoty ukládají jako key value pairs
	authorizationPayloadKey = "authorization_payload"
)

// this is not middleware it is just higher order function which will returns middleware
func authMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
	// this anonymous function inside is actualy the middleware
	return func(ctx *gin.Context) {
		// vytáhne z header req položku s názvem "authorization"
		authorizationHeader := ctx.GetHeader(authorizationHeaderKey)
		// pokud není authorizationHeader poskytnut zabije kontext a odpoví s 401 (unauthorized)
		if len(authorizationHeader) == 0 {
			err := errors.New("authorization header is not provided")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}
		// Fields() function will split auth header by space
		fields := strings.Fields(authorizationHeader)
		// očekáváme, že z toho vzniknout vždy min dva elementy
		// jelikož např. pro JWT by obsah pole authorization v headru měl vypadat jako bearer $Token
		if len(fields) < 2 {
			err := errors.New("invalid authorization header format")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		// pro snadnější porovnání převedena první položka z pole do lower case
		authorizationType := strings.ToLower(fields[0])
		if authorizationType != authorizationTypeBearer {
			err := fmt.Errorf("unsupported authorization type %s", authorizationType)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		// druhá položka v poli by měla být token k verifikaci
		accessToken := fields[1]

		payload, err := tokenMaker.VerifyToken(accessToken)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		// uložení payloadu z verifikovaného tokenu do contextu jako key value pair
		ctx.Set(authorizationPayloadKey, payload)

		// posunutí requestu k dalšímu handleru (k dalšímu middlewaru v chainu middlewarů nebo ke konečné funkci v endpointu)
		ctx.Next()
	}
}
