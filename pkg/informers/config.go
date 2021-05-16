package informers

import (
	"encoding/json"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"math/rand"
	"os"
	"path"
	"strings"
	"time"
)

type ResyncPeriodFunc func() time.Duration

// ChannelConfig defines a channel to receive notifications
type ChannelConfig struct {
	// Type is the type of the channel
	Type ChannelType `json:"type" yaml:"type"`

	Callback *ChannelCallbackConfig `json:"callback" yaml:"callback"`
	Group    *ChannelGroupConfig    `json:"group" yaml:"group"`
	Print    *ChannelPrintConfig    `json:"print" yaml:"print"`
	Telegram *ChannelTelegramConfig `json:"telegram" yaml:"telegram"`
}

type ResourceConfig struct {
	// Resource is the resource to watch, e.g. "deployments.v1.apps"
	Resource string `json:"resource" yaml:"resource"`

	// NoticeWhenAdded determine whether to notice when a resource is added
	NoticeWhenAdded bool `json:"noticeWhenAdded" yaml:"noticeWhenAdded"`

	// NoticeWhenDeleted determine whether to notice when a resource is deleted
	NoticeWhenDeleted bool `json:"noticeWhenDeleted" yaml:"noticeWhenDeleted"`

	// NoticeWhenUpdated determine whether to notice when a resource is updated,
	// When UpdateOn is not nil, only notice when fields in UpdateOn is changed
	NoticeWhenUpdated bool `json:"noticeWhenUpdated" yaml:"noticeWhenUpdated"`

	// UpdateOn defines fields to watch, used with NoticeWhenUpdated
	UpdateOn []string `json:"updateOn" yaml:"updateOn"`

	// ChannelNames defines channels to send notification
	ChannelNames []ChannelName `json:"channelNames" yaml:"channelNames"`

	// ResyncPeriod is the resync period in reflectors for this resource
	ResyncPeriod string `json:"resyncPeriod" yaml:"resyncPeriod"`

	// Workers is the number of workers
	Workers int `json:"workers" yaml:"workers"`
}

func (c *ResourceConfig) BuildResyncPeriodFuncWithDefault(defaultFunc ResyncPeriodFunc) (ResyncPeriodFunc, error) {
	f, set, err := c.BuildResyncPeriodFunc()
	if err != nil {
		return nil, err
	}
	if !set {
		return defaultFunc, nil
	}
	return f, nil
}

func (c *ResourceConfig) BuildResyncPeriodFunc() (f func() time.Duration, set bool, err error) {
	return BuildResyncPeriodFunc(c.ResyncPeriod)
}

type NamespaceConfig struct {
	// Namespace is the namespace to watch, default to "", which means all namespaces
	Namespace string `json:"namespace" yaml:"namespace"`

	// Resources is the resources you want to watch
	Resources []ResourceConfig `json:"resources" yaml:"resources"`

	// DefaultChannelNames defines default channels in this namespace
	DefaultChannelNames []ChannelName `json:"defaultChannelNames" yaml:"defaultChannelNames"`

	// DefaultWorkers is the default number of workers in this namespace
	DefaultWorkers int `json:"defaultWorkers" yaml:"defaultWorkers"`

	// MinResyncPeriod is the resync period in reflectors in this namespace;
	// actual resync period will be random between MinResyncPeriod and 2*minResyncPeriod.
	MinResyncPeriod string `json:"minResyncPeriod" yaml:"minResyncPeriod"`
}

func (c *NamespaceConfig) BuildResyncPeriodFuncWithDefault(defaultFunc ResyncPeriodFunc) (ResyncPeriodFunc, error) {
	f, set, err := c.BuildResyncPeriodFunc()
	if err != nil {
		return nil, err
	}
	if !set {
		return defaultFunc, nil
	}
	return f, nil
}

func (c *NamespaceConfig) BuildResyncPeriodFunc() (f ResyncPeriodFunc, set bool, err error) {
	return BuildResyncPeriodFunc(c.MinResyncPeriod)
}

type Config struct {
	// Namespaces defines namespaces and resources to watch
	Namespaces []NamespaceConfig `json:"namespaces" yaml:"namespaces"`

	// Channels defines channels that receive notifications
	Channels map[ChannelName]ChannelConfig `json:"channels" yaml:"channels"`

	// DefaultChannelNames defines default channels
	DefaultChannelNames []ChannelName `json:"defaultChannelNames" yaml:"defaultChannelNames"`

	// DefaultWorkers is the default number of workers
	DefaultWorkers int `json:"defaultWorkers" yaml:"defaultWorkers"`

	// MinResyncPeriod is the resync period in reflectors;
	// actual resync period will be random between MinResyncPeriod and 2*minResyncPeriod.
	MinResyncPeriod string `json:"minResyncPeriod" yaml:"minResyncPeriod"`
}

func (c *Config) BuildResyncPeriodFunc() (f ResyncPeriodFunc, err error) {
	if c.MinResyncPeriod == "" {
		c.MinResyncPeriod = "12h"
	}
	f, _, err = BuildResyncPeriodFunc(c.MinResyncPeriod)
	return f, err
}

func BuildResyncPeriodFunc(resyncPeriod string) (f ResyncPeriodFunc, set bool, err error) {
	duration, set, err := ParseResyncPeriod(resyncPeriod)
	if err != nil {
		return nil, false, err
	}
	if !set {
		return nil, false, nil
	}
	durationFloat := float64(duration.Nanoseconds())
	// generate time.Duration between duration and 2*duration
	return func() time.Duration {
		factor := rand.Float64() + 1
		return time.Duration(durationFloat * factor)
	}, true, nil
}

func ParseResyncPeriod(resyncPeriod string) (f time.Duration, set bool, err error) {
	if resyncPeriod == "" {
		return 0, false, nil
	}
	duration, err := time.ParseDuration(resyncPeriod)
	if err != nil {
		return 0, false, errors.Wrap(err, "time.ParseDuration error")
	}
	return duration, true, nil
}

// LoadConfigFromFile loads config from file
func LoadConfigFromFile(file string) (*Config, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, errors.Wrap(err, "os.Open error")
	}
	var config Config
	switch t := strings.ToLower(path.Ext(file)); t {
	case ".json":
		err = json.NewDecoder(f).Decode(&config)
		if err != nil {
			return nil, errors.Wrap(err, "json decode error")
		}
	case ".yaml":
		err = yaml.NewDecoder(f).Decode(&config)
		if err != nil {
			return nil, errors.Wrap(err, "yaml decode error")
		}
	default:
		return nil, errors.Errorf("unsupported file type: %s", t)
	}
	return &config, nil
}
