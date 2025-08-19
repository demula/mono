<!-- markdownlint-disable-next-line first-line-h1 no-inline-html -->
<div align="center"><img src="logo.png" alt="Monkeys venerates monolith"/></div>

<!-- markdownlint-disable-next-line no-inline-html -->
# <div align="center">Mono - Go monorepo missing tool</div>

Update all `go.mod` and `go.sum` files of a mono-repo to the given tagged release.
<!-- markdownlint-disable-next-line no-inline-html line-length -->
## <div align="center"> [![Go Report Card](https://goreportcard.com/badge/github.com/demula/mono)](https://goreportcard.com/report/github.com/demula/mono) [![codecov](https://codecov.io/gh/demula/mono/branch/main/graph/badge.svg)](https://codecov.io/gh/demula/mono) </div>

## Installation

### Pre-built binary

Grab your pre-built binary from the
[Releases](https://github.com/demula/mono/releases) page.

### From source

As a command in `PATH`:

```bash
go install "github.com/demula/mono@latest"
```

or as a tool in the project (not recommended, dependencies mixed with your
project):

```bash
go get -tool "github.com/demula/mono@latest"
```

## Usage

On the same directory you have `go.work`:

```bash
mono release --only-go-mod-sum "v0.1.0-alpha.1"
```

The flag `--only-go-mod-sum` is required as we want to leave room for this tool
to become a simplified version of
[Cocogitto](https://github.com/cocogitto/cocogitto) or
[goreleaser](https://github.com/goreleaser/goreleaser) without being
incompatible with existing users.

> [!IMPORTANT] Do not run this command while running any other go command. It
> does not lock the files used and it is meant to only be run after all
> modifications (go mod tidy and others) are done.

### Example

There is an example repository that you can use for testing the functionality
out: [mono-example](https://github.com/demula/mono-example)

```bash
git clone httts://github.com/demula/mono-example
cd mono-example
make init # Install mono globally
make example-release # Creates a new RC version that can be committed and tagged.
```

That repository shows how to use the `mono` command to prepare for a release.
`mono` at the moment does not commit or tag the repository for you.

> [!NOTE] Do not go too crazy creating tags and asking `go` to download them. The
> Go package repository does **NOT** delete anything (even if you repo is
> private).

## Pricing

If you use this project inside a successful company I do expect some
compensation. Yes, it is hard to go through approval workflows in a corporation
and yes, it is Apache 2 licensed but open source as we know it can only survive
if there is some kind of compensation for the people behind it. It is hard to
justify putting aside free time to do maintenance if there is little incentive.

## Why

It must be **me** but releasing a mono repo where the modules depend on each
other is not as straight forward as I hope.

Let's say you have the following 3 modules:

```txt
api/
  go.mod
  go.sum
core/
  go.mod
  go.sum
cli/
  go.mod
  go.sum
server/
  go.mod
  go.sum
go.work
go.work.sum
```

They are clearly connected to each other where you have the dependency chain:
`api<-core<-cli|server` .

The reasoning of having this 4 modules is not that you are a former Java
developer but:

- You're exposing a protobuf API that will also be consumed by other languages.
- You have core business logic shared by CLI and server

The problem you have is that you need to tag a commit to create a version of
each module. This forces the following dilemma, would you release each module
with a different commit/tag? or will you release all in the same commit?

This tool is for the later.

On this approach you need to carefully modify and update the `go.mod`s and the
`go.sum`s in the same commit so it can be tagged all at once.

> [!NOTE] The interdependencies should have already pre-aligned with
> `go work sync` to avoid surprises.

First step is to set `core` module dependency to `api` to the new version
`@x.y.z` in its `go.mod` this changes its hash so we need to recalculate it for
the current commit.

```txt
github.com/demula/mono/api v0.1.0-alpha.1 h1:kq5EqPL5j5EvsGSjvqvJkMYdXAbkc7IbKvVyGdKKH6Y=
github.com/demula/mono/api v0.1.0-alpha.1/go.mod h1:+8G1dboc9Dw+NZyv37tCp6UKmqXbzLitwYmVWvgPcg4=
```

Changing the `core` module's `go.mod` and `go.sum` will force us in turn to
change its signature/hashes in `cli` and `server`.

```txt
github.com/demula/mono/core v0.1.0-alpha.1 h1:NIvaJDMOsjHA8n1jAhLSgzrAzy1Hgr+hNrb57e+94F0=
github.com/demula/mono/core v0.1.0-alpha.1/go.mod h1:TIyPZe4MgqvfeYDBFedMoGGpEw/LqOeaOT+nhxU+yHo=
```

Notice that we do not need to change anything in `api` as it does not depend in
anything that we need to change when tagging.

## Why not just commit go.work

Personally I wished for more guidance from the go dev team on this topic when
they created the workspaces feature. That said, these are my reasons:

- Broken importing behaviour for external modules. Any module outside the
workspace using the modules in the workspace will not benefit from the
"local path resolution" that the `go.work` offers. This can lead to broken code
or unexpected behaviour. Debugging a quirk in this scenario must be terribly
fun where execution inside workspace outputs 'A' (and all tests are green) but
outside outputs 'B'.
- Convention. Other go devs expect the usage and behaviour marked in the
community guidelines when workspaces were released explicitly calling to
**NOT** commit the file in question.

## Shortcomings

While this command works great for creating a release, it does not help on the
everyday development. Keep in mind that when doing changes across different
modules you will need to commit the changes in order to still get a consistent
repository. Maybe in the future we can expand this command to help on this use
case where we can check if there are uncommitted changes or outdated
interdependencies when working with `go.work` as pre-commit check (and maybe a
`--fix` option when possible).

## Attributions

The file `gosum/gosum.go` is a modified version of the golang source code of
[`go/src/cmd/go/internal/modfetch`](https://cs.opensource.google/go/go/+/refs/tags/go1.25.0:src/cmd/go/internal/modfetch/fetch.go;drc=07a279794dff7ef3371710f1de4b3f9fc4ef4987)
created by The Go Authors.
