package main

var (
	// Version variable will be replaced at link time when building the artifact
	// with:
	//
	//   go build -ldflags "-s -w -X main.Version=${VERSION}"
	//
	// where `VERSION=$(cat VERSION)` VERSION being the file at the root of the
	// project repository.
	Version = "<NOT PROPERLY GENERATED>"

	// BuildDate variable will be replaced at link time when building the artifact
	// with:
	//
	//   go build -ldflags "-s -w -X main.BuildDate=${BUILD_DATE}"
	//
	// where `BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')` or
	// `BUILD_DATE=$(git show -s --format=%cI "$(git rev-parse HEAD)")` from CI for
	// reproducible builds.
	BuildDate = "<NOT PROPERLY GENERATED>"

	// GitHash variable will be replaced at link time when building the artifact
	// with:
	//
	//   go build -ldflags "-s -w -X main.GitHash=${GIT_HASH}"
	//
	// where `GIT_HASH=$(git rev-parse HEAD)` HEAD can be changed to the actual
	// commits used for the building.
	GitHash = "<NOT PROPERLY GENERATED>"

	// GitTreeState variable will be replaced at link time when building the artifact
	// with:
	//
	//   go build -ldflags "-s -w -X main.GitTreeState=${GIT_TREE_STATE}"
	//
	// where `GIT_TREE_STATE=$(if [ -z "$(git status --porcelain)" ]; then
	// echo "clean"; else echo "dirty"; fi)`.
	GitTreeState = "<NOT PROPERLY GENERATED>"
)