package consumers

import (
	"context"
	"encoding/json"
	"inventory/backend/internal/domain"
	"inventory/backend/internal/message_broker"
	"inventory/backend/internal/repository"
	"inventory/backend/internal/services"

	"github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

type ReportingConsumer struct {
	ReportingService *services.ReportingService
	JobRepo          *repository.JobRepository
}

func NewReportingConsumer(reportingService *services.ReportingService, jobRepo *repository.JobRepository) *ReportingConsumer {
	return &ReportingConsumer{
		ReportingService: reportingService,
		JobRepo:          jobRepo,
	}
}

func (c *ReportingConsumer) Start(ctx context.Context) context.CancelFunc {
	return message_broker.Subscribe(ctx, "inventory", "reporting", "report.generate", c.handleReportingDelivery)
}

func (c *ReportingConsumer) handleReportingDelivery(ctx context.Context, deliveries <-chan amqp091.Delivery) {
	for {
		select {
		case <-ctx.Done():
			return
		case d, ok := <-deliveries:
			if !ok {
				return
			}
			c.processReportingDelivery(d)
		}
	}
}

func (c *ReportingConsumer) processReportingDelivery(d amqp091.Delivery) {
	var job domain.Job
	if err := json.Unmarshal(d.Body, &job); err != nil {
		logrus.Errorf("Failed to unmarshal job payload: %v", err)
		d.Nack(false, false) // Negative acknowledgement, don't requeue
		return
	}

	logrus.Infof("Handling reporting delivery for job %d: %s", job.ID, job.Type)

	if err := c.ReportingService.GenerateReport(&job); err != nil {
		logrus.Errorf("Failed to generate report for job %d: %v", job.ID, err)
		job.Status = "FAILED"
		job.LastError = err.Error()
		if err := c.JobRepo.UpdateJob(&job); err != nil {
			logrus.Errorf("Failed to update job status for job %d: %v", job.ID, err)
		}
		d.Nack(false, false) // Nack and don't requeue
		return
	}

	d.Ack(false) // Acknowledge the message
}
