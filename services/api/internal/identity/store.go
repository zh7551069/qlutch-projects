package identity

import (
	"context"
	"errors"
)

// errAccountExists is returned when a phone number is already bound.
var errAccountExists = errors.New("account exists")

// Store persists learner accounts, onboarding answers, locale, region
// preference, and session token hashes.
// OTP challenges, rate limits, and idempotency records stay in the API process.
// SaveProfile must not change the bound login phone.
type Store interface {
	FindByPhone(ctx context.Context, region, national string) (*user, error)
	FindByID(ctx context.Context, id string) (*user, error)
	Insert(ctx context.Context, account *user) error
	DeleteAccount(ctx context.Context, id string) error
	SaveProfile(ctx context.Context, id, locale, regionPref string) error
	SaveAnswer(ctx context.Context, id, step, answer string) error
	SaveFirstPack(ctx context.Context, id string, has bool) error
	InsertSession(ctx context.Context, sess *session) error
	DeleteSession(ctx context.Context, id string) error
	FindSessionByAccess(ctx context.Context, hash [32]byte) (*session, error)
	FindSessionByRefresh(ctx context.Context, hash [32]byte) (*session, error)
}

func cloneUser(in *user) *user {
	if in == nil {
		return nil
	}
	out := *in
	out.Answers = map[string]string{}
	for key, value := range in.Answers {
		out.Answers[key] = value
	}
	return &out
}

func cloneSession(in *session) *session {
	if in == nil {
		return nil
	}
	out := *in
	return &out
}
