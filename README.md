# Terraform Provider pfSense

[![Registry](https://img.shields.io/badge/pfsense-Terraform%20Registry-blue)](https://registry.terraform.io/providers/marshallford/pfsense/latest/docs)

Used to configure [pfSense](https://www.pfsense.org/) firewall/router systems with Terraform. Validated with pfSense CE, compatibility with pfSense Plus is not guaranteed.

> [!WARNING]
> All versions released prior to `v1.0.0` are to be considered [breaking changes](https://semver.org/#how-do-i-know-when-to-release-100).

## Support Matrix

| Release  | pfSense            | Terraform      |
| :------: | :----------------: | :------------: |
| < v1.0.0 | >= 2.6.0, <= 2.7.2 | >= 1.6, <= 1.8 |

## Development Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads)
- [Go](https://golang.org/doc/install)

## Building The Provider

1. Clone the repository
2. Enter the repository directory
3. Build the provider using the Go `install` command:

```shell
go install
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#development-requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `make docs`.

In order to run the full suite of Acceptance tests, run `make test/acc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```shell
make test/acc
```
