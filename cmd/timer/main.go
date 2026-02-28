package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"4cus-guard/internal/config"
	database "4cus-guard/internal/db"
	Message "4cus-guard/internal/message"
	"4cus-guard/internal/pubsub"
	"encoding/json"
)

func main() {
	conf := config.LoadConfig()
	db, _ := database.InitDB(conf.DBPath)
	defer db.Close()

	ctx := context.Background()
	rdb, err := pubsub.NewRedisBroker(ctx, conf.RedisAddr, conf.RedisPass)
	if err != nil {
		log.Fatalf("Fail to init: %v", err)
	}

	// channel listen to OS
	quitChannel := make(chan os.Signal, 1)
	//Subscribe to signals: SIGINT (Ctrl+C) và SIGTERM (Shut down)
	signal.Notify(quitChannel, syscall.SIGINT, syscall.SIGTERM)

	msgChannel, error := rdb.Subscribe(ctx, "Timer")
	if error != nil {
		log.Fatal(error)
	}

	//Listen to channels
	for {
		select {
		//Command
		case payload := <-msgChannel:
			var msg Message.Message
			json.Unmarshal([]byte(payload), &msg)

			now := msg.Timestamp
			action := msg.Action
			switch action {
			case "start":
				//check re-type start
				var id int
				err := db.QueryRow(`SELECT id FROM focus_sessions WHERE status = 'active'`).Scan(&id)
				if err == nil || err == sql.ErrNoRows {
					query := `INSERT INTO focus_sessions (start_time, status) VALUES (?, ?)`
					db.Exec(query, now, "active")
				} else {
					fmt.Println("You have already started")
				}

			case "stop":
				query := `UPDATE focus_sessions SET end_time = ?, status = 'finished' WHERE status = 'active'`
				db.Exec(query, now)
			}

		//Case user ctrl C or shut down without stop cmd
		case osSignal := <-quitChannel:
			now := time.Now().Unix()

			query := `UPDATE focus_sessions SET end_time = ?, status = 'finished' WHERE status = 'active'`
			db.Exec(query, now)

			fmt.Println(osSignal.String())
			os.Exit(0)
		}

	}
}
