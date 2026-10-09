package luma

import (
	"encoding/json"
	"time"
)

// The service's own types: what handlers use and what callers receive as JSON.

type Calendar struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	URL  string `json:"url"`
}

type Event struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	StartAt  time.Time `json:"start_at"`
	EndAt    time.Time `json:"end_at"`
	Timezone string    `json:"timezone"`
	URL      string    `json:"url"`
	// Platform is "luma" for Luma events and something else for external events listed on the calendar.
	Platform string `json:"platform"`
}

type Tier struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Questions   []Question `json:"questions"`
}

// Question is one registration question members answer when they join a tier.
type Question struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Type  string `json:"type"`
}

type Member struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	Answers   []Answer  `json:"answers"`
}

// Answer is a member's answer to a registration question. Values is always a list (docs/adr/0001):
// one item for single-choice and text answers, several for multi-select, none when unanswered.
type Answer struct {
	QuestionID string   `json:"question_id"`
	Label      string   `json:"label"`
	Type       string   `json:"type"`
	Values     []string `json:"values"`
}

// Luma's wire format, decoded only as far as the types above need.

type wireCalendar struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	URL  string `json:"url"`
}

func (w wireCalendar) toCalendar() Calendar {
	return Calendar{ID: w.ID, Name: w.Name, Slug: w.Slug, URL: w.URL}
}

type wireEvent struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	StartAt  time.Time `json:"start_at"`
	EndAt    time.Time `json:"end_at"`
	Timezone string    `json:"timezone"`
	URL      string    `json:"url"`
	Platform string    `json:"platform"`
}

func (w wireEvent) toEvent() Event {
	return Event(w)
}

type wireTier struct {
	ID                    string  `json:"id"`
	Name                  string  `json:"name"`
	Description           *string `json:"description"`
	RegistrationQuestions []struct {
		ID           string `json:"id"`
		Label        string `json:"label"`
		QuestionType string `json:"question_type"`
	} `json:"registration_questions"`
}

func (w wireTier) toTier() Tier {
	t := Tier{ID: w.ID, Name: w.Name, Questions: []Question{}}
	if w.Description != nil {
		t.Description = *w.Description
	}
	for _, q := range w.RegistrationQuestions {
		t.Questions = append(t.Questions, Question{ID: q.ID, Label: q.Label, Type: q.QuestionType})
	}
	return t
}

type wireContact struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Name       *string   `json:"name"`
	FirstName  *string   `json:"first_name"`
	LastName   *string   `json:"last_name"`
	Email      string    `json:"email"`
	CreatedAt  time.Time `json:"created_at"`
	Membership *struct {
		RegistrationAnswers []wireAnswer `json:"registration_answers"`
	} `json:"membership"`
}

func (w wireContact) toMember() Member {
	m := Member{
		ID:        w.ID,
		UserID:    w.UserID,
		Name:      deref(w.Name),
		FirstName: deref(w.FirstName),
		LastName:  deref(w.LastName),
		Email:     w.Email,
		CreatedAt: w.CreatedAt,
		Answers:   []Answer{},
	}
	if w.Membership != nil {
		for _, a := range w.Membership.RegistrationAnswers {
			m.Answers = append(m.Answers, a.toAnswer())
		}
	}
	return m
}

type wireAnswer struct {
	QuestionID   string          `json:"question_id"`
	Label        string          `json:"label"`
	QuestionType string          `json:"question_type"`
	Value        json.RawMessage `json:"value"`
}

func (w wireAnswer) toAnswer() Answer {
	return Answer{
		QuestionID: w.QuestionID,
		Label:      w.Label,
		Type:       w.QuestionType,
		Values:     answerValues(w.Value),
	}
}

// answerValues flattens Luma's per-type answer values (string, list of strings, boolean, or a
// company object) into a list of strings. null and empty answers become an empty list.
func answerValues(raw json.RawMessage) []string {
	var s string
	if json.Unmarshal(raw, &s) == nil {
		if s == "" {
			return []string{}
		}
		return []string{s}
	}
	var list []string
	if json.Unmarshal(raw, &list) == nil && list != nil {
		return list
	}
	var b bool
	if json.Unmarshal(raw, &b) == nil {
		if b {
			return []string{"true"}
		}
		return []string{"false"}
	}
	// A "company" answer: only the company name is kept for now (docs/adr/0001).
	var company struct {
		Company *string `json:"company"`
	}
	if json.Unmarshal(raw, &company) == nil && company.Company != nil && *company.Company != "" {
		return []string{*company.Company}
	}
	return []string{}
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
