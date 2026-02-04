/*
note:
- behaviour is expected as:
  - https://github.com/prothegee/drogon-examples/blob/main/drogon-kafka/tools/producer_ctl/main.cc
*/
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	// "math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"showcase-backend-go/pkg"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func main() {
	// r := rand.New(rand.NewSource(time.Now().UnixNano()))

	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": "127.0.0.1:9092",
	})
	if err != nil {
		log.Fatalf("fail to create kafka producer: %v", err.Error())
	}

	defer p.Close()

	topic := pkg.GOKAFKA_STOCK_TRADE_TOPIC
	trade := pkg.StockTrade{}
	trades := []*pkg.StockTrade_tj{
		trade.StockTradeNew(300_000.00, "USD", "BIZ1"),
		trade.StockTradeNew(400_000.00, "USD", "BIZ2"),
		trade.StockTradeNew(500_000.00, "USD", "BIZ3"),
	}

	ctx, cancel := context.WithCancel(context.Background())

	sigCh := make(chan os.Signal, 1)

	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		sc := <-sigCh
		fmt.Printf("\nSIG (%d) shutingdown gracefully at %s\n", sc, pkg.TimestampNow())
		cancel()
	}()

	// main loop
	for {
		select {
		case <-ctx.Done():
			{
				log.Print("producer stop\n")
				return
			}
		default:
			{
				for _, t := range trades {
					t.Update()
				}

				data, _ := json.Marshal(trades)
				payload := string(data)

				err = p.Produce(&kafka.Message{
					TopicPartition: kafka.TopicPartition{
						Topic:     &topic,
						Partition: kafka.PartitionAny,
					},
					Value: []byte(payload),
				}, nil)
				if err != nil {
					log.Printf("producer error: %v\n", err.Error())
				}

				// ensure sent
				p.Flush(pkg.GOKAFKA_DELAY_MS)
				time.Sleep(time.Duration(pkg.GOKAFKA_DELAY_MS) * time.Millisecond)
			}
		}
	}
}
