package websocket

import (
	"encoding/json"
	"sync"
)

type Question struct {
	Text       string   `json:"question,omitempty"`
	Choices    []string `json:"possibilities,omitempty"`
	Votes      []int    `json:"votes,omitempty"`
	ID         string   `json:"id,omitempty"`
}

type ClientState struct {
	Voted        bool
	IsAdmin      bool
	Authenticated bool
}

type Command struct {
	Command string          `json:"command"`
	Data    json.RawMessage `json:"data"`
}

type VoteData struct {
	Votenumber int `json:"votenumber"`
}

type LoginData struct {
	Adminname string `json:"adminname"`
	Adminpass string `json:"adminpass"`
}

type QuestionData struct {
	Question     string   `json:"question"`
	Possibilities []string `json:"possibilities"`
}

var (
	AdminUsername = "admin"
	AdminPassword = "admin"
)

type Store struct {
	Question    *Question
	VotedClients map[string]bool
	Admins      map[*Client]bool
	mu          sync.RWMutex
}

func NewStore() *Store {
	return &Store{
		VotedClients: make(map[string]bool),
		Admins:       make(map[*Client]bool),
	}
}

func (s *Store) SetQuestion(text string, choices []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Question = &Question{
		Text:    text,
		Choices: choices,
		Votes:   make([]int, len(choices)),
		ID:      generateID(),
	}
}

func (s *Store) GetQuestion() *Question {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Question
}

func (s *Store) ClearQuestion() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Question = nil
}

func (s *Store) HasVoted(clientID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.VotedClients[clientID]
}

func (s *Store) MarkVoted(clientID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.VotedClients[clientID] = true
}

func (s *Store) ClearVotes() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Question != nil {
		s.Question.Votes = make([]int, len(s.Question.Choices))
	}
}

func (s *Store) ClearVotedClients() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.VotedClients = make(map[string]bool)
}

func (s *Store) AddAdmin(client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Admins[client] = true
}

func (s *Store) RemoveAdmin(client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.Admins, client)
}

func (s *Store) GetAdmins() map[*Client]bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Admins
}

func generateID() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = letters[i%len(letters)]
	}
	return string(b)
}
