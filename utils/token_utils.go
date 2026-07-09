package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"

	. "github.com/go-yaaf/yaaf-common-net/model"
)

// tokenSecret (AES key) and signingKy (JWT HMAC key) MUST be provided by the
// application via TokenUtils().WithSecrets(...). They are intentionally empty by
// default so the library fails closed instead of running on shared, source-visible
// secrets that would allow anyone to forge tokens and API keys.
var tokenSecret = []byte{}
var signingKy = []byte{}

// TokenUtilsStruct is a structure for token utilities
type TokenUtilsStruct struct {
}

var doOnceForTokenUtils sync.Once

var tokenUtilsSingleton *TokenUtilsStruct = nil

// TokenUtils is a factory method that acts as a static member
func TokenUtils() *TokenUtilsStruct {
	doOnceForTokenUtils.Do(func() {
		tokenUtilsSingleton = &TokenUtilsStruct{}
	})
	return tokenUtilsSingleton
}

// WithSecrets sets the API secret and signing key
func (t *TokenUtilsStruct) WithSecrets(apiSecret, signingKey string) *TokenUtilsStruct {
	if len(apiSecret) < 32 {
		panic(errors.New("api secret too short"))
	}
	tokenSecret = []byte(apiSecret[:32])

	if len(signingKey) < 32 {
		panic(errors.New("signing key too short"))
	}
	signingKy = []byte(signingKey[:32])
	return t
}

// region Access Token parsing helpers ---------------------------------------------------------------------------------

// TokenClaims represents the claims in a JWT token
type TokenClaims struct {
	jwt.RegisteredClaims
	TokenData
}

// CreateToken build JWT token from Token Data structure
func (t *TokenUtilsStruct) CreateToken(td *TokenData) (string, error) {
	t.ensureKeys()

	claims := TokenClaims{}
	claims.AccountId = td.AccountId
	claims.SubjectId = td.SubjectId
	claims.SubjectType = td.SubjectType
	claims.SubjectRole = td.SubjectRole
	claims.Status = td.Status
	claims.ExpiresIn = td.ExpiresIn
	claims.Subject = td.SubjectId

	// Bind the registered "exp" claim to the caller-supplied expiration so the JWT
	// library actually rejects expired tokens. ExpiresIn is an absolute epoch-millis
	// timestamp; when it is 0 the token carries no expiry (backward compatible).
	if td.ExpiresIn > 0 {
		claims.ExpiresAt = jwt.NewNumericDate(time.UnixMilli(td.ExpiresIn))
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(signingKy)
}

// ParseToken rebuild Token Data structure from JWT token
func (t *TokenUtilsStruct) ParseToken(tokenString string) (*TokenData, error) {
	t.ensureKeys()

	// Parse the token, pinning the accepted signing algorithm to HS256 to prevent
	// algorithm-substitution attacks (e.g. "none" or an asymmetric alg).
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return signingKy, nil
	}, jwt.WithValidMethods([]string{"HS256"}))

	if err != nil {
		return nil, err
	}

	// Validate the token and extract the claims
	if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
		return &TokenData{
			AccountId:   claims.AccountId,
			SubjectId:   claims.SubjectId,
			SubjectType: claims.SubjectType,
			SubjectRole: claims.SubjectRole,
			Status:      claims.Status,
			ExpiresIn:   claims.ExpiresIn,
		}, nil
	} else {
		return nil, fmt.Errorf("invalid token")
	}
}

// endregion

// region API Key parsing helpers --------------------------------------------------------------------------------------

// CreateApiKey generate API Key from application name
func (t *TokenUtilsStruct) CreateApiKey(appName string) (string, error) {
	t.ensureKeys()
	return encrypt(appName)
}

// ParseApiKey extract application name from API key
func (t *TokenUtilsStruct) ParseApiKey(apiKey string) (string, error) {
	t.ensureKeys()
	return decrypt(apiKey)
}

func (t *TokenUtilsStruct) ensureKeys() {
	if len(tokenSecret) < 32 {
		panic(errors.New("encryption secret is not set, please use WithSecrets to set it"))
	}
	if len(signingKy) < 32 {
		panic(errors.New("encryption signing key is not set, please use WithSecrets to set it"))
	}
}

// endregion

// region PRIVATE SECTION ----------------------------------------------------------------------------------------------

// encrypt string using authenticated AES-GCM and return hex.
// The random nonce is prepended to the ciphertext. GCM provides both
// confidentiality and integrity, so tampered/forged ciphertext is rejected on
// decrypt (unlike the previous unauthenticated CTR mode).
func encrypt(value string) (string, error) {

	block, err := aes.NewCipher(tokenSecret)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, er := io.ReadFull(rand.Reader, nonce); er != nil {
		return "", er
	}

	cipherText := gcm.Seal(nonce, nonce, []byte(value), nil)
	return hex.EncodeToString(cipherText), nil
}

// decrypt hex string using authenticated AES-GCM
func decrypt(value string) (string, error) {
	cipherTextBytes, err := hex.DecodeString(value)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(tokenSecret)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(cipherTextBytes) < nonceSize {
		return "", fmt.Errorf("cipher text too short")
	}

	nonce, cipherText := cipherTextBytes[:nonceSize], cipherTextBytes[nonceSize:]
	plain, err := gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return "", err
	}

	return string(plain), nil
}

// endregion
