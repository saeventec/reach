module github.com/saeventec/reach/services/hello

go 1.27.3

require (
	connectrpc.com/connect/v2 v2.0.2
	github.com/jackc/pgx/v5 v5.11.0
	github.com/saeventec/reach/packages/go v0.0.0
	google.golang.org/protobuf v1.36.12
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/text v0.29.0 // indirect
)

// The shared Go code lives in this repo, not on the internet. "replace"
// tells Go to use the folder instead of downloading the module.
// Learn more: docs/learn/go.md#sharing-code-between-modules
replace github.com/saeventec/reach/packages/go => ../../packages/go
