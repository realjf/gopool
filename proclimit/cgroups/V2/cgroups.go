package V2

import (
	"fmt"
	"os/exec"

	cgroupsv2 "github.com/containerd/cgroups/v2"
	"github.com/pkg/errors"

	"github.com/realjf/gopool/proclimit"
)

// Option allows for customizing the behaviour of the Cgroup limiter.
//
// Options will typically perform operations on cgroup.LinuxResources.
type Option func(cgroup *Cgroup2)

// WithName sets the name of the Cgroup. If not specified, a random UUID will
// be generated as the name
func WithName(name string) Option {
	return func(cgroup *Cgroup2) {
		cgroup.Name = name
	}
}

// WithCPULimit sets the maximum CPU limit (as a percentage) allowed for all processes within the Cgroup.
// The percentage is based on a single CPU core. That is to say, 50 allows for the use of half of a core,
// 200 allows for the use of two cores, etc.
//
// `cpu.cfs_period_us` will be set to 100000 (100ms) unless it has been overridden (e.g. by an Option that
// is added before this Option. Note that no such Option has been implemented currently).
//
// `cpu.cfs_quota_us` will be set to cpuLimit percent of `cpu.cfs_period_us`.
func WithCPULimit(cpuLimit proclimit.Percent) Option {
	return func(cgroup *Cgroup2) {
		if cgroup.LinuxResources.CPU == nil {
			cgroup.LinuxResources.CPU = &cgroupsv2.CPU{}
		}
		var period *uint64 = new(uint64)
		*period = 100000

		var quota *int64 = new(int64)
		*quota = int64(*period * uint64(cpuLimit) / 100)
		cgroup.LinuxResources.CPU.Max = cgroupsv2.NewCPUMax(quota, period)
	}
}

// WithMemoryLimit sets the maximum amount of virtual memory allowed for all processes within the Cgroup.
//
// `memory.max_usage_in_bytes` is set to memory
func WithMemoryLimit(memory proclimit.Memory) Option {
	return func(cgroup *Cgroup2) {
		if cgroup.LinuxResources.Memory == nil {
			cgroup.LinuxResources.Memory = &cgroupsv2.Memory{}
		}

		var limit *int64 = new(int64)
		*limit = int64(memory)
		cgroup.LinuxResources.Memory.Max = limit
	}
}

// Cgroup represents a cgroup in a Linux system. Resource limits can be
// configured by modifying LinuxResources through Options. Modifying
// LinuxResources after calling New(...) will have no effect.
type Cgroup2 struct {
	Name           string
	LinuxResources *cgroupsv2.Resources
	cgroup         *cgroupsv2.Manager
}

// New creates a new Cgroup. Resource limits and the name of the Cgroup can be defined
// using Option arguments.
func New(options ...Option) (*Cgroup2, error) {
	c := &Cgroup2{
		LinuxResources: &cgroupsv2.Resources{},
	}
	for _, opt := range options {
		opt(c)
	}
	var err error
	if c.Name == "" {
		c.Name, err = proclimit.RandomName()
		if err != nil {
			return nil, err
		}
	}

	c.cgroup, err = cgroupsv2.NewSystemd("", fmt.Sprintf("/%s", c.Name), -1, c.LinuxResources)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cgroup")
	}
	return c, nil
}

// New loads an existing Cgroup by name
func Existing(name string) (*Cgroup2, error) {
	c := &Cgroup2{
		Name: name,
	}
	var err error
	c.cgroup, err = cgroupsv2.LoadSystemd("", fmt.Sprintf("/%s", name))
	if err != nil {
		return nil, errors.Wrap(err, "failed to load cgroup")
	}
	return c, nil
}

// Command constructs a wrapped Cmd struct to execute the named program with the given arguments.
// This wrapped Cmd will be added to the Cgroup when it is started.
func (c *Cgroup2) Command(name string, arg ...string) *proclimit.Cmd {
	return c.Wrap(exec.Command(name, arg...))
}

// Wrap takes an existing exec.Cmd and converts it into a proclimit.Cmd.
// When the returned Cmd is started, it will have the resources applied.
//
// If cmd has already been started, Wrap will panic. To limit the resources
// of a running process, use Limit instead.
func (c *Cgroup2) Wrap(cmd *exec.Cmd) *proclimit.Cmd {
	if cmd.Process != nil {
		panic("cmd has already been started")
	}
	return &proclimit.Cmd{
		Cmd:     cmd,
		Limiter: c,
	}
}

// Limit applies Cgroup resource limits to a running process by its pid.
func (c *Cgroup2) Limit(pid int) error {
	return c.cgroup.AddProc(uint64(pid))
}

// Close deletes the Cgroup definition from the filesystem.
func (c *Cgroup2) Close() error {
	return c.cgroup.Delete()
}
