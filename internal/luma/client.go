// Package luma is the only code that talks to the Luma API (docs/pcr/0003). It maps Luma's wire
// format to this service's own types, so nothing outside this package depends on Luma's JSON.
package luma

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	DefaultBaseURL = "https://public-api.luma.com"

	// pageSize is what we ask for per page; Luma enforces its own maximum.
	pageSize = 100
	// maxPages bounds a listing, so a pagination bug can't loop forever against the rate limit
	// (200 requests/minute per calendar key).
	maxPages = 200
)

type Client struct {
	apiKey  string
	baseURL string
	http    *http.Client
}

// New creates a client for the calendar the API key belongs to. baseURL is DefaultBaseURL outside tests.
func New(apiKey, baseURL string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: baseURL,
		// Bounded: net/http's zero-value client never times out.
		http: &http.Client{Timeout: 15 * time.Second},
	}
}

// APIError is a non-2xx response from Luma.
type APIError struct {
	Status int
	Body   string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("luma API: status %d: %s", e.Status, e.Body)
}

func (c *Client) GetCalendar(ctx context.Context) (Calendar, error) {
	var raw wireCalendar
	if err := c.get(ctx, "/v1/calendars/get", nil, &raw); err != nil {
		return Calendar{}, err
	}
	return raw.toCalendar(), nil
}

// ListEvents returns the calendar's events. Zero after/before means no bound.
func (c *Client) ListEvents(ctx context.Context, after, before time.Time) ([]Event, error) {
	query := url.Values{}
	if !after.IsZero() {
		query.Set("after", after.UTC().Format(time.RFC3339))
	}
	if !before.IsZero() {
		query.Set("before", before.UTC().Format(time.RFC3339))
	}
	raw, err := listAll[wireEvent](ctx, c, "/v1/calendars/events/list", query)
	if err != nil {
		return nil, err
	}
	events := make([]Event, 0, len(raw))
	for _, e := range raw {
		events = append(events, e.toEvent())
	}
	return events, nil
}

func (c *Client) ListTiers(ctx context.Context) ([]Tier, error) {
	raw, err := listAll[wireTier](ctx, c, "/v1/memberships/tiers/list", nil)
	if err != nil {
		return nil, err
	}
	tiers := make([]Tier, 0, len(raw))
	for _, t := range raw {
		tiers = append(tiers, t.toTier())
	}
	return tiers, nil
}

// ListMembers returns the approved members of a membership tier.
func (c *Client) ListMembers(ctx context.Context, tierID string) ([]Member, error) {
	query := url.Values{}
	query.Set("calendar_membership_tier_id", tierID)
	query.Set("membership_status", "approved")
	raw, err := listAll[wireContact](ctx, c, "/v1/calendars/contacts/list", query)
	if err != nil {
		return nil, err
	}
	members := make([]Member, 0, len(raw))
	for _, contact := range raw {
		members = append(members, contact.toMember())
	}
	return members, nil
}

type page[T any] struct {
	Entries    []T    `json:"entries"`
	HasMore    bool   `json:"has_more"`
	NextCursor string `json:"next_cursor"`
}

// listAll follows Luma's cursor pagination to the end.
func listAll[T any](ctx context.Context, c *Client, path string, query url.Values) ([]T, error) {
	if query == nil {
		query = url.Values{}
	}
	query.Set("pagination_limit", strconv.Itoa(pageSize))

	var all []T
	for range maxPages {
		var p page[T]
		if err := c.get(ctx, path, query, &p); err != nil {
			return nil, err
		}
		all = append(all, p.Entries...)
		if !p.HasMore || p.NextCursor == "" {
			return all, nil
		}
		query.Set("pagination_cursor", p.NextCursor)
	}
	return nil, fmt.Errorf("luma API: %s: more than %d pages", path, maxPages)
}

func (c *Client) get(ctx context.Context, path string, query url.Values, out any) error {
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("x-luma-api-key", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("luma API: %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return &APIError{Status: resp.StatusCode, Body: string(body)}
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("luma API: %s: decoding response: %w", path, err)
	}
	return nil
}
