## telehandler client watch

Watch the output of a job

### Synopsis

Watch the output of a job starting at the beginning of execution through process termination.
	
Jobs do not need to be running to watch output.
If the job is not running, all historical output from process start to finish is retrieved.

```
telehandler client watch <job_id> [flags]
```

### Options

```
  -h, --help   help for watch
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

