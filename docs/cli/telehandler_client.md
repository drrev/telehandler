## telehandler client

client is used to run subcommands over gRPC

### Options

```
  -c, --cert string     Client cert path (default "ssl/client.pem")
  -h, --help            help for client
  -k, --key string      Client key path (default "ssl/client-key.pem")
  -s, --server string   Address of a Telehandler server (default "localhost:6443")
```

### Options inherited from parent commands

```
      --cgroup-root string   Path to cgroup v2 mount (default "/sys/fs/cgroup")
  -r, --root string          Root CA cert path (default "ssl/root.pem")
```

### SEE ALSO

* [telehandler](telehandler.md)	 - Telehandler is a simple service that is used to start, stop, query status, and watch the output of an arbitrary Linux process over gRPC.
* [telehandler client benchmark](telehandler_client_benchmark.md)	 - A brief description of your command
* [telehandler client run](telehandler_client_run.md)	 - Run a Linux command using a Telehandler server
* [telehandler client stop](telehandler_client_stop.md)	 - Attempts to stop the given job
* [telehandler client watch](telehandler_client_watch.md)	 - Watch the output of a job

