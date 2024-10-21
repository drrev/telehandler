package config

// TODO: Add configuration.
// Since the CLI is implemented with github.com/spf13/cobra, use github.com/spf13/viper to bind config.

// Many defaults and critical CLI flags were already added, but there are quite a few configuration items that were left out for brevity.
// From cgroup constraints to namespaces, buffer sizes in many places (looking at you internal/foreman/foreman.go#WatchJobOutput),
// logging/levels/format, deadlines, timeout values, etc. would greatly benefit from configuration.
//
// To save time, configuration was omitted entirely where possible.
//
// In order to give an idea how I owuld normally split out configuration:
// Configuration is split by package or type depending on the desired granularity,
// such that each entrypoint into a type or package--for example `New...(s *Settings, ...)`--would
// take in a required param struct.
//
// To make this work at the high level, these configurations can either manually be added to a Config type,
// which is commonly seen, or the configurations can be registered to a top-level Config from their local `init()`.
