package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"github.com/creafly/notifications/internal/domain/service"
)

type InvitationRequestedEvent struct {
	TenantID    uuid.UUID `json:"tenantId"`
	TenantName  string    `json:"tenantName"`
	InviterID   uuid.UUID `json:"inviterId"`
	InviterName string    `json:"inviterName"`
	InviteeID   uuid.UUID `json:"inviteeId"`
	Email       string    `json:"email"`
}

type InvitationsConsumer struct {
	invitationService service.InvitationService
	consumer          sarama.ConsumerGroup
	topics            []string
	ready             chan bool
}

func NewInvitationsConsumer(
	brokers []string,
	groupID string,
	invitationService service.InvitationService,
) (*InvitationsConsumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, err
	}

	return &InvitationsConsumer{
		invitationService: invitationService,
		consumer:          consumer,
		topics:            []string{"invitations"},
		ready:             make(chan bool),
	}, nil
}

func (c *InvitationsConsumer) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if err := c.consumer.Consume(ctx, c.topics, c); err != nil {
					log.Printf("[InvitationsConsumer] Error consuming: %v", err)
				}
				if ctx.Err() != nil {
					return
				}
				c.ready = make(chan bool)
			}
		}
	}()

	<-c.ready
	log.Println("[InvitationsConsumer] Started and ready")
}

func (c *InvitationsConsumer) Stop() error {
	return c.consumer.Close()
}

func (c *InvitationsConsumer) Setup(sarama.ConsumerGroupSession) error {
	close(c.ready)
	return nil
}

func (c *InvitationsConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *InvitationsConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				log.Println("[InvitationsConsumer] Message channel was closed")
				return nil
			}

			if err := c.processMessage(session.Context(), message); err != nil {
				log.Printf("[InvitationsConsumer] Error processing message: %v", err)
			}

			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}

func (c *InvitationsConsumer) processMessage(ctx context.Context, message *sarama.ConsumerMessage) error {
	var eventType string
	for _, header := range message.Headers {
		if string(header.Key) == "event_type" {
			eventType = string(header.Value)
			break
		}
	}

	log.Printf("[InvitationsConsumer] Received event: type=%s, key=%s", eventType, string(message.Key))

	switch eventType {
	case "invitations.requested":
		return c.handleInvitationRequested(ctx, message.Value)
	default:
		log.Printf("[InvitationsConsumer] Unknown/unhandled event type: %s", eventType)
		return nil
	}
}

func (c *InvitationsConsumer) handleInvitationRequested(ctx context.Context, value []byte) error {
	var event InvitationRequestedEvent
	if err := json.Unmarshal(value, &event); err != nil {
		log.Printf("[InvitationsConsumer] Error unmarshaling event: %v", err)
		return err
	}

	log.Printf("[InvitationsConsumer] Creating invitation for user %s to tenant %s", event.InviteeID, event.TenantID)

	_, err := c.invitationService.Create(ctx, service.CreateInvitationInput{
		TenantID:    event.TenantID,
		TenantName:  event.TenantName,
		InviterID:   event.InviterID,
		InviterName: event.InviterName,
		InviteeID:   event.InviteeID,
		Email:       event.Email,
	})
	if err != nil {
		log.Printf("[InvitationsConsumer] Error creating invitation: %v", err)
		return err
	}

	log.Printf("[InvitationsConsumer] Invitation created successfully for user %s", event.InviteeID)
	return nil
}
