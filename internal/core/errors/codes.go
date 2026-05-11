package errors

// General
const (
	ErrInternal     = "INTERNAL_ERROR"
	ErrBadRequest   = "BAD_REQUEST"
	ErrUnauthorized = "UNAUTHORIZED"
	ErrForbidden    = "FORBIDDEN"
	ErrNotFound     = "NOT_FOUND"
	ErrConflict     = "CONFLICT"
	ErrRateLimited  = "RATE_LIMITED"
)

// Auth
const (
	ErrInvalidCredentials = "INVALID_CREDENTIALS"
	ErrTokenExpired       = "TOKEN_EXPIRED"
	ErrTokenInvalid       = "TOKEN_INVALID"
	ErrSessionRevoked     = "SESSION_REVOKED"
	ErrOTPInvalid         = "OTP_INVALID"
	ErrOTPExpired         = "OTP_EXPIRED"
)

// User
const (
	ErrUserExists = "USER_EXISTS"
	ErrUserBanned = "USER_BANNED"
)

// Platform
const (
	ErrPlatformNotSupported  = "PLATFORM_NOT_SUPPORTED"
	ErrPlatformAuthFailed    = "PLATFORM_AUTH_FAILED"
	ErrPlatformBindingExists = "PLATFORM_BINDING_EXISTS"
)
