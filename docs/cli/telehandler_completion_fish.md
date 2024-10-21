## telehandler completion fish

Generate the autocompletion script for fish

### Synopsis

Generate the autocompletion script for the fish shell.

To load completions in your current shell session:

	telehandler completion fish | source

To load completions for every new session, execute once:

	telehandler completion fish > ~/.config/fish/completions/telehandler.fish

You will need to start a new shell for this setup to take effect.


```
telehandler completion fish [flags]
```

### Options

```
  -h, --help              help for fish
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
      --cgroup-root string   Path to cgroup v2 mount (default "/sys/fs/cgroup")
  -r, --root string          Root CA cert path (default "ssl/root.pem")
```

### SEE ALSO

* [telehandler completion](telehandler_completion.md)	 - Generate the autocompletion script for the specified shell

