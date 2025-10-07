package firebase

import (
	"context"
	"fmt"
	"os"
	"sync"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

var (
	// Singleton Firebase app instance
	firebaseApp  *firebase.App
	firebaseAuth *auth.Client
	once         sync.Once
	initError    error
)

// InitializeFirebase initializes the Firebase app singleton
func InitializeFirebase(ctx context.Context) error {
	once.Do(func() {
		projectID := os.Getenv("FIREBASE_PROJECT_ID")
		if projectID == "" {
			initError = fmt.Errorf("FIREBASE_PROJECT_ID environment variable not set")
			return
		}

		credsPath := os.Getenv("FIREBASE_CREDENTIALS_PATH")

		var opts []option.ClientOption
		if credsPath != "" {
			// Use service account credentials file
			opts = append(opts, option.WithCredentialsFile(credsPath))
		}
		// If no credentials path, Firebase SDK will use Application Default Credentials
		// This works in Cloud Run and other GCP environments

		config := &firebase.Config{
			ProjectID: projectID,
		}

		app, err := firebase.NewApp(ctx, config, opts...)
		if err != nil {
			initError = fmt.Errorf("failed to initialize Firebase app: %w", err)
			return
		}

		authClient, err := app.Auth(ctx)
		if err != nil {
			initError = fmt.Errorf("failed to initialize Firebase Auth client: %w", err)
			return
		}

		firebaseApp = app
		firebaseAuth = authClient
	})

	return initError
}

// GetAuth returns the Firebase Auth client singleton
// Must call InitializeFirebase first
func GetAuth() (*auth.Client, error) {
	if firebaseAuth == nil {
		return nil, fmt.Errorf("Firebase not initialized. Call InitializeFirebase first")
	}
	return firebaseAuth, nil
}

// VerifyIDToken verifies a Firebase ID token and returns the decoded token
func VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error) {
	client, err := GetAuth()
	if err != nil {
		return nil, err
	}

	token, err := client.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}

	// Only log in development/debug mode, without exposing sensitive claims
	if os.Getenv("ENVIRONMENT") == "development" || os.Getenv("LOG_LEVEL") == "debug" {
		fmt.Printf("DEBUG: Token verified successfully\n")
	}

	return token, nil
}

// IsInitialized returns true if Firebase has been initialized
func IsInitialized() bool {
	return firebaseAuth != nil
}
