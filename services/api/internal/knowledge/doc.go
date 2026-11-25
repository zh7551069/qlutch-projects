// Package knowledge implements learner knowledge packs, presigned uploads,
// import acceptance, job queries, and first-pack skip.
//
// Object bytes stay in an S3-compatible store behind Blobs. This package does
// not name a cloud vendor and does not import a vendor SDK. Import acceptance
// stores status accepted and enqueues a job id. It does not mark a job
// map_ready. A later worker owns that transition. Tests use TestingSetJobStatus.
package knowledge
