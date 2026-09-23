package connectionhub

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// newTestHub returns a fresh, isolated connectionHub instance so tests don't
// share state with each other or with the package-level ConnectionHub.
func newTestHub() *connectionHub {
	return &connectionHub{
		Rooms: ConnectionNamespaces{
			Roles:        make(map[string]map[*Connection]*Client),
			Organization: make(map[string]map[*Connection]*Client),
			Usernames:    make(map[string]map[*Connection]*Client),
		},
		Emit:       make(chan Message),
		Broadcast:  make(chan Message),
		Register:   make(chan Client),
		Unregister: make(chan Client),
		Clients:    make(clients),
	}
}

func newTestClient(user, org, username string) Client {
	return Client{
		User:         user,
		Organization: org,
		Username:     username,
		Roles:        []string{},
		Conn:         &Connection{Send: make(chan []byte, 1)},
	}
}

// sendOrTimeout sends to an unbuffered channel and fails the test if the
// send blocks, instead of hanging the test suite forever. This is how we
// detect a dead hub goroutine (the exact failure mode of the bug below).
func sendOrTimeout[T any](t *testing.T, ch chan<- T, value T, msg string) {
	t.Helper()
	select {
	case ch <- value:
	case <-time.After(time.Second):
		t.Fatal(msg)
	}
}

// Regression test: a bare `return` in the Broadcast case of Run() used to
// exit the hub goroutine permanently on the first broadcast message. Since
// Register/Unregister/Emit are unbuffered with no other receiver, every
// subsequent send would then block forever, silently breaking WS delivery
// for every user on the pod. Run() must survive a Broadcast message.
func TestRunSurvivesBroadcastMessage(t *testing.T) {
	h := newTestHub()
	go h.Run()

	sendOrTimeout(t, h.Broadcast, Message{Origin: "test-source"}, "sending a Broadcast message blocked — hub did not accept it")

	client := newTestClient("user-1", "org-1", "username-1")
	sendOrTimeout(t, h.Register, client, "Register blocked after a Broadcast message — hub goroutine died")

	// give Run() a moment to process the register before asserting state
	assert.Eventually(t, func() bool {
		return h.Clients["user-1"] != nil
	}, time.Second, 10*time.Millisecond, "client was never registered after hub received a broadcast message")
}

// A second Broadcast message must also be a no-op, not just the first.
func TestRunSurvivesMultipleBroadcastMessages(t *testing.T) {
	h := newTestHub()
	go h.Run()

	sendOrTimeout(t, h.Broadcast, Message{Origin: "source-1"}, "first Broadcast send blocked")
	sendOrTimeout(t, h.Broadcast, Message{Origin: "source-2"}, "second Broadcast send blocked — hub died after first broadcast")

	client := newTestClient("user-1", "org-1", "username-1")
	sendOrTimeout(t, h.Register, client, "Register blocked after two Broadcast messages — hub goroutine died")
}

func TestRegisterClientAddsToAllIndexes(t *testing.T) {
	h := newTestHub()
	client := newTestClient("user-1", "org-1", "username-1")

	registerClient(client, h)

	assert.NotNil(t, h.Clients["user-1"])
	assert.Contains(t, h.Rooms.Organization["org-1"], client.Conn)
	assert.Contains(t, h.Rooms.Usernames["username-1"], client.Conn)
}

func TestRegisterClientDoesNotOverwriteExistingUser(t *testing.T) {
	h := newTestHub()
	first := newTestClient("user-1", "org-1", "username-1")
	second := newTestClient("user-1", "org-2", "username-2")

	registerClient(first, h)
	registerClient(second, h)

	// registerClient only sets h.Clients[user] when it is nil, so the first
	// registration wins for direct user-targeted delivery.
	assert.Equal(t, "org-1", h.Clients["user-1"].Organization)
	// but room indexes (keyed by connection, not user) still pick up both.
	assert.Contains(t, h.Rooms.Organization["org-2"], second.Conn)
}

func TestUnregisterClientRemovesFromAllIndexes(t *testing.T) {
	h := newTestHub()
	client := newTestClient("user-1", "org-1", "username-1")
	registerClient(client, h)

	unregisterClient(client, h)

	assert.Nil(t, h.Clients["user-1"])
	assert.NotContains(t, h.Rooms.Organization["org-1"], client.Conn)
	assert.NotContains(t, h.Rooms.Usernames["username-1"], client.Conn)
}

func TestEmitMessageDeliversToMatchingUser(t *testing.T) {
	h := newTestHub()
	client := newTestClient("user-1", "org-1", "username-1")
	registerClient(client, h)

	payload := []byte(`{"hello":"world"}`)
	emitMessage(Message{
		Data:         payload,
		Destinations: MessageDestinations{Users: []string{"user-1"}},
	}, h)

	select {
	case got := <-client.Conn.Send:
		assert.Equal(t, payload, got)
	default:
		t.Fatal("expected message on client's Send channel, got none")
	}
}

func TestEmitMessageDeliversToMatchingOrgAndUsername(t *testing.T) {
	h := newTestHub()
	orgClient := newTestClient("user-1", "org-1", "username-1")
	usernameClient := newTestClient("user-2", "org-2", "username-2")
	registerClient(orgClient, h)
	registerClient(usernameClient, h)

	payload := []byte(`{"hello":"world"}`)
	emitMessage(Message{
		Data: payload,
		Destinations: MessageDestinations{
			Organizations: []string{"org-1"},
			Usernames:     []string{"username-2"},
		},
	}, h)

	select {
	case got := <-orgClient.Conn.Send:
		assert.Equal(t, payload, got)
	default:
		t.Fatal("expected message on org-matched client's Send channel, got none")
	}

	select {
	case got := <-usernameClient.Conn.Send:
		assert.Equal(t, payload, got)
	default:
		t.Fatal("expected message on username-matched client's Send channel, got none")
	}
}

func TestEmitMessageNoMatchingDestinationDoesNotPanic(t *testing.T) {
	h := newTestHub()

	assert.NotPanics(t, func() {
		emitMessage(Message{
			Data:         []byte(`{"hello":"world"}`),
			Destinations: MessageDestinations{Users: []string{"nobody-connected"}},
		}, h)
	})
}

// emitMessage unregisters a client whose Send channel is full (backpressure),
// treating it as an unresponsive/dead connection. Confirm that path runs
// through unregisterClient rather than panicking or deadlocking.
func TestEmitMessageUnregistersClientWithFullSendBuffer(t *testing.T) {
	h := newTestHub()
	client := Client{
		User:         "user-1",
		Organization: "org-1",
		Username:     "username-1",
		Roles:        []string{},
		Conn:         &Connection{Send: make(chan []byte)}, // unbuffered, no reader => always full
	}
	registerClient(client, h)

	emitMessage(Message{
		Data:         []byte(`{"hello":"world"}`),
		Destinations: MessageDestinations{Users: []string{"user-1"}},
	}, h)

	assert.Nil(t, h.Clients["user-1"])
}
