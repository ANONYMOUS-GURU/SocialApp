package database

import (
	"context"
	"fmt"
	"log"
	"os"

	"database/sql"
	ws "g_chat/wsConnections"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

var (
	myDatabase *DB
	queries    *Queries
	pool       *pgxpool.Pool
)

var dsn string = "host=localhost user=postgres password=*Fu1ab%1 dbname=hog-gg port=5432 sslmode=disable"

func CreateDatabaseInstance() error {
	myDb, err := sql.Open("postgres", dsn)

	defer func() {
		if err != nil {
			myDb.Close()
		}
	}()

	if err != nil {
		log.Fatalf("error connecting to database %v", err)
		return err
	}

	if err = myDb.Ping(); err != nil {
		myDb.Close()
		log.Fatalf("error pinging database %v", err)
		return err
	}

	myDatabase = &DB{myDb}

	createQueries()

	return nil
}

func InitializeListener(connectionManager *ws.ConnectionManager) error {
	newPool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Printf("connection error : cannot connect to DB : f(initializeListener) : error -> %v", err)
		return err
	}
	pool = newPool

	go listen(connectionManager)

	return nil
}

func createQueries() {
	queries = New(getDatabase())
}

func getDatabase() *DB {
	return myDatabase
}

func getQueries() *Queries {
	return queries
}

func (db *DB) MigrateDatabase() error {
	return nil
}

func listen(connectionManager *ws.ConnectionManager) {
	// get a connection for notification
	conn, err := pool.Acquire(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error acquiring connection:", err)
		os.Exit(1)
	}
	defer conn.Release()

	// register chat_written channel
	_, err = conn.Exec(context.Background(), "listen chat_written")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error listening to chat_written channel:", err)
		os.Exit(1)
	}

	// register chat_received channel
	_, err = conn.Exec(context.Background(), "listen chat_received")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error listening to chat_received channel:", err)
		os.Exit(1)
	}

	// register chat_read channel
	_, err = conn.Exec(context.Background(), "listen chat_read")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error listening to chat_read channel:", err)
		os.Exit(1)
	}

	// handle notifications
	for {
		notification, err := conn.Conn().WaitForNotification(context.Background())
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error waiting for notification:", err)
			os.Exit(1)
		}

		_, ok := connectionManager.OutgoingHandlers[notification.Channel]

		if !ok {
			log.Printf("channel not registered")
		} else {
			err := connectionManager.OutgoingHandlers[notification.Channel](notification.Channel, notification.Payload)

			if err != nil {
				log.Printf("error doing %v operation with payload %v", notification.Channel, notification.Payload)
			}
		}
	}
}
