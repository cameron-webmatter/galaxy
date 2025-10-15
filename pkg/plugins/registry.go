package plugins

import (
	"github.com/galaxy/galaxy/pkg/config"
)

func NewDefaultManager(cfg *config.Config) *Manager {
	mgr := NewManager(cfg)
	return mgr
}
