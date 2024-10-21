# telehandler
A prototype job worker service that provides an API to run arbitrary Linux processes.

See the [design documentation](docs/design.md) for more information.

## CLI Usage

Markdown formatted documentation for all commands is available in [docs/cli/telehandler.md](docs/cli/telehandler.md).

At any time, `--help` or `help <command>` can be used to get information about any command.

Autocompletion can be generated using `./telehandler completion`, if desired.

### Run Without a Server

The [`local`](docs/cli/telehandler_local.md) command can be used to run a command without the need to set up a server.

When running the `local` command, `root` must be used for cgroup setup+teardown. If
the user has a non-root cgroup v2 setup properly, the path may be given at runtime.

Before proceeding, `telehandler` must be built:

```bash
user@host.internal>$ make
```

Example usage:

```bash
user@host.internal>$ sudo ./telehandler local -- bash -c 'for i in {1..500}; do echo $i; done'
```

See also: [docs/cli/telehandler_local.md](docs/cli/telehandler_local.md).

### Run with Client+Server

When running the `server` command, `root` must be used for cgroup setup+teardown. If
the user has a non-root cgroup v2 setup properly, the path may be given at runtime.

Before proceeding, `telehandler` must be built and certs must be generated:


```bash
user@host.internal>$ make certs build
```

#### Running a Server

Refer to: [docs/cli/telehandler_server.md](docs/cli/telehandler_server.md)

By default, the server runs in the foreground, so it is best to run it in a **dedicated** terminal:

```bash
user@host.internal>$ sudo ./telehandler server
2024/10/14 20:30:19 INFO Listening addr=:6443
```

#### Client Commands

Refer to: [docs/cli/telehandler_client.md](docs/cli/telehandler_client.md)

When a `client` command is run, a Job ID file is created for convenience. The path for this file can be changed with the `-j` flag.
By default, the path is `./job_id`.

Jobs are created for running commands using the [`client run`](docs/cli/telehandler_client_run.md) command.

In a **separate** terminal:
```bash
user@host.internal>$ ./telehandler client run -- bash -c 'for i in {1..500}; do echo $i; done'
1
2
3
...
500
user@host.internal>$ ./telehandler client run -- cat ./cmd/run.go
package cmd
...
```

To run as a different user, specify the path to a different cert and key:
```bash
# note, adding sleep 1 in here so that the job runs longer
user@host.internal>$ ./telehandler client run \
    -c ssl/bubba.pem \
    -k ssl/bubba-key.pem -- \
    -j bubba_job_id \
    bash -c 'for i in {1..100}; do echo $i; sleep 1; done'
```

To watch output from any commands--or to retrieve historical logs, run the [`client watch`](docs/cli/telehandler_client_watch.md) command:
```bash
user@host.internal>$ ./telehandler client watch $(cat job_id) # or $(cat bubba_job_id) to use the ID above
```

Finally, jobs can be interrupted at any point using [`client stop`](docs/cli/telehandler_client_stop.md):
```bash
user@host.internal>$ ./telehandler client stop $(cat job_id)
```

#### Benchmark

Refer to: [docs/cli/telehandler_client_benchmark.md](docs/cli/telehandler_client_benchmark.md)

A rudimentary e2e benchmark is included under `./telehandler client benchmark <command> [args...]`. This will start a new job to run the given command and spread it across `100` watchers (customizable with `--watchers`).
All output is saved to files named `out-*`. All data should be the same across all watchers and can be verified by hashing the files.

The rate per seconds is printed for each watcher every second with a final average at the end.

> Using this simple benchmark, I sent the ArchLinux ISO across 100 workers with a cumulative avg rate of 1GiB/s on a AMD 5950x. All files hashed equal to the source.
> 
> The image was sent into STDOUT with `cat`:   
> `./telehandler client benchmark cat data/archlinux-2024.10.01-x86_64.iso`

## Generating Certs

Install [cfssl](https://github.com/cloudflare/cfssl) and add to `$PATH`. For example: `go install github.com/cloudflare/cfssl/cmd/...@latest`

Once `cfssl` is installed, generate certs with `make certs`.

All commands have defaults to point to the `ssl` directory already.

Two different client certs are included to validate CN-based auth. To change certs to the "bubba" cert, use `client <subcommand> -c ./ssl/bubba.pem -k ./ssl/bubba-key.pem [-- args...]`. 

Refer to the [client docs](./docs/cli/telehandler_client.md) for more information.

## Building

Telehandler is build using `make` or `make build`.

That will generate a `telehandler` binary for the current host OS+Arch. To build a Linux binary, use `GOOS=linux make`.

Telehandler **will not** compile for any target other than Linux.

