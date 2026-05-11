package auth

import (
	"testing"
)

func TestSessionStore_CreateAndValidate(t *testing.T) {
	store := NewSessionStore(newMockRedisClient())

	session, err := store.Create("user-1", "token-hash-abc", 60*60*24*7)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if session.UserID != "user-1" {
		t.Errorf("expected user-1, got %s", session.UserID)
	}
	if session.ID == "" {
		t.Error("expected non-empty session ID")
	}

	got, err := store.Get(session.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.UserID != "user-1" {
		t.Errorf("expected user-1, got %s", got.UserID)
	}
}

func TestSessionStore_Revoke(t *testing.T) {
	store := NewSessionStore(newMockRedisClient())

	session, _ := store.Create("user-1", "token-hash-abc", 60*60*24*7)
	err := store.Revoke(session.ID)
	if err != nil {
		t.Fatalf("Revoke failed: %v", err)
	}

	_, err = store.Get(session.ID)
	if err == nil {
		t.Error("expected error after revocation")
	}
}
