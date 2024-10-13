## telehandler client run

Run a Linux command using a Telehandler server

### Synopsis

Run a Linux command using a Telehandler server.
	
Commands that include args should signal the end of args for this command with --
For example: run -- bash -c echo hello"

```
telehandler client run <command> [args...] [flags]
```

### Options

```
  -h, --help             help for run
  -j, --jidfile string   A file to write the ID of the Job. (default "job_id")
```

### Options inherited from parent commands

```
  -c, --cert string          Client cert path (default "ssl/client.pem")
      --cgroup-root string   Path to cgroup v2 mount (default "/sys/fs/cgroup")
  -k, --key string           Client key path (default "ssl/client-key.pem")
  -r, --root string          Root CA cert path (default "ssl/root.pem")
  -s, --server string        Address of a Telehandler server (default "localhost:6443")
```

### SEE ALSO

* [telehandler client](telehandler_client.md)	 - client is used to run subcommands over gRPC

