// Package identity implements the learner OTP, session, onboarding, and profile API.
// SMS is mocked. Accounts, onboarding answers, locale, region preference, and
// session token hashes go through Store. OTP challenges, rate limits, and
// idempotency records stay in process memory.
// See docs/contracts/identity-a2.md and docs/contracts/identity-a3.md.
package identity
