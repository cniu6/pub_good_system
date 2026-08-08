package utils

import (
	"errors"
	"unicode"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

var (
	errPasswordTooShort    = errors.New("Password length must be at least 8 characters")
	errPasswordTooLong     = errors.New("Password length cannot exceed 72 bytes")
	errPasswordNeedLetter  = errors.New("Password must contain at least one letter")
	errPasswordNeedNumber  = errors.New("Password must contain at least one digit")
)

// ValidatePasswordStrength 校验密码强度与 bcrypt 安全边界。
func ValidatePasswordStrength(password string) error {
	if utf8.RuneCountInString(password) < 8 {
		return errPasswordTooShort
	}
	if len(password) > 72 {
		return errPasswordTooLong
	}

	hasLetter := false
	hasNumber := false
	for _, r := range password {
		if unicode.IsLetter(r) {
			hasLetter = true
		}
		if unicode.IsDigit(r) {
			hasNumber = true
		}
	}

	if !hasLetter {
		return errPasswordNeedLetter
	}
	if !hasNumber {
		return errPasswordNeedNumber
	}
	return nil
}

// HashPassword generates a bcrypt hash of the password
func HashPassword(password string) (string, error) {
	if err := ValidatePasswordStrength(password); err != nil {
		return "", err
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func IsPasswordValidationError(err error) bool {
	return errors.Is(err, errPasswordTooShort) ||
		errors.Is(err, errPasswordTooLong) ||
		errors.Is(err, errPasswordNeedLetter) ||
		errors.Is(err, errPasswordNeedNumber)
}

// CheckPasswordHash compares a bcrypt hashed password with its possible plaintext equivalent
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

