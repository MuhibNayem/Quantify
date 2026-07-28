package message_broker

import (
	"context"
	"testing"

	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/require"
)

func TestDefaultTopology(t *testing.T) {
	topology := DefaultTopology()

	require.Equal(t, ExchangeName, topology.Exchange)
	require.Equal(t, map[string]string{
		QueueBulkImport:        RoutingBulkImport,
		QueueBulkImportConfirm: RoutingBulkImportConfirm,
		QueueBulkExport:        RoutingBulkExport,
		QueueStockAlerts:       RoutingStockAdjusted,
		QueueAlerts:            RoutingAlertTriggered,
		QueueReporting:         RoutingReportGenerate,
	}, topology.Queues)
}

func TestSubscribeRejectsUnknownQueue(t *testing.T) {
	// Subscribe requires exchange, queueName, routingKey. Unknown queue returns nil cancel.
	cancel := Subscribe(context.Background(), ExchangeName, "unknown-queue", "unknown.routing", func(context.Context, <-chan amqp091.Delivery) {})
	require.Nil(t, cancel)
}
