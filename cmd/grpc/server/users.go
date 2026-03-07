package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	userpb "github.com/chris-a-kaiser-7/go-rest-template/cmd/grpc/proto/user"

	"github.com/chris-a-kaiser-7/go-rest-template/internal/data"
	"github.com/chris-a-kaiser-7/go-rest-template/internal/validator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *application) ActivateUser(_ context.Context, req *userpb.GetUserRequest) (*userpb.AddUserResponse, error) {
	return nil, nil
}

func (s *application) GetUser(_ context.Context, req *userpb.GetUserRequest) (*userpb.GetUserResponse, error) {
	return nil, nil
}

func (s *application) RegisterUser(_ context.Context, req *userpb.AddUserRequest) (*userpb.AddUserResponse, error) {
	s.logger.PrintInfo("register user start", nil)
	// Copy the data from the request body into a new User struct. Notice also that
	// we set the Activated field to false, which isn't strictly necessary because
	// the Activated field will have the zero-value of false by default. But setting
	// this explicitly helps to make our intentions clear to anyone reading the code.
	user := &data.User{
		Name:      req.Name,
		Email:     req.Email,
		Activated: true, //Set to true for development
	}

	// Use the Password.Set() method to generate and store the hashed and plaintext
	// passwords.
	err := user.Password.Set(req.Password)
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "%v", err)
	}

	v := validator.New()

	// Validate the user struct and return the error messages to the client if
	// any of the checks fail.
	if data.ValidateUser(v, user); !v.Valid() {
		return nil, status.Errorf(codes.InvalidArgument, "%s", v.Errors)
	}

	// Insert the user data into the database.
	err = s.models.Users.Insert(user)
	if err != nil {
		switch {
		// If we get an ErrDuplicateEmail error, use the v.AddError() method to manually add
		// a message to the validator instance, and then call our failedValidationResponse
		// helper().
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			return nil, status.Errorf(codes.InvalidArgument, "%s", v.Errors)
		default:
			return nil, status.Errorf(codes.Unavailable, "%v", err)
		}
	}

	err = s.models.Permissions.AddForUser(user.ID, "keys:add", "keys:get", "keys:delete")
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "%v", err)
	}

	// After the user record has been created in the database, generate a new activation
	// token for the user.
	token, err := s.models.Tokens.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "%v", err)
	}

	// Launch a goroutine which runs an anonymous function that sends the welcome email using
	// the background helper function.
	s.background(func() {
		// Create map to act as a 'holding structure' for the data we send to the weclome email
		// template.
		data := map[string]interface{}{
			"activationToken": token.Plaintext,
			"userID":          user.ID,
		}

		// Call the Send() method on our Mailer, passing in the user's email address, name of the
		// template file, and the data map containing the activationToken and the user's ID.
		err = s.mailer.Send(user.Email, "user_welcome.tmpl", data)
		if err != nil {
			// Importantly, if there is an error sending the email then we log the error
			// instead of raising a server error like before when we handled
			// the email send functionality without a goroutine
			s.logger.PrintError(err, nil)
		}
	})

	// Note that we also change this to send the client a 202 Accepted status code which
	// indicates that the request has been accepted for processing, but the processing has
	// not been completed.

	//err = app.writeJSON(w, http.StatusAccepted, envelope{"user": user}, nil)
	//TODO: have a real response
	s.logger.PrintInfo("register user end", nil)
	res := userpb.AddUserResponse{Id: fmt.Sprint(user.ID)}
	return &res, nil
}
