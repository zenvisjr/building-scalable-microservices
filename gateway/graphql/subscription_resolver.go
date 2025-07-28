package graphql

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/zenvisjr/building-scalable-microservices/logger"
)

type subscriptionResolver struct {
	server *Server
}

func (r *subscriptionResolver) OrderStatusChanged(ctx context.Context, orderID *string) (<-chan *OrderStatusUpdate, error) {
	Logs := logger.GetGlobalLogger()

	if orderID != nil {
		Logs.Info(ctx, "OrderStatusChanged called for specific order ID: "+*orderID)
	} else {
		Logs.Info(ctx, "OrderStatusChanged called for ALL orders")
	}

	ch := make(chan *OrderStatusUpdate, 10)

	// Subscribe to NATS messages
	sub, err := r.server.nats.Subscribe("order.status.changed", func(msg *nats.Msg) {
		msgCtx := context.Background()
		Logs.Info(msgCtx, "Received order status changed event")

		var update OrderStatusUpdate
		if err := json.Unmarshal(msg.Data, &update); err != nil {
			Logs.Error(msgCtx, "Failed to unmarshal order status update: "+err.Error())
			return
		}

		shouldSend := orderID == nil || update.OrderID == *orderID

		if shouldSend {
			Logs.Info(msgCtx, "Sending update for order ID: "+update.OrderID)
			select {
			case ch <- &update:
				Logs.Info(msgCtx, "Successfully sent update to channel for order ID: "+update.OrderID)
			case <-ctx.Done():
				return
			default:
				Logs.Error(msgCtx, "Channel full, dropping update for order ID: "+update.OrderID)
			}
		} else {
			Logs.Info(msgCtx, "Filtering out update for order ID: "+update.OrderID)
		}
	})

	if err != nil {
		Logs.Error(ctx, "Failed to subscribe to order status changed event: "+err.Error())
		close(ch)
		return nil, err
	}

	// Send a heartbeat every 30 seconds to keep the connection alive
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				Logs.Info(ctx, "Subscription heartbeat - connection is alive")
			case <-ctx.Done():
				return
			}
		}
	}()

	// Handle subscription cleanup
	go func() {
		<-ctx.Done()
		Logs.Info(ctx, "Context done, cleaning up subscription")
		if err := sub.Unsubscribe(); err != nil {
			Logs.Error(ctx, "Failed to unsubscribe: "+err.Error())
		}
		close(ch)
		Logs.Info(ctx, "Subscription cleanup completed")
	}()

	return ch, nil
}

// func (r *subscriptionResolver) OrderStatusChanged(ctx context.Context, orderID *string) (<-chan *OrderStatusUpdate, error) {
// 	Logs := logger.GetGlobalLogger()

// 	if orderID != nil {
// 		Logs.Info(ctx, "OrderStatusChanged called for specific order ID: " + *orderID)
// 	} else {
// 		Logs.Info(ctx, "OrderStatusChanged called for ALL orders")
// 	}

// 	ch := make(chan *OrderStatusUpdate, 10)

// 	sub, err := r.server.nats.Subscribe("order.status.changed", func(msg *nats.Msg) {
// 		Logs.Info(ctx, "Received order status changed event")

// 		var update OrderStatusUpdate
// 		if err := json.Unmarshal(msg.Data, &update); err != nil {
// 			Logs.Error(ctx, "Failed to unmarshal order status update: " + err.Error())
// 			return
// 		}

// 		if orderID == nil || update.OrderID == *orderID {
// 			Logs.Info(ctx, "Publishing update for order ID: " + update.OrderID)
// 			ch <- &update
// 		}
// 	})
// 	if err != nil {
// 		Logs.Error(ctx, "Failed to subscribe to order status changed event: " + err.Error())
// 		return nil, err
// 	}

// 	go func() {
// 		<-ctx.Done()
// 		Logs.Info(ctx, "Unsubscribing from order status changed event")
// 		sub.Unsubscribe()
// 		close(ch)
// 	}()

// 	return ch, nil
// }

// unc (r *subscriptionResolver) OrderStatusChanged(ctx context.Context, orderID string) (<-chan *OrderStatusUpdate, error) {
// 	Logs := logger.GetGlobalLogger()
// 	Logs.Info(ctx, "OrderStatusChanged called for order ID: "+orderID)

// 	ch := make(chan *OrderStatusUpdate, 10)

// 	js, err := r.server.nats.JetStream()
// 	if err != nil {
// 		Logs.Error(ctx, "Failed to get JetStream context: "+err.Error())
// 		return nil, err
// 	}

// 	sub, err := js.Subscribe("order.status.changed", func(msg *nats.Msg) {
// 		Logs.Info(ctx, "Received order status changed event")
// 		var update OrderStatusUpdate
// 		if err := json.Unmarshal(msg.Data, &update); err != nil {
// 			Logs.Error(ctx, "Failed to unmarshal order status update: "+err.Error())
// 			return
// 		}
// 		if update.OrderID == orderID {
// 			Logs.Info(ctx, "Order status changed for order ID: " + orderID)
// 			ch <- &update
// 		}
// 		msg.Ack() // ✅ acknowledge it manually
// 	},
// 		// nats.Durable("order-status-" + orderID), // ✅ unique durable name
// 		nats.ManualAck(),                         // ✅ required with durable
// 		nats.DeliverAll(),                        // ✅ replay from beginning
// 	)

// 	if err != nil {
// 		Logs.Error(ctx, "Failed to subscribe to order status changed event: "+err.Error())
// 		return nil, err
// 	}

// 	go func() {
// 		<-ctx.Done()
// 		sub.Unsubscribe()
// 		Logs.Info(ctx, "Unsubscribing from order status changed event for order ID: "+orderID)
// 		close(ch)
// 	}()

// 	return ch, nil
// }
