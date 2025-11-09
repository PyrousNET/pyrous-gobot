package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/mattermost/mattermost/server/public/model"

	"github.com/pyrousnet/pyrous-gobot/internal/handler"
	"github.com/pyrousnet/pyrous-gobot/internal/mmclient"
)

type wsSessionState struct {
	connectionID string
	lastSequence int64
}

func superviseWebSocket(ctx context.Context, mmClient *mmclient.MMClient, handler *handler.Handler, quit chan bool) {
	state := &wsSessionState{}
	var attempt int

	for {
		if ctx.Err() != nil {
			return
		}

		ws, resumed, err := establishWebSocket(mmClient, state)
		if err != nil {
			delay := backoffDelay(attempt)
			log.Printf("failed to connect to websocket: %v; retrying in %s", err, delay)
			attempt++
			if !sleepWithContext(ctx, delay) {
				return
			}
			continue
		}

		attempt = 0
		if resumed {
			log.Printf("resumed websocket session at seq=%d", state.lastSequence)
		} else {
			log.Print("connected new websocket session")
		}

		sessionCtx, cancel := context.WithCancel(ctx)
		err = runWebSocketSession(sessionCtx, ws, handler, quit, state)
		cancel()

		if errors.Is(err, context.Canceled) || ctx.Err() != nil {
			return
		}

		if err != nil {
			log.Printf("websocket session ended: %v", err)
		} else {
			log.Print("websocket session ended")
		}

		delay := backoffDelay(attempt)
		attempt++
		log.Printf("reconnecting websocket in %s", delay)
		if !sleepWithContext(ctx, delay) {
			return
		}
	}
}

func establishWebSocket(mmClient *mmclient.MMClient, state *wsSessionState) (*model.WebSocketClient, bool, error) {
	if state.connectionID != "" && state.lastSequence > 0 {
		ws, err := mmClient.NewReliableWebSocketClient(state.connectionID, state.lastSequence)
		if err == nil {
			return ws, true, nil
		}
		log.Printf("failed to resume websocket connection_id=%s: %v; falling back to clean session", state.connectionID, err)
	}

	ws, err := mmClient.NewWebSocketClient()
	if err != nil {
		return nil, false, err
	}
	state.connectionID = ""
	state.lastSequence = 0

	return ws, false, nil
}

func runWebSocketSession(ctx context.Context, ws *model.WebSocketClient, handler *handler.Handler, quit chan bool, state *wsSessionState) error {
	ws.Listen()
	defer ws.Close()

	go drainWebSocketResponses(ctx, ws)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ws.PingTimeoutChannel:
			return errors.New("websocket heartbeat timeout")
		case event, ok := <-ws.EventChannel:
			if !ok {
				if ws.ListenError != nil {
					return fmt.Errorf("websocket listener error: %w", ws.ListenError)
				}
				return errors.New("websocket channel closed")
			}
			if event == nil {
				continue
			}

			if seq := event.GetSequence(); seq > 0 {
				state.lastSequence = seq
			}

			if event.EventType() == model.WebsocketEventHello {
				if connID, ok := event.GetData()["connection_id"].(string); ok && connID != "" {
					state.connectionID = connID
				}
			}

			go handler.HandleWebSocketResponse(quit, event)
		}
	}
}

func drainWebSocketResponses(ctx context.Context, ws *model.WebSocketClient) {
	for {
		select {
		case <-ctx.Done():
			return
		case resp, ok := <-ws.ResponseChannel:
			if !ok {
				return
			}
			if resp != nil && resp.Status != model.StatusOk {
				log.Printf("websocket response status=%s error=%s", resp.Status, resp.Error)
			}
		}
	}
}

func backoffDelay(attempt int) time.Duration {
	const (
		base = time.Second
		max  = 30 * time.Second
	)

	power := math.Pow(2, float64(attempt))
	delay := time.Duration(power) * base
	if delay > max {
		delay = max
	}

	jitter := time.Duration(rand.Int63n(int64(base)))
	return delay + jitter
}

func sleepWithContext(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
