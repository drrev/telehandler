## telehandler completion powershell

Generate the autocompletion script for powershell

### Synopsis

Generate the autocompletion script for powershell.

To load completions in your current shell session:

	telehandler completion powershell | Out-String | Invoke-Expression

To load completions for every new session, add the output of the above command
to your powershell profile.


```
telehandler completion powershell [flags]
```

### Options

```
  -h, --help              help for powershell
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
      --cgroup-root string   Path to cgroup v2 mount (default "/sys/fs/cgroup")
  -r, --root string          Root CA cert path (default "ssl/root.pem")
```

### SEE ALSO

* [telehandler completion](telehandler_completion.md)	 - Generate the autocompletion script for the specified shell

