package dwl

import (
	"dwl/progress"
	"dwl/settings"
)

type Loader interface {
	Connect(*settings.Settings) error
	Load()
	Stop()
	Complete() bool
	GetProgress() progress.Progress

	GetError() error
}
