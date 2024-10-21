## telehandler

Telehandler is a simple service that is used to start, stop, query status, and watch the output of an arbitrary Linux process over gRPC.

### Options

```
      --cgroup-root string   Path to cgroup v2 mount (default "/sys/fs/cgroup")
  -h, --help                 help for telehandler
  -r, --root string          Root CA cert path (default "ssl/root.pem")
```

### SEE ALSO

* [telehandler client](telehandler_client.md)	 - client is used to run subcommands over gRPC
* [telehandler completion](telehandler_completion.md)	 - Generate the autocompletion script for the specified shell
* [telehandler local](telehandler_local.md)	 - Run the given command with cgroup and namespace enforcement.
* [telehandler server](telehandler_server.md)	 - Starts a gRPC server for running and managing jobs

