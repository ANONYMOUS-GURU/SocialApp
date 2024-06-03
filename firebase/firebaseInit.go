package firebase

import (
	"context"
	"log"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"google.golang.org/api/option"
)

const firebaseConfigFile = "./serviceAccount.json"

var (
	app        *firebase.App
	authClient *auth.Client
	ctx        = context.Background()
)

func CreateFirebaseApp() error {
	if app == nil {
		var err error
		opt := option.WithCredentialsFile(firebaseConfigFile)
		app, err = firebase.NewApp(ctx, nil, opt)
		if err != nil {
			log.Fatalf("Error getting Firebase app %v", err)
			return err
		}

		if authClient == nil {
			err = createAuthClient()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func createAuthClient() error {
	var err error
	authClient, err = app.Auth(ctx)
	if err != nil {
		log.Fatalf("Error getting Firebase auth client %v", err)
		return err
	}

	return nil
}

func GetFirebaseApp() *firebase.App {
	return app
}

func GetFirebaseAuthClient() *auth.Client {
	return authClient
}
