## telehandler completion zsh

Generate the autocompletion script for zsh

### Synopsis

Generate the autocompletion script for the zsh shell.

If shell completion is not already enabled in your environment you will need
to enable it.  You can execute the following once:

	echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions in your current shell session:

	source <(telehandler completion zsh)

To load completions for every new session, execute once:

#### Linux:

	telehandler completion zsh > "${fpath[1]}/_telehandler"

#### macOS:

	telehandler completion zsh > $(brew --prefix)/share/zsh/site-functions/_telehandler

You will need to start a new shell for this setup to take effect.


```
telehandler completion zsh [flags]
```

### Options

```
  -h, --help              help for zsh
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
      --cgroup-root string   Path to cgroup v2 mount (default "/sys/fs/cgroup")
  -r, --root string          Root CA cert path (default "ssl/root.pem")
```

### SEE ALSO

* [telehandler completion](telehandler_completion.md)	 - Generate the autocompletion script for the specified shell

