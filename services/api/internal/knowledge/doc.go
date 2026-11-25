// Package knowledge implements learner knowledge packs, presigned uploads,
// import acceptance, job queries, and first-pack skip.
//
// Object bytes stay in an S3-compatible store behind Blobs. This package does
// not name a cloud vendor and does not import a vendor SDK. Import acceptance
// stores status accepted and enqueues a job id. It does not mark a job
// map_ready. ProcessImport runs S01 through airuntime, writes an enterable map,
// then sets map_ready. Tests may still call TestingSetJobStatus. That fixture
// is not an HTTP route and is not the production path.
package knowledge
