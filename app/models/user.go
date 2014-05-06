package models

type User struct {
	Email           string
	HashedPassword  string
	FullName        string
	ClassBucket     []string
	APIKey          string
	Verified        bool
	VerificationKey string
}
