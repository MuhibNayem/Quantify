package message_broker

const (
	ExchangeName = "inventory"

	QueueBulkImport        = "bulk-import-queue"
	QueueBulkImportConfirm = "bulk-import-confirm-queue"
	QueueBulkExport        = "bulk-export-queue"
	QueueStockAlerts       = "stock-alerts-queue"
	QueueAlerts            = "alerts"
	QueueReporting         = "reporting"

	RoutingBulkImport        = "bulk.import"
	RoutingBulkImportConfirm = "bulk.import.confirm"
	RoutingBulkExport        = "bulk.export"
	RoutingStockAdjusted     = "stock.adjusted"
	RoutingAlertTriggered    = "alert.triggered"
	RoutingReportGenerate    = "report.generate"
	RoutingProductCreated    = "product.created"
	RoutingProductUpdated    = "product.updated"
	RoutingProductDeleted    = "product.deleted"
)

// Topology declares the queue/routing-key mapping for the inventory system.
type Topology struct {
	Exchange string
	Queues   map[string]string
}

func DefaultTopology() Topology {
	return Topology{
		Exchange: ExchangeName,
		Queues: map[string]string{
			QueueBulkImport:        RoutingBulkImport,
			QueueBulkImportConfirm: RoutingBulkImportConfirm,
			QueueBulkExport:        RoutingBulkExport,
			QueueStockAlerts:       RoutingStockAdjusted,
			QueueAlerts:            RoutingAlertTriggered,
			QueueReporting:         RoutingReportGenerate,
		},
	}
}
