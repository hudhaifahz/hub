package wails

import "github.com/getAlby/hub/guidedmigration"

type guidedImportSelectionResponse struct {
	Cancelled bool                           `json:"cancelled"`
	Summary   *guidedmigration.SourceSummary `json:"summary,omitempty"`
}

type guidedImportStageRequest struct {
	SourcePath string `json:"sourcePath"`
}

type guidedImportSourceRequest struct {
	SourcePath string `json:"sourcePath"`
}

type guidedImportVerifyRequest struct {
	UnlockPassword string `json:"unlockPassword"`
}

type guidedImportActivateRequest struct {
	UnlockPassword         string `json:"unlockPassword"`
	NodeIDConfirmation     string `json:"nodeIdConfirmation"`
	ActivationConfirmation string `json:"activationConfirmation"`
}
