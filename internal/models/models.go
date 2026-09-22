package models

// ProcessRequest is the frozen contract between the door and the engine.
type ProcessRequest struct {
	Payload   string `json:"payload"`
	PayloadID string `json:"payload_id"`
}

// ProcessResponse is the frozen contract result.
type ProcessResponse struct {
	Result string `json:"result"`
}

// Span is a detected fragment of text containing PII.
type Span struct {
	Start      int     `json:"start"`
	End        int     `json:"end"`
	Type       string  `json:"type"`
	Confidence float64 `json:"confidence"`
	Source     string  `json:"source"`
}

// CorrRecord is the original<->mask correspondence stored for demasking.
type CorrRecord struct {
	Original  string   `json:"original"`
	Mask      string   `json:"mask"`
	Types     []string `json:"types"`
	CreatedTS float64  `json:"created_ts"`
}

// SystemConfig is the per-system masking policy.
type SystemConfig struct {
	SystemID      string `json:"system_id"`
	Enabled       bool   `json:"enabled"`
	Types         string `json:"types"`
	MaskMode      string `json:"mask_mode"`
	DemaskEnabled bool   `json:"demask_enabled"`
}
