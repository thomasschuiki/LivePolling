package websocket

import (
	"context"
	"encoding/json"
	"log"
)

func HandleMessage(client *Client, pool *Pool, store *Store, data []byte) {
	var cmd Command
	if err := json.Unmarshal(data, &cmd); err != nil {
		log.Printf("error unmarshaling command: %v", err)
		return
	}

	switch cmd.Command {
	case "init":
		var initData string
		if err := json.Unmarshal(cmd.Data, &initData); err != nil {
			log.Printf("error unmarshaling init: %v", err)
			return
		}
		handleInit(client, store, initData == "admin")
	case "vote":
		handleVote(client, store, pool, cmd.Data)
	case "login":
		handleLogin(client, store, cmd.Data)
	case "statistics":
		handleStatistics(client, store)
	case "question":
		handleQuestion(client, store, pool, cmd.Data)
	case "clear":
		handleClear(client, store, pool)
	case "reset":
		handleReset(client, store, pool)
	default:
		log.Printf("unknown command: %s", cmd.Command)
	}
}

func handleInit(client *Client, store *Store, isAdmin bool) {
	if isAdmin {
		if client.Authenticated {
			client.Send(context.Background(), Message{Type: "login", Data: toRawJSON(true)})
		} else {
			client.Send(context.Background(), Message{Type: "login", Data: toRawJSON(false)})
		}
	} else {
		question := store.GetQuestion()
		if question != nil {
			client.Send(context.Background(), Message{
				Type: "asking",
				Question: question.Text,
				Possibilities: question.Choices,
			})
		}
	}
}

func handleVote(client *Client, store *Store, pool *Pool, data json.RawMessage) {
	var vote VoteData
	if err := json.Unmarshal(data, &vote); err != nil {
		log.Printf("error unmarshaling vote: %v", err)
		return
	}

	question := store.GetQuestion()
	if question == nil {
		client.Send(context.Background(), Message{Type: "voted", Data: json.RawMessage(`"No active question"`)})
		return
	}

	if vote.Votenumber < 0 || vote.Votenumber >= len(question.Votes) {
		client.Send(context.Background(), Message{Type: "voted", Data: json.RawMessage(`"Invalid choice"`)})
		return
	}

	if store.HasVoted(client.ID) {
		client.Send(context.Background(), Message{Type: "voted", Data: json.RawMessage(`"You already voted!"`)})
		return
	}

	store.mu.Lock()
	question.Votes[vote.Votenumber]++
	store.mu.Unlock()

	store.MarkVoted(client.ID)

	client.Send(context.Background(), Message{Type: "voted", Data: json.RawMessage(`"Thanks for voting!"`)})

	broadcastStats(store, pool)
}

func handleLogin(client *Client, store *Store, data json.RawMessage) {
	var login LoginData
	if err := json.Unmarshal(data, &login); err != nil {
		log.Printf("error unmarshaling login: %v", err)
		return
	}

	var msg string
	var success bool

	if login.Adminname == AdminUsername && login.Adminpass == AdminPassword {
		client.IsAdmin = true
		client.Authenticated = true
		store.AddAdmin(client)
		msg = "you are now logged in"
		success = true
	} else {
		msg = "the password or the username didn't match"
		success = false
	}

	client.Send(context.Background(), Message{
		Type: "msg",
		Data: json.RawMessage(`"` + msg + `"`),
	})
	client.Send(context.Background(), Message{
		Type: "login",
		Data: toRawJSON(success),
	})
}

func handleStatistics(client *Client, store *Store) {
	question := store.GetQuestion()
	if question == nil {
		client.Send(context.Background(), Message{Type: "msg", Data: json.RawMessage(`"No active question"`)})
		return
	}
	broadcastToClient(client, question)
}

func handleQuestion(client *Client, store *Store, pool *Pool, data json.RawMessage) {
	if !client.Authenticated {
		client.Send(context.Background(), Message{Type: "msg", Data: json.RawMessage(`"Not authenticated"`)})
		return
	}

	var qdata QuestionData
	if err := json.Unmarshal(data, &qdata); err != nil {
		log.Printf("error unmarshaling question: %v", err)
		return
	}

	store.SetQuestion(qdata.Question, qdata.Possibilities)
	store.ClearVotes()
	store.ClearVotedClients()

	question := store.GetQuestion()
	for c := range pool.Clients {
		c.Send(context.Background(), Message{
			Type:          "asking",
			Question:      question.Text,
			Possibilities: question.Choices,
		})
	}
}

func handleClear(client *Client, store *Store, pool *Pool) {
	if !client.Authenticated {
		client.Send(context.Background(), Message{Type: "msg", Data: json.RawMessage(`"Not authenticated"`)})
		return
	}

	store.ClearQuestion()
	store.ClearVotes()
	store.ClearVotedClients()

	for c := range pool.Clients {
		c.Send(context.Background(), Message{Type: "clear", Data: json.RawMessage(`"Question cleared"`)})
	}
}

func handleReset(client *Client, store *Store, pool *Pool) {
	if !client.Authenticated {
		client.Send(context.Background(), Message{Type: "msg", Data: json.RawMessage(`"Not authenticated"`)})
		return
	}

	store.ClearVotes()
	store.ClearVotedClients()

	broadcastStats(store, pool)
}

func broadcastStats(store *Store, pool *Pool) {
	question := store.GetQuestion()
	if question == nil {
		return
	}

	for admin := range store.GetAdmins() {
		broadcastToClient(admin, question)
	}
}

func broadcastToClient(client *Client, question *Question) {
	client.Send(context.Background(), Message{
		Type:          "statistics",
		Text:          question.Text,
		Choices:       question.Choices,
		Votes:         question.Votes,
		ID:            question.ID,
	})
}

func boolToJSON(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func toRawJSON(v interface{}) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}
