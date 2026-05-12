package users

import (
	"context"
	"testing"
)

func TestUserService_GetUser(t *testing.T) {
	svc := NewUserService(&mockUserQueries{})
	user, err := svc.GetUser(context.Background(), "550e8400-e29b-41d4-a716-446655440000")
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}
	if user.ID == "" {
		t.Error("expected non-empty user ID")
	}
}

func TestUserService_GetUser_InvalidID(t *testing.T) {
	svc := NewUserService(&mockUserQueries{})
	_, err := svc.GetUser(context.Background(), "not-a-uuid")
	if err == nil {
		t.Error("expected error for invalid UUID")
	}
}

func TestUserService_UpdateProfile(t *testing.T) {
	svc := NewUserService(&mockUserQueries{})
	name := "New Name"
	err := svc.UpdateProfile(context.Background(), "550e8400-e29b-41d4-a716-446655440000", UpdateProfileRequest{
		DisplayName: &name,
	})
	if err != nil {
		t.Fatalf("UpdateProfile failed: %v", err)
	}
}
