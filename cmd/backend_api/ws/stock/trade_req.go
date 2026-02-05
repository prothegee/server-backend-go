package backend_ws_stock

import (
	"log"
	"net/http"
	"showcase-backend-go/pkg"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	connections    = make(map[*websocket.Conn]struct{})
	connectionsMtx sync.Mutex

	consumerRunning bool
	consumerMtx     sync.Mutex
)

// --------------------------------------------------------- //

func addConnection(conn *websocket.Conn) {
	connectionsMtx.Lock()
	connections[conn] = struct{}{}

	shouldStart := !consumerRunning

	if shouldStart {
		consumerRunning = true
	}
	connectionsMtx.Unlock()

	if shouldStart {
		go startKafkaConsumer()
	}
}

func removeConnection(conn *websocket.Conn) {
	connectionsMtx.Lock()

	delete(connections, conn)

	consumerMtx.Lock()

	if len(connections) == 0 {
		consumerRunning = false
	}
	consumerMtx.Unlock()
	connectionsMtx.Unlock()
}

func getActiveConnections() []*websocket.Conn {
	connectionsMtx.Lock()
	defer connectionsMtx.Unlock()

	conns := make([]*websocket.Conn, 0, len(connections))

	for c := range connections {
		conns = append(conns, c)
	}

	return conns
}

// --------------------------------------------------------- //

func setConsumerRunning(running bool) {
	consumerMtx.Lock()
	consumerRunning = running
	consumerMtx.Unlock()
}

func isConsumerRunning() bool {
	consumerMtx.Lock()
	defer consumerMtx.Unlock()
	return consumerRunning
}

// --------------------------------------------------------- //

func startKafkaConsumer() {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "127.0.0.1:9092",
		"group.id":          "grp-consumer1",
	})
	if err != nil {
		log.Printf("fail to create consumer: %v\n", err.Error())
		setConsumerRunning(false)
		return
	}
	defer consumer.Close()

	err = consumer.Subscribe(pkg.GOKAFKA_STOCK_TRADE_TOPIC, nil)
	if err != nil {
		log.Printf("consumer failed to subscribe: %v\n", err.Error())
		setConsumerRunning(false)
		return
	}

	for isConsumerRunning() {
		msg, err := consumer.ReadMessage(
			time.Millisecond * pkg.GOKAFKA_DELAY_MS)

		if err == nil {
			payload := msg.Value

			conns := getActiveConnections()
			badConns := make([]*websocket.Conn, 0)

			for _, conn := range conns {
				err := conn.WriteMessage(
					websocket.TextMessage, payload)
				if err != nil {
					badConns = append(badConns, conn)
				}
			}

			// remove broken connection
			for _, conn := range badConns {
				removeConnection(conn)
				conn.Close()
			}
		}
		// ignore
	}

	setConsumerRunning(false)
}

// --------------------------------------------------------- //

const BackendWsStockTradeHint = "/ws/stock/trade"

func BackendWsStockTrade(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("fail to upgrade connection to websocket: %v\n", err.Error())
		return
	}
	defer conn.Close()

	addConnection(conn)
	defer removeConnection(conn)

	log.Print("connection establish\n")

	// read message until connection close
	// end-user only consume what publisher do
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}

	log.Print("connection closed\n")
}
