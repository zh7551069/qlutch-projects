// Package airuntime is the only model entry for Qlutch services.
//
// Business packages call Call, RunS01, or RunS02. They do not call a provider
// SDK. RunS01 and RunS02 send one fragment or one node per Call. A textbook is
// not placed into one template. Prompts, source text, and prices are not stored.
// Ledger rows are append-only. An unconfigured provider returns 未配置提供方.
package airuntime
