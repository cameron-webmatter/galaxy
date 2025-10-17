package plugins

import (
	"github.com/cameron-webmatter/galaxy/pkg/config"
)

func NewDefaultManager(cfg *config.Config) *Manager {
	mgr := NewManager(cfg)
	return mgr
}
