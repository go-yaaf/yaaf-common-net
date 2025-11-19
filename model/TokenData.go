package model

// TokenData model represents user info encrypted with the JWT token
// @Data
type TokenData struct {
	AccountId   string `json:"accountId"`   // Account ID
	SubjectId   string `json:"subjectId"`   // Authenticated subject ID (can be user, or service account)
	SubjectType int    `json:"subjectType"` // Subject type enum
	SubjectRole int    `json:"subjectRole"` // Role of user in the account (Role should be specified as Flags (Bitmask)
	Status      int    `json:"status"`      // User status enum
	ExpiresIn   int64  `json:"expiresIn"`   // Token expiration [Epoch milliseconds Timestamp]
}
