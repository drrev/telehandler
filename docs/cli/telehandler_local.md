## telehandler local

Run the given command with cgroup and namespace enforcement.

### Synopsis

Run the given command with cgroup and namespace enforcement.

The local command does not require a running Telehandler server, all commands
are executed in the local environment.


```
telehandler local <cmd> [args...] [flags]
```

### Options

```
  -h, --help   help for local
```

### Options inherited from parent commands

```
      --cgroup-root string   Path to cgroup v2 mount (default "/sys/fs/cgroup")
  -r, --root string          Root CA cert path (default "ssl/root.pem")
```

### SEE ALSO

* [telehandler](telehandler.md)	 - Telehandler is a simple service that is used to start, stop, query status, and watch the output of an arbitrary Linux process over gRPC.

