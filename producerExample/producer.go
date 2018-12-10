// Example channel-based Apache Kafka producer
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
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func main() {

	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <broker> <topic>\n",
			os.Args[0])
		os.Exit(1)
	}

	broker := os.Args[1]
	topic := os.Args[2]

	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": broker})

	if err != nil {
		fmt.Printf("Failed to create producer: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created Producer %v\n", p)

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

	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				m := ev
				if m.TopicPartition.Error != nil {
					log.Printf("Delivery failed: %v\n", m.TopicPartition.Error)
				} else {
					log.Printf("Delivered message to topic %s [%d] at offset %v\n",
						*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)
				}

			default:
				log.Printf("Ignored event: %s\n", ev)
			}
		}
	}()

	go func() {
		for i := uint64(0); i < 10000000; i++ {
			select {
			case <-time.After(time.Second):
				log.Printf("Sending message counter %d ", i)
				p.ProduceChannel() <- &kafka.Message{
					TopicPartition: kafka.TopicPartition{
						Topic:     &topic,
						Partition: kafka.PartitionAny},
					Key:   []byte(time.Now().Format(time.RFC3339)),
					Value: []byte(strconv.FormatUint(i, 10)),
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// wait for delivery report goroutine to finish
	_ = <-ctx.Done()

	log.Println("Flushing last messages")
	p.Flush(500)
	if p.Len() > 0 {
		log.Printf("Some messages not delivered %d", p.Len())
	}

	p.Close()
}
