package watches

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/joyent/containerpilot/commands"
	"github.com/joyent/containerpilot/discovery"
	"github.com/joyent/containerpilot/utils"
)

// Config configures the watch
type Config struct {
	Name             string `mapstructure:"name"`
	serviceName      string
	Poll             int         `mapstructure:"poll"` // time in seconds
	Exec             interface{} `mapstructure:"onChange"`
	exec             *commands.Command
	Tag              string `mapstructure:"tag"`
	Timeout          string `mapstructure:"timeout"`
	timeout          time.Duration
	discoveryService discovery.Backend
}

// NewConfigs parses json config into a validated slice of Configs
func NewConfigs(raw []interface{}, disc discovery.Backend) ([]*Config, error) {
	var watches []*Config
	if raw == nil {
		return watches, nil
	}
	if err := utils.DecodeRaw(raw, &watches); err != nil {
		return watches, fmt.Errorf("Watch configuration error: %v", err)
	}
	for _, watch := range watches {
		if err := watch.Validate(disc); err != nil {
			return watches, err
		}
	}
	return watches, nil
}

// Validate ensures Config meets all requirements
func (cfg *Config) Validate(disc discovery.Backend) error {
	if err := utils.ValidateServiceName(cfg.Name); err != nil {
		return err
	}

	cfg.serviceName = cfg.Name
	cfg.Name = cfg.Name + ".watch"

	if cfg.Exec == nil {
		// TODO v3: this error message is tied to existing config syntax
		return fmt.Errorf("`onChange` is required in watch %s", cfg.serviceName)
	}
	if cfg.Timeout == "" {
		cfg.Timeout = fmt.Sprintf("%ds", cfg.Poll)
	}
	timeout, err := utils.GetTimeout(cfg.Timeout)
	if err != nil {
		return fmt.Errorf("could not parse `timeout` in watch %s: %v", cfg.serviceName, err)
	}
	cfg.timeout = timeout

	if cfg.Poll < 1 {
		return fmt.Errorf("`poll` must be > 0 in watch %s", cfg.serviceName)
	}
	cmd, err := commands.NewCommand(cfg.Exec, cfg.timeout,
		log.Fields{"watch": cfg.Name})
	if err != nil {
		// TODO v3: this error message is tied to existing config syntax
		return fmt.Errorf("could not parse `onChange` in watch %s: %s",
			cfg.serviceName, err)
	}
	cmd.Name = cfg.Name
	cfg.exec = cmd
	cfg.discoveryService = disc
	return nil
}

// String implements the stdlib fmt.Stringer interface for pretty-printing
func (cfg *Config) String() string {
	return "watches.Config[" + cfg.Name + "]"
}
