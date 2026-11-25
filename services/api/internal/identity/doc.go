// Package identity implements the learner OTP, session, onboarding, and profile API.
// SMS is mocked. Accounts, onboarding answers, locale, region preference,
// has_first_pack, and session token hashes go through Store. OTP challenges,
// rate limits, and idempotency records stay in process memory.
// The identity HTTP API does not write has_first_pack. The knowledge domain does,
// through CompleteFirstPack. See docs/contracts/identity-a2.md and docs/contracts/identity-a3.md.
package identity
