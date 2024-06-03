package websockets

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TODO : also store accessToken and store it in client struct, validate on every request
type WSTicket struct {
	ticket    string
	userId    string
	createdAt time.Time
}

type Tickets map[string]*WSTicket

// CreateNewTicketsMap will create a new ticketmap and start the retention given the set period
func CreateNewTicketsMap(ctx context.Context, retentionPeriod time.Duration) *Tickets {
	ticketManager := make(Tickets)
	go ticketManager.discardOldTickets(ctx, retentionPeriod)

	return &ticketManager
}

func (t *Tickets) generateTicket(ctx *gin.Context) {
	wsTicket := &WSTicket{
		ticket:    uuid.NewString(), //userId+IP+timestamp(micros) OR uuid
		userId:    ctx.Keys["userId"].(string),
		createdAt: time.Now(),
	}

	(*t)[wsTicket.ticket] = wsTicket

	fmt.Printf("ticket = %v", wsTicket)

	ctx.JSON(http.StatusOK, gin.H{
		"ticket": wsTicket.ticket,
	})
}

func (t *Tickets) validateTicket(ctx *gin.Context) (string, bool) {
	ticket := ctx.Query("ticket")

	if _, ok := (*t)[ticket]; !ok {
		return "", false
	}

	userId := (*t)[ticket].userId

	delete(*t, ticket)

	return userId, true
}

func (t *Tickets) discardOldTickets(ctx context.Context, retentionPeriod time.Duration) {
	ticker := time.NewTicker(time.Second * 5)

	for {
		select {
		case <-ticker.C:
			for _, ticket := range *t {
				// Add Retention to Created and check if it is expired
				if ticket.createdAt.Add(retentionPeriod).Before(time.Now()) {
					delete(*t, ticket.ticket)
				}
			}
		case <-ctx.Done():
			return
		}
	}
}
