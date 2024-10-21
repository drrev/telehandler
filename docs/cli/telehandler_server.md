## telehandler server

Starts a gRPC server for running and managing jobs

```
telehandler server [flags]
```

### Options

```
  -c, --cert string       Server cert path (default "ssl/server.pem")
  -h, --help              help for server
  -k, --key string        Server key path (default "ssl/server-key.pem")
  -l, --listen string     ip:port to listen on for incoming connections (default ":6443")
  -p, --protocol string   protocol for incoming connections (default "tcp")
```

### Options inherited from parent commands

```
      --cgroup-root string   Path to cgroup v2 mount (default "/sys/fs/cgroup")
  -r, --root string          Root CA cert path (default "ssl/root.pem")
```

### SEE ALSO

* [telehandler](telehandler.md)	 - Telehandler is a simple service that is used to start, stop, query status, and watch the output of an arbitrary Linux process over gRPC.

