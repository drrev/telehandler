# telehandler
A prototype job worker service that provides an API to run arbitrary Linux processes.

See the [design documentation](docs/design.md) for more information.

## CLI Usage

Markdown formatted documentation for all commands is available in [docs/cli/telehandler.md](docs/cli/telehandler.md).

At any time, `--help` or `help <command>` can be used to get information about any command.

Autocompletion can be generated using `./telehandler completion` if desired.

## Generating Certs

Install [cfssl](https://github.com/cloudflare/cfssl) and add to `$PATH`.

Generate certs with `make certs`.

All commands have defaults to point to the `ssl` directory already.

Two different client certs are included to validate CN-based auth. To change certs to the "bubba" cert, use `client <subcommand> -c ssl/bubba.pem -k ssl/bubba-key.pem [-- args...]`. See:  [client docs](./docs/cli/telehandler_client.md) for more information.

## Building

Telehandler is build using `make` or `make build`.

That will generate a `telehandler` binary for the current host OS+Arch. To build a Linux binary, use `GOOS=linux make`.

Telehandler **will not** compile for any target other than Linux.

