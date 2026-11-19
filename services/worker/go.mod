module github.com/zh7551069/qlutch-projects/services/worker

go 1.27.4

require (
	github.com/redis/go-redis/v9 v9.23.0
	github.com/zh7551069/qlutch-projects/services/airuntime v0.0.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	go.uber.org/atomic v1.12.0 // indirect
	golang.org/x/sys v0.48.0 // indirect
)

replace github.com/zh7551069/qlutch-projects/services/airuntime => ../airuntime
