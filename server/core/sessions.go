package core

import (
	"math/rand"
	"sync"
	"time"

	"github.com/privacybydesign/gabi"
	"github.com/privacybydesign/gabi/big"
	"github.com/privacybydesign/irmago"
	"github.com/privacybydesign/irmago/server"
)

type session struct {
	sync.Mutex

	action   irma.Action
	token    string
	version  *irma.ProtocolVersion
	rrequest irma.RequestorRequest
	request  irma.SessionRequest

	status     server.Status
	lastActive time.Time
	returned   bool
	result     *server.SessionResult

	kssProofs map[irma.SchemeManagerIdentifier]*gabi.ProofP
}

type sessionStore interface {
	get(token string) *session
	add(token string, session *session)
	update(token string, session *session)
	deleteExpired()
}

type memorySessionStore struct {
	sync.RWMutex
	m map[string]*session
}

const (
	maxSessionLifetime = 5 * time.Minute  // After this a session is cancelled
	expiryTicker       = 10 * time.Second // Every so often we check if any session has expired
)

const sessionChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var (
	minProtocolVersion = irma.NewVersion(2, 4)
	maxProtocolVersion = irma.NewVersion(2, 4)

	sessions sessionStore = &memorySessionStore{
		m: make(map[string]*session),
	}
)

func init() {
	rand.Seed(time.Now().UnixNano())
	go sessions.deleteExpired()
}

func (s *memorySessionStore) get(token string) *session {
	s.RLock()
	defer s.RUnlock()
	return s.m[token]
}

func (s *memorySessionStore) add(token string, session *session) {
	s.Lock()
	defer s.Unlock()
	s.m[token] = session
}

func (s *memorySessionStore) update(token string, session *session) {
	// nop
}

func (s memorySessionStore) deleteExpired() {
	// First check which sessions have expired
	// We don't need a write lock for this yet, so postpone that for actual deleting
	s.RLock()
	expired := make([]string, 0, len(s.m))
	for token, session := range s.m {
		session.Lock()

		timeout := 5 * time.Minute
		if session.status == server.StatusInitialized && session.rrequest.Base().ClientTimeout != 0 {
			timeout = time.Duration(session.rrequest.Base().ClientTimeout) * time.Second
		}

		if session.lastActive.Add(timeout).Before(time.Now()) {
			if !session.finished() {
				conf.Logger.Infof("Session %s expired", token)
				session.markAlive()
				session.setStatus(server.StatusTimeout)
			} else {
				conf.Logger.Infof("Deleting %s", token)
				expired = append(expired, token)
			}
		}
		session.Unlock()
	}
	s.RUnlock()

	// Using a write lock, delete the expired sessions
	s.Lock()
	for _, token := range expired {
		delete(s.m, token)
	}
	s.Unlock()

	// Schedule next run
	time.AfterFunc(expiryTicker, func() {
		s.deleteExpired()
	})
}

var one *big.Int = big.NewInt(1)

func newSession(action irma.Action, request irma.RequestorRequest) *session {
	token := newSessionToken()
	s := &session{
		action:     action,
		rrequest:   request,
		request:    request.SessionRequest(),
		lastActive: time.Now(),
		token:      token,
		result: &server.SessionResult{
			Token: token,
			Type:  action,
		},
	}
	s.setStatus(server.StatusInitialized)
	nonce, _ := gabi.RandomBigInt(gabi.DefaultSystemParameters[2048].Lstatzk)
	s.request.SetNonce(nonce)
	s.request.SetContext(one)
	sessions.add(token, s)
	return s
}

func newSessionToken() string {
	b := make([]byte, 20)
	for i := range b {
		b[i] = sessionChars[rand.Int63()%int64(len(sessionChars))]
	}
	return string(b)
}
