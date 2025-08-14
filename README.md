# mono

Update all go.mod and go.sum files of a mono-repo to the given tagged release.

## Usage

On the same directory you have `go.work`:

```bash
mono release "v0.1.0-alpha.1"
```

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

> ![NOTE]: The interdependencies should have already pre-aligned with
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
they create the workspaces feature. That said, these are my reasons:

- Broken importing behaviour for external modules. Any module outside the
workspace using the modules in the workspace will not benefit from the
"local path resolution" that the `go.work` offers. This can lead to broken code
or unexpected behaviour. Debugging a quirk in this scenario must be terribly
fun where execution inside workspace outputs 'A' (and all tests are green) but
outside outputs 'B'.
- Convention. Other go devs expect the usage and behaviour marked in the
community guidelines when workspaces were released explicitly calling to
**NOT** commit the file in question.
