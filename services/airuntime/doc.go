// Package airuntime is the only model entry for Qlutch services.
//
// Business packages must not call a provider SDK themselves. Cache, tier
// routing, and ledger storage are intentionally absent; this package only
// holds the scene-code signature and reports that no provider is configured.
package airuntime
