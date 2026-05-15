package eventbus

import (
	"log/slog"
	"sync"

	"github.com/roma-glushko/cargo/internal/cargo"
)

const (
	TopicCargoHandled     = "cargo.handled"
	TopicCargoMisdirected = "cargo.misdirected"
	TopicCargoArrived     = "cargo.arrived"
)

type HandlerFunc func(id cargo.TrackingID)

type message struct {
	topic   string
	payload cargo.TrackingID
}

type Bus struct {
	mu       sync.RWMutex
	handlers map[string][]HandlerFunc
	ch       chan message
	done     chan struct{}
	logger   *slog.Logger
}

func New(bufferSize int, logger *slog.Logger) *Bus {
	return &Bus{
		handlers: make(map[string][]HandlerFunc),
		ch:       make(chan message, bufferSize),
		done:     make(chan struct{}),
		logger:   logger,
	}
}

func (b *Bus) Subscribe(topic string, handler HandlerFunc) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[topic] = append(b.handlers[topic], handler)
}

func (b *Bus) Publish(topic string, id cargo.TrackingID) {
	b.ch <- message{topic: topic, payload: id}
}

func (b *Bus) Start() {
	go func() {
		for {
			select {
			case msg, ok := <-b.ch:
				if !ok {
					close(b.done)
					return
				}
				b.dispatch(msg)
			}
		}
	}()
}

func (b *Bus) Stop() {
	close(b.ch)
	<-b.done
}

func (b *Bus) dispatch(msg message) {
	b.mu.RLock()
	handlers := b.handlers[msg.topic]
	b.mu.RUnlock()

	for _, h := range handlers {
		func() {
			defer func() {
				if r := recover(); r != nil {
					b.logger.Error("event handler panicked",
						"topic", msg.topic,
						"trackingId", string(msg.payload),
						"panic", r,
					)
				}
			}()
			h(msg.payload)
		}()
	}
}
