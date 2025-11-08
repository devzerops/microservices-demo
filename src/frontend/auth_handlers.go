// Copyright 2018 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"net/http"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	pb "github.com/GoogleCloudPlatform/microservices-demo/src/frontend/genproto"
)

// loginHandler handles both GET (show login form) and POST (process login)
func (fe *frontendServer) loginHandler(w http.ResponseWriter, r *http.Request) {
	log := r.Context().Value(ctxKeyLog{}).(logrus.FieldLogger)

	if r.Method == http.MethodGet {
		// Show login form
		if err := templates.ExecuteTemplate(w, "login", map[string]interface{}{
			"baseUrl":          baseUrl,
			"frontendMessage":  frontendMessage,
			"is_cymbal_brand":  isCymbalBrand,
		}); err != nil {
			log.WithError(err).Error("failed to render login template")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Process login (POST)
	email := r.FormValue("email")
	password := r.FormValue("password")

	if email == "" || password == "" {
		renderLoginError(w, "Email and password are required")
		return
	}

	// Call auth service to login
	authClient := pb.NewAuthServiceClient(fe.authSvcConn)
	resp, err := authClient.Login(r.Context(), &pb.LoginRequest{
		Email:    email,
		Password: password,
	})

	if err != nil {
		log.WithError(err).Warn("login failed")
		renderLoginError(w, "Invalid email or password")
		return
	}

	// Set auth token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     cookieAuthToken,
		Value:    resp.Token,
		MaxAge:   cookieMaxAge,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	log.WithField("user_id", resp.User.UserId).Info("user logged in successfully")

	// Redirect to home page
	http.Redirect(w, r, baseUrl+"/", http.StatusSeeOther)
}

// signupHandler handles both GET (show signup form) and POST (process signup)
func (fe *frontendServer) signupHandler(w http.ResponseWriter, r *http.Request) {
	log := r.Context().Value(ctxKeyLog{}).(logrus.FieldLogger)

	if r.Method == http.MethodGet {
		// Show signup form
		if err := templates.ExecuteTemplate(w, "signup", map[string]interface{}{
			"baseUrl":          baseUrl,
			"frontendMessage":  frontendMessage,
			"is_cymbal_brand":  isCymbalBrand,
		}); err != nil {
			log.WithError(err).Error("failed to render signup template")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Process signup (POST)
	name := r.FormValue("name")
	email := r.FormValue("email")
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")

	// Validate input
	if name == "" || email == "" || password == "" {
		renderSignupError(w, "All fields are required")
		return
	}

	if password != confirmPassword {
		renderSignupError(w, "Passwords do not match")
		return
	}

	if len(password) < 6 {
		renderSignupError(w, "Password must be at least 6 characters")
		return
	}

	// Call auth service to register
	authClient := pb.NewAuthServiceClient(fe.authSvcConn)
	resp, err := authClient.Register(r.Context(), &pb.RegisterRequest{
		Email:    email,
		Password: password,
		Name:     name,
	})

	if err != nil {
		log.WithError(err).Warn("registration failed")
		renderSignupError(w, "Registration failed. Email may already be in use.")
		return
	}

	// Set auth token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     cookieAuthToken,
		Value:    resp.Token,
		MaxAge:   cookieMaxAge,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	log.WithField("user_id", resp.User.UserId).Info("user registered successfully")

	// Redirect to home page
	http.Redirect(w, r, baseUrl+"/", http.StatusSeeOther)
}

// renderLoginError renders the login page with an error message
func renderLoginError(w http.ResponseWriter, errorMsg string) {
	templates.ExecuteTemplate(w, "login", map[string]interface{}{
		"baseUrl":          baseUrl,
		"frontendMessage":  frontendMessage,
		"is_cymbal_brand":  isCymbalBrand,
		"error":            errorMsg,
	})
}

// renderSignupError renders the signup page with an error message
func renderSignupError(w http.ResponseWriter, errorMsg string) {
	templates.ExecuteTemplate(w, "signup", map[string]interface{}{
		"baseUrl":          baseUrl,
		"frontendMessage":  frontendMessage,
		"is_cymbal_brand":  isCymbalBrand,
		"error":            errorMsg,
	})
}

// getCurrentUser retrieves the current user from the auth token
func (fe *frontendServer) getCurrentUser(r *http.Request) (*pb.User, error) {
	cookie, err := r.Cookie(cookieAuthToken)
	if err != nil {
		return nil, err
	}

	authClient := pb.NewAuthServiceClient(fe.authSvcConn)
	resp, err := authClient.ValidateToken(r.Context(), &pb.ValidateTokenRequest{
		Token: cookie.Value,
	})

	if err != nil || !resp.Valid {
		return nil, errors.New("invalid token")
	}

	return resp.User, nil
}
