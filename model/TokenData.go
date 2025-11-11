package model

// TokenData model represents user info encrypted with the JWT token
// @Data
type TokenData struct {
	SubjectId   string `json:"subjectId"`   // Authenticated subject ID (can be user, or service account)
	SubjectType int    `json:"subjectType"` // Subject type enum
	Status      int    `json:"status"`      // User status enum
	ExpiresIn   int64  `json:"expiresIn"`   // Token expiration [Epoch milliseconds Timestamp]
}
