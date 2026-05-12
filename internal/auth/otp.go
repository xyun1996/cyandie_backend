package auth

import (
	"context"
	"log/slog"
)

// OTPNotifier sends OTP codes for login/verification.
type OTPNotifier interface {
	SendOTP(ctx context.Context, target string, code string) error
}

// LogNotifier logs OTP codes instead of sending them. For development only.
type LogNotifier struct{}

func (LogNotifier) SendOTP(_ context.Context, target string, code string) error {
	slog.Info("OTP code", "target", target, "code", code)
	return nil
}
