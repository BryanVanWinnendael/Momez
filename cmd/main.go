package main

import (
	"context"
	"momez/db"
	"momez/handlers"
	"momez/services"
	"os"

	"cloud.google.com/go/storage"
	firebase "firebase.google.com/go/v4"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"google.golang.org/api/option"
)

func main() {
	godotenv.Load()
	e := echo.New()

	var (
		SECRET_KEY string = os.Getenv("SECRET_KEY")
		DB_NAME    string = "app_data.db"
	)

	e.Static("/", "assets")
	e.Static("/css", "css")
	e.Static("/static", "static")

	e.HTTPErrorHandler = handlers.CustomHTTPErrorHandler

	// Firebase Storage
	opt := option.WithCredentialsFile("serviceAccountKey.json")
	ctx := context.Background()

	client, err := storage.NewClient(ctx, opt)
	if err != nil {
		e.Logger.Fatalf("failed to create storage client: %s", err)
	}
	defer client.Close()

	// Firebase Database
	conf := &firebase.Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}

	app, err := firebase.NewApp(ctx, conf, opt)
	if err != nil {
		e.Logger.Fatalf("failed to create firebase app: %s", err)
	}

	database, err := app.Database(ctx)
	if err != nil {
		e.Logger.Fatalf("failed to create database client: %s", err)
	}

	firebase_db := db.NewDB(client, database)

	// Session Middleware
	e.Use(session.Middleware(sessions.NewCookieStore([]byte(SECRET_KEY))))

	store, err := db.NewStore(DB_NAME)
	if err != nil {
		e.Logger.Fatalf("failed to create store: %s", err)
	}

	us := services.NewUserServices(services.User{}, store, firebase_db)
	ah := handlers.NewAuthHandler(us)

	ps := services.NewPostServices(firebase_db)
	ph := handlers.NewPostHandler(ps)

	// Setting Routes
	handlers.SetupRoutes(e, ah, ph)

	// Start Server
	e.Logger.Fatal(e.Start(":3000"))
}
