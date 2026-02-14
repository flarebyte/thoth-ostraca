# Contributing

TODO: update

Welcome! And many thanks for taking the time to contribute!

Before getting started, please read the [Technical Design
documentation](TECHNICAL_DESIGN.md) to understand the architectural and
coding guidelines used in this project.

There are many ways you can contribute: fixing bugs, writing documentation,
improving tests, suggesting features, or reviewing code.

Please note we have a [Code of Conduct](CODE_OF_CONDUCT.md); please follow it
in all your interactions with the project.

## Build the project locally

Make sure you have Go installed (preferably the latest stable version). Clone
the repository and build from the root:

```bash
git clone https://github.com/flarebyte/clingy-code-detective.git
cd clingy-code-detective
go build ./...
```

The following commands should get you started:

Setup an alias:

```bash
alias broth='npx baldrick-broth'
```

or if you prefer to always use the latest version:

```bash
alias broth='npx baldrick-broth@latest'
```

Run the unit tests:

```bash
broth test unit
```

A list of [most used commands](MAINTENANCE.md) is available:

```bash
broth
```

Please keep an eye on test coverage, bundle size and documentation.
When you are ready for a pull request:

```bash
broth release ready
```

You can also simulate [Github actions](https://docs.github.com/en/actions)
locally with [act](https://github.com/nektos/act).
You will need to setup `.actrc` with the node.js docker image `-P
ubuntu-latest=node:16-buster`

To run the pipeline:

```bash
broth github act
```

## Pull Request Process

1.  Make sure that an issue describing the intended code change exists and
    that this issue has been accepted.

When you are about to do a pull-request:

```bash
broth release ready -pr
```

Then you can create the pull-request:

```bash
broth release pr
```

## Publishing the library

This would be done by the main maintainers of the project. Locally for now as
updates are pretty infrequent, and some of tests have to be done manually.

```bash
broth release publish
```
