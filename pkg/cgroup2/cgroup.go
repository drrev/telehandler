package cgroup2

var requiredControllers = []string{"cpu", "memory", "io"}

func Create(basePath string) (err error) {
	// TODO: Implement
	return nil
}

func Add(basePath string, pid int) error {
	// TODO: Implement
	return nil
}

func Cleanup(basePath string) error {
	// TODO: Implement
	return nil
}
