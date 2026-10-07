package luma

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListMembersFollowsPagesAndFlattensAnswers(t *testing.T) {
	var cursors []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/calendars/contacts/list", r.URL.Path)
		assert.Equal(t, "test-key", r.Header.Get("x-luma-api-key"))
		assert.Equal(t, "tier-community", r.URL.Query().Get("calendar_membership_tier_id"))
		assert.Equal(t, "approved", r.URL.Query().Get("membership_status"))
		cursors = append(cursors, r.URL.Query().Get("pagination_cursor"))

		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("pagination_cursor") == "" {
			_, _ = w.Write([]byte(`{"entries":[{"id":"c1","user_id":"u1","name":"Ada","first_name":"Ada","last_name":null,
				"email":"ada@example.com","created_at":"2026-09-01T10:00:00Z","membership":{"status":"approved",
				"registration_answers":[
					{"question_id":"q1","label":"Graduation year","question_type":"select","value":"2027"},
					{"question_id":"q2","label":"Areas of interest","question_type":"select","value":["NLP","Robotics"]},
					{"question_id":"q3","label":"Code of conduct","question_type":"agree-check","value":true},
					{"question_id":"q4","label":"Company","question_type":"company","value":{"company":"KTH","job_title":null}},
					{"question_id":"q5","label":"Anything else?","question_type":"text","value":null}
				]}}],"has_more":true,"next_cursor":"page-2"}`))
			return
		}
		_, _ = w.Write([]byte(`{"entries":[{"id":"c2","user_id":"u2","name":null,"first_name":null,"last_name":null,
			"email":"b@example.com","created_at":"2026-09-02T10:00:00Z","membership":null}],"has_more":false}`))
	}))
	defer srv.Close()

	members, err := New("test-key", srv.URL).ListMembers(context.Background(), "tier-community")
	require.NoError(t, err)

	assert.Equal(t, []string{"", "page-2"}, cursors)
	require.Len(t, members, 2)
	assert.Equal(t, "Ada", members[0].Name)
	assert.Equal(t, "", members[0].LastName)
	assert.Equal(t, []Answer{
		{QuestionID: "q1", Label: "Graduation year", Type: "select", Values: []string{"2027"}},
		{QuestionID: "q2", Label: "Areas of interest", Type: "select", Values: []string{"NLP", "Robotics"}},
		{QuestionID: "q3", Label: "Code of conduct", Type: "agree-check", Values: []string{"true"}},
		{QuestionID: "q4", Label: "Company", Type: "company", Values: []string{"KTH"}},
		{QuestionID: "q5", Label: "Anything else?", Type: "text", Values: []string{}},
	}, members[0].Answers)
	assert.Equal(t, []Answer{}, members[1].Answers)
}

func TestNon2xxIsAnAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"invalid api key"}`, http.StatusUnauthorized)
	}))
	defer srv.Close()

	_, err := New("bad-key", srv.URL).ListTiers(context.Background())

	var apiErr *APIError
	require.True(t, errors.As(err, &apiErr))
	assert.Equal(t, http.StatusUnauthorized, apiErr.Status)
}
