package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kthaisociety/luma-service/internal/luma"
)

type fakeLuma struct {
	members   []luma.Member
	gotTierID string
	err       error
}

func (f *fakeLuma) GetCalendar(context.Context) (luma.Calendar, error) {
	return luma.Calendar{ID: "cal-1"}, f.err
}

func (f *fakeLuma) ListEvents(context.Context, time.Time, time.Time) ([]luma.Event, error) {
	return []luma.Event{}, f.err
}

func (f *fakeLuma) ListTiers(context.Context) ([]luma.Tier, error) {
	return []luma.Tier{}, f.err
}

func (f *fakeLuma) ListMembers(_ context.Context, tierID string) ([]luma.Member, error) {
	f.gotTierID = tierID
	return f.members, f.err
}

func newTestRouter(l LumaReader) *gin.Engine {
	gin.SetMode(gin.TestMode)
	return NewRouter(Deps{
		Luma:            l,
		PingDB:          func(context.Context) error { return nil },
		ServiceSecret:   "s3cret",
		CommunityTierID: "tier-community",
	})
}

func get(r http.Handler, path, secret string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if secret != "" {
		req.Header.Set("X-Service-Secret", secret)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestV1RequiresTheServiceSecret(t *testing.T) {
	r := newTestRouter(&fakeLuma{})

	assert.Equal(t, http.StatusUnauthorized, get(r, "/v1/tiers", "").Code)
	assert.Equal(t, http.StatusUnauthorized, get(r, "/v1/tiers", "wrong").Code)
	assert.Equal(t, http.StatusOK, get(r, "/v1/tiers", "s3cret").Code)
}

func TestEmptyConfiguredSecretRejectsEverything(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := NewRouter(Deps{Luma: &fakeLuma{}, PingDB: func(context.Context) error { return nil }})

	assert.Equal(t, http.StatusUnauthorized, get(r, "/v1/tiers", "anything").Code)
}

func TestHealthNeedsNoSecret(t *testing.T) {
	assert.Equal(t, http.StatusOK, get(newTestRouter(&fakeLuma{}), "/health", "").Code)
}

func TestCommunityMembersUsesTheConfiguredTier(t *testing.T) {
	fake := &fakeLuma{members: []luma.Member{{ID: "c1", Answers: []luma.Answer{}}}}

	w := get(newTestRouter(fake), "/v1/community/members", "s3cret")

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "tier-community", fake.gotTierID)
	var body struct {
		TierID  string        `json:"tier_id"`
		Members []luma.Member `json:"members"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "tier-community", body.TierID)
	assert.Len(t, body.Members, 1)
}

func TestLumaFailureIsABadGateway(t *testing.T) {
	w := get(newTestRouter(&fakeLuma{err: errors.New("boom")}), "/v1/community/members", "s3cret")

	assert.Equal(t, http.StatusBadGateway, w.Code)
	assert.NotContains(t, w.Body.String(), "boom")
}

func TestEventsRejectsABadTime(t *testing.T) {
	w := get(newTestRouter(&fakeLuma{}), "/v1/events?after=yesterday", "s3cret")

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
