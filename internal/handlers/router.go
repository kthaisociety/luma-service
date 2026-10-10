// Package handlers is the HTTP layer. It depends on the interfaces below, not on concrete clients,
// so tests can run it against fakes.
package handlers

import (
	"context"
	"crypto/subtle"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/kthaisociety/luma-service/internal/luma"
)

// LumaReader is what the handlers need from Luma. *luma.Client implements it.
type LumaReader interface {
	GetCalendar(ctx context.Context) (luma.Calendar, error)
	ListEvents(ctx context.Context, after, before time.Time) ([]luma.Event, error)
	ListTiers(ctx context.Context) ([]luma.Tier, error)
	ListMembers(ctx context.Context, tierID string) ([]luma.Member, error)
}

type Deps struct {
	Luma            LumaReader
	PingDB          func(ctx context.Context) error
	ServiceSecret   string
	CommunityTierID string
}

func NewRouter(d Deps) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/health", health(d.PingDB))

	v1 := r.Group("/v1", requireServiceSecret(d.ServiceSecret))
	v1.GET("/calendar", getCalendar(d.Luma))
	v1.GET("/events", listEvents(d.Luma))
	v1.GET("/tiers", listTiers(d.Luma))
	v1.GET("/community/members", listMembers(d.Luma, d.CommunityTierID))

	return r
}

// requireServiceSecret rejects any request without the exact shared secret (docs/pcr/0004). An empty
// configured secret rejects everything: it never means "no secret required".
func requireServiceSecret(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		provided := c.GetHeader("X-Service-Secret")
		if secret == "" || provided == "" ||
			subtle.ConstantTimeCompare([]byte(provided), []byte(secret)) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Next()
	}
}

func health(pingDB func(ctx context.Context) error) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := pingDB(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"service": "luma-service", "status": "unhealthy", "database": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"service": "luma-service", "status": "healthy"})
	}
}

func getCalendar(l LumaReader) gin.HandlerFunc {
	return func(c *gin.Context) {
		calendar, err := l.GetCalendar(c.Request.Context())
		if err != nil {
			lumaError(c, err)
			return
		}
		c.JSON(http.StatusOK, calendar)
	}
}

// listEvents takes optional after/before query parameters, RFC 3339.
func listEvents(l LumaReader) gin.HandlerFunc {
	return func(c *gin.Context) {
		after, ok := timeParam(c, "after")
		if !ok {
			return
		}
		before, ok := timeParam(c, "before")
		if !ok {
			return
		}
		events, err := l.ListEvents(c.Request.Context(), after, before)
		if err != nil {
			lumaError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"events": events})
	}
}

func listTiers(l LumaReader) gin.HandlerFunc {
	return func(c *gin.Context) {
		tiers, err := l.ListTiers(c.Request.Context())
		if err != nil {
			lumaError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"tiers": tiers})
	}
}

func listMembers(l LumaReader, tierID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		members, err := l.ListMembers(c.Request.Context(), tierID)
		if err != nil {
			lumaError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"tier_id": tierID, "members": members})
	}
}

func timeParam(c *gin.Context, name string) (time.Time, bool) {
	raw := c.Query(name)
	if raw == "" {
		return time.Time{}, true
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": name + " must be an RFC 3339 time"})
		return time.Time{}, false
	}
	return t, true
}

// lumaError answers 502: the failure is upstream. The detail goes to the log, not to the caller.
func lumaError(c *gin.Context, err error) {
	_ = c.Error(err)
	c.JSON(http.StatusBadGateway, gin.H{"error": "luma request failed"})
}
