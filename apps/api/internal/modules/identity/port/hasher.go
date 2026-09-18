package port

type PasswordHasher interface {
	HashPassword(password string) (string, error)
	ComparePassword(hash, password string) (bool, error)
}
