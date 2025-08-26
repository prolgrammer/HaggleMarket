package apperror

import (
	"fmt"
)

const (
	NotFound           = "NOT_FOUND"
	DatabaseError      = "DATABASE_ERROR"
	InternalError      = "INTERNAL_ERROR"
	DuplicateError     = "DUPLICATE_ERROR"
	BadRequest         = "BAD_REQUEST"
	Unauthorized       = "UNAUTHORIZED"
	InvalidToken       = "INVALID_TOKEN"
	InvalidFingerprint = "INVALID_FINGERPRINT"
	Conflict           = "CONFLICT"
	Forbidden          = "FORBIDDEN"
)

type AppError interface {
	error
	Code() string
	Message() string
	Unwrap() error
}

type Error struct {
	ErrCode    string
	ErrMsg     string
	WrappedErr error
}

func (e *Error) Error() string {
	if e.WrappedErr != nil {
		return fmt.Sprintf("%s: %v", e.ErrMsg, e.WrappedErr)
	}
	return e.ErrMsg
}
func (e *Error) Code() string    { return e.ErrCode }
func (e *Error) Message() string { return e.ErrMsg }
func (e *Error) Unwrap() error   { return e.WrappedErr }

func New(code, message string, cause error) *Error {
	return &Error{ErrCode: code, ErrMsg: message, WrappedErr: cause}
}

var (
	ErrProductNotFound = New(NotFound, "product not found", nil)
	ErrStoreNotFound   = New(NotFound, "store not found", nil)

	ErrOfferNotFound = New(NotFound, "offer not found", nil)

	ErrUserNotFound             = New(NotFound, "user not found", nil)
	ErrIncorrectPassword        = New(Unauthorized, "incorrect password", nil)
	ErrFailedToGeneratePassword = New(InternalError, "failed to generate password", nil)
	ErrInvalidFingerprint       = New(InvalidFingerprint, "fingerprints don't match", nil)

	ErrInvalidToken  = New(InvalidToken, "invalid token", nil)
	ErrTokenNotFound = New(NotFound, "token not found", nil)

	ErrNotificationNotFound = New(NotFound, "notification not found", nil)
)

// ReviewError представляет ошибку, связанную с отзывами
type ReviewError struct {
	Code    string
	Message string
}

// Error реализует интерфейс error
func (e *ReviewError) Error() string {
	return e.Message
}

// Константы для кодов ошибок отзывов
const (
	ReviewNotFound      = "review_not_found"
	ReviewDuplicate     = "review_duplicate"
	ReviewDatabaseError = "review_database_error"
	ReviewUnauthorized  = "review_unauthorized"
)

// NewReviewError создает новую ошибку отзыва
func NewReviewError(code string, message string) *ReviewError {
	return &ReviewError{
		Code:    code,
		Message: message,
	}
}

// GuestOfferError represents an error related to guest offers
type GuestOfferError struct {
	Code    string
	Message string
}

// Error implements the error interface
func (e *GuestOfferError) Error() string {
	return e.Message
}

// Constants for guest offer error codes
const (
	GuestOfferInvalidData   = "guest_offer_invalid_data"
	GuestOfferProcessFailed = "guest_offer_process_failed"
	GuestOfferStoreNotFound = "guest_offer_store_not_found"
	GuestOfferDatabaseError = "guest_offer_database_error"
)

// NewGuestOfferError creates a new guest offer error
func NewGuestOfferError(code string, message string) *GuestOfferError {
	return &GuestOfferError{
		Code:    code,
		Message: message,
	}
}
