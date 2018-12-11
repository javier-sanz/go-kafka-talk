// Example channel-based high-level Apache Kafka consumer
package main

/**
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	_ "github.com/kshvakov/clickhouse"
)

func createMessagesTable(dbConn *sql.DB) {
	_, err := dbConn.Exec(`CREATE TABLE IF NOT EXISTS default.Messages
		(
		   MessageTime DateTime,
		   MessageDate Date,
		   MessageCount UInt64
		)
		ENGINE = MergeTree(MessageDate, (MessageTime, MessageDate), 8192)`)
	if err != nil {
		log.Printf("table creation-error: %s", err.Error())
		panic(err.Error())
	}
}

func getStatement(transaction *sql.Tx) *sql.Stmt {

	statement, err := transaction.Prepare(`INSERT INTO default.Messages (MessageTime, MessageDate, MessageCount) 
		VALUES (?, ?, ?)`)

	if err != nil {
		log.Printf("error insert sql%s", err.Error())
		panic(err.Error())
	}
	return statement
}

func process(statement *sql.Stmt, key, value string) error {
	timeValue, _ := time.Parse(time.RFC3339, key)
	count, _ := strconv.ParseUint(value, 10, 64)
	_, err := statement.Exec(
		timeValue,
		timeValue,
		count,
	)

	if err != nil {
		log.Printf("insert exec error: %s", err.Error())
		return err
	}

	return nil
}

func main() {

	if len(os.Args) < 5 {
		fmt.Fprintf(os.Stderr, "Usage: %s <broker> <group> <topic> <db>\n",
			os.Args[0])
		os.Exit(1)
	}

	broker := os.Args[1]
	group := os.Args[2]
	topics := os.Args[3]
	dbLink := os.Args[4]

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":               broker,
		"group.id":                        group,
		"session.timeout.ms":              6000,
		"go.events.channel.enable":        true,
		"go.application.rebalance.enable": true,
		"enable.auto.commit":              false,
		"default.topic.config": kafka.ConfigMap{
			"auto.offset.reset":   "earliest",
			"offset.store.method": "broker",
		},
	})

	if err != nil {
		log.Printf("Failed to create consumer: %s\n", err)
		os.Exit(1)
	}

	log.Printf("Created Consumer %v\n", c)

	err = c.SubscribeTopics([]string{topics}, nil)
	if err != nil {
		log.Fatalf("Failed to create consumer: %s\n", err)
	}

	dbConn, err := sql.Open("clickhouse", dbLink)
	if err != nil {
		log.Fatalf("DB connection error %s", err.Error())
	}

	signalsCh := make(chan os.Signal, 1)
	// we catch SIGUSR1 and stop all tickers when received
	signal.Notify(signalsCh, syscall.SIGTERM, syscall.SIGINT)

	ctx := context.Background()
	//Graceful shutdown of our services
	ctx, cancelFunc := context.WithCancel(ctx)

	go func() {
		<-signalsCh
		log.Printf("cancel signal received")
		//now we propagate the signal
		cancelFunc()
	}()

	createMessagesTable(dbConn)
	commitTicker := time.NewTicker(5 * time.Second)
	defer commitTicker.Stop()
	transaction, err := dbConn.Begin()
	if err != nil {
		log.Printf("DB transaction error %s", err.Error())
		panic(err.Error())
	}
	stmt := getStatement(transaction)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Caught signal: terminating\n")
			c.Close()
			return
		case ev := <-c.Events():
			switch e := ev.(type) {
			case kafka.AssignedPartitions:
				log.Printf("%% %v\n", e)
				c.Assign(e.Partitions)
			case kafka.RevokedPartitions:
				log.Printf("%% %v\n", e)
				c.Unassign()
			case *kafka.Message:
				log.Printf("%% Message received: (%s, %s)",
					string(e.Key), string(e.Value))
				process(stmt, string(e.Key), string(e.Value))
			case kafka.PartitionEOF:
				log.Printf("%% Reached %v\n", e)
			case kafka.Error:
				log.Printf("%% Error: %v\n", e)
			}
		case <-commitTicker.C:
			err := transaction.Commit()
			if err != nil {
				log.Printf("DB commit error %s, reseting window", err.Error())
				break
			}
			transaction, err = dbConn.Begin()
			if err != nil {
				log.Printf("DB transaction error %s, reseting window", err.Error())
				panic(err.Error())
			}
			stmt = getStatement(transaction)
			// Commiting offsets
			c.Commit()
		}
	}

}
