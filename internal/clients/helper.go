package clients

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/MitP1997/golang-user-management/internal/constants"
	"github.com/MitP1997/golang-user-management/internal/datatypes"
	"github.com/google/uuid"
)

func generateToken(tokenType datatypes.TokenType) (token string) {
	if tokenType == constants.TokenTypeUuid {
		uuid := uuid.New().String()
		token = strings.Replace(uuid, "-", "", -1)
		return
	}
	if tokenType == constants.TokenType6DigitOtp {
		// generate 6 digit otp
		otp := rand.Intn(1000000)
		token = fmt.Sprintf("%06d", otp)
		return
	}
	return
}
