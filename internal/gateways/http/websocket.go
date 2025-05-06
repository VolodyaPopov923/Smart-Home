package http

import (
	"context"
	"encoding/json"
	"errors"
	"homework/internal/usecase"
	"log"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/coder/websocket"
)

type WebSocketHandler struct {
	useCases UseCases
	mu       sync.Mutex
	conns    map[*websocket.Conn]struct{}
}

func NewWebSocketHandler(useCases UseCases) *WebSocketHandler {
	return &WebSocketHandler{
		useCases: useCases,
		conns:    make(map[*websocket.Conn]struct{}),
	}
}

func (h *WebSocketHandler) Handle(c *gin.Context, id int64) error {
	ctx := c.Request.Context()

	_, err := h.useCases.Sensor.GetSensorByID(ctx, id)
	if err != nil {
		if errors.Is(err, usecase.ErrSensorNotFound) {
			c.Status(404)
		} else {
			c.Status(500)
		}
		return err
	}

	conn, err := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return err
	}

	connCtx, cancelConn := context.WithCancel(ctx)
	defer cancelConn()

	h.mu.Lock()
	h.conns[conn] = struct{}{}
	h.mu.Unlock()

	connClosed := make(chan struct{})

	go func() {
		defer func() {
			h.mu.Lock()
			delete(h.conns, conn)
			h.mu.Unlock()
			cancelConn()
			if err := conn.Close(websocket.StatusNormalClosure, "connection closed"); err != nil {
				log.Printf("Error closing connection: %v", err)
			}

			close(connClosed)
		}()

		for {
			_, _, err := conn.Read(connCtx)
			if err != nil {
				if !errors.Is(err, context.Canceled) {
					log.Printf("WebSocket read error: %v", err)
				}
				break
			}
		}
	}()

	go func() {
		select {
		case <-connClosed:
			return
		case <-connCtx.Done():
			return
		case <-time.After(200 * time.Millisecond):
		}

		select {
		case <-connClosed:
			return
		case <-connCtx.Done():
			return
		default:
			lastEvent, err := h.useCases.Event.GetLastEventBySensorID(connCtx, id)
			if err != nil {
				log.Printf("Error getting last event: %v", err)
				return
			}

			if lastEvent == nil {
				log.Printf("No events found for sensor ID: %d", id)
				return
			}

			data, err := json.Marshal(lastEvent)
			if err != nil {
				log.Printf("Error marshaling event data: %v", err)
				return
			}

			select {
			case <-connClosed:
				return
			case <-connCtx.Done():
				return
			default:
				err := conn.Write(connCtx, websocket.MessageText, data)
				if err != nil {
					log.Printf("WebSocket write error: %v", err)
				}
			}
		}
	}()

	select {
	case <-connClosed:
	case <-ctx.Done():
	}

	return nil
}

func (h *WebSocketHandler) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	h.mu.Lock()
	conns := make([]*websocket.Conn, 0, len(h.conns))
	for conn := range h.conns {
		conns = append(conns, conn)
	}
	h.conns = make(map[*websocket.Conn]struct{})
	h.mu.Unlock()

	var wg sync.WaitGroup
	for _, conn := range conns {
		wg.Add(1)
		go func(c *websocket.Conn) {
			defer wg.Done()
			if err := c.CloseRead(ctx); err != nil {
				log.Printf("Error closing read: %v", err)
			}

			if err := c.Close(websocket.StatusNormalClosure, "server shutting down"); err != nil {
				log.Printf("Error closing connection: %v", err)
			}
		}(conn)
	}
	wg.Wait()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}
