package fake

import (
	"context"
	"math/rand"
	"fmt"
	"log/slog"
	"time"

	"github.com/BahramRousta/notification-service/internal/notification/domain"
)



type Provider struct{
	logger *slog.Logger
}


func New(l *slog.Logger)*Provider{
	return &Provider{
		logger: l,
	}
}


func (p *Provider) Send (ctx context.Context, msg *domain.Message)(string, error){
	fakeID := fmt.Sprintf("fake-%d-%04d", time.Now().UnixMilli(), rand.Intn(9999))

	p.logger.Info("[fake provider] SMS would be sent here",
		slog.String("msg_id", msg.ID),
		slog.String("recipient", msg.Recipient),
		slog.String("body", msg.Body),
		slog.String("provider_msg_id", fakeID),
		slog.String("channel", string(msg.Channel)),
	)
	
	return fakeID, nil
}