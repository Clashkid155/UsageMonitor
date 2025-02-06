package main

import (
	"github.com/charmbracelet/bubbles/table"
	usageTracker "github.com/clashkid155/usage-monitor"
)

type (
	AppModel struct {
		table          table.Model
		currentNetwork *currentNetwork
	}
	/*tableModel struct {
		table table.Model
	}*/

	currentNetwork struct {
		usages        []usageTracker.Usage
		totalUpload   uint64
		totalDownload uint64
		totalUsage    uint64
	}
)
