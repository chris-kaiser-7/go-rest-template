package data

import (
	"testing"
)

func TestUserModel_InsertValid(t *testing.T) {
	userPass := password{}
	err := userPass.Set("securePassword123")
	if err != nil {
		t.Fatalf("failed to create password: %v", err)
	}
	testUser := &User{
		Name:      "TestUser2",
		Email:     "validEmail2@gmail.com",
		Password:  userPass,
		Activated: true,
	}
	err = daWrappers.Users.Insert(testUser)
	if err != nil {
		t.Fatalf("failed to insert testUser: %v", err)
	}
	if testUser.ID <= 1 {
		t.Fatalf("expected ID > 1, got %d", testUser.ID)
	}
	if testUser.Version != 1 {
		t.Fatalf("expected Version == 1, got %d", testUser.Version)
	}
}

func TestUserModel_InsertInvalidDuplicateEmail(t *testing.T) {
	userPass := password{}
	err := userPass.Set("securePassword123")
	if err != nil {
		t.Fatalf("failed to create password: %v", err)
	}
	testUser := &User{
		Name:      "DuplicateUser",
		Email:     "validEmail@gmail.com",
		Password:  userPass,
		Activated: true,
	}
	err = daWrappers.Users.Insert(testUser)
	if err != ErrDuplicateEmail {
		t.Fatalf("expected error ErrDuplicateEmail, instead got: %v", err)
	}
}

func TestUserModel_GetByEmailValid(t *testing.T) {
	user, err := daWrappers.Users.GetByEmail("validEmail@gmail.com")
	if err != nil {
		t.Fatalf("failed to get user with validEmail@gmail.com: %v", err)
	}
	if user.Name != "TestUser1" {
		t.Fatalf("expected name == TestUser1, got %s", user.Name)
	}
}

func TestUserModel_GetByEmailInvalid(t *testing.T) {
	_, err := daWrappers.Users.GetByEmail("invalidEmail@gmail.com")
	if err != ErrRecordNotFound {
		t.Fatalf("expected error ErrRecordNotFound, instead got: %v", err)
	}
}

//TODO: test getfortoken
