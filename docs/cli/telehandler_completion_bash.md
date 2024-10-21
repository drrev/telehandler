## telehandler completion bash

Generate the autocompletion script for bash

### Synopsis

Generate the autocompletion script for the bash shell.

This script depends on the 'bash-completion' package.
If it is not installed already, you can install it via your OS's package manager.

To load completions in your current shell session:

	source <(telehandler completion bash)

To load completions for every new session, execute once:

#### Linux:

	telehandler completion bash > /etc/bash_completion.d/telehandler

#### macOS:

	telehandler completion bash > $(brew --prefix)/etc/bash_completion.d/telehandler

You will need to start a new shell for this setup to take effect.


```
telehandler completion bash
```

### Options

```
  -h, --help              help for bash
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
      --cgroup-root string   Path to cgroup v2 mount (default "/sys/fs/cgroup")
  -r, --root string          Root CA cert path (default "ssl/root.pem")
```

### SEE ALSO

* [telehandler completion](telehandler_completion.md)	 - Generate the autocompletion script for the specified shell

