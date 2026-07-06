// Published as github.com/pulsight-xyz/pulsight-go (mirrored from elio's
// sdks/go by the sdk-* jobs in .gitlab-ci.yml). The handwritten ergonomic
// layer is stdlib-only; `go mod tidy` adds the oapi-codegen runtime deps
// after the first `make sdk-go` generates the client core.
module github.com/pulsight-xyz/pulsight-go

go 1.25

require github.com/oapi-codegen/runtime v1.4.2

require (
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
)
