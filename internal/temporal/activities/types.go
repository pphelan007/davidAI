// Package activities contains Temporal activity definitions and client.
package activities

// AssetInfo represents information about an audio asset
type AssetInfo struct {
	AssetID     string        `json:"asset_id"`
	FilePath    string        `json:"file_path"`
	ContentHash string        `json:"content_hash"`
	Metadata    AudioMetadata `json:"metadata"`
}

// AudioMetadata contains basic audio file metadata
type AudioMetadata struct {
	SampleRate int     `json:"sample_rate"` // samples per second
	Duration   float64 `json:"duration"`    // duration in seconds
	Channels   int     `json:"channels"`    // number of audio channels
}

// IngestRawAudioInput is the input for the IngestRawAudio activity
type IngestRawAudioInput struct {
	FilePath string `json:"file_path"`
}

// IngestRawAudioOutput is the output from the IngestRawAudio activity
type IngestRawAudioOutput struct {
	Asset AssetInfo `json:"asset"`
}

// TrimSilenceInput is the input for the TrimSilence activity
type TrimSilenceInput struct {
	AssetID            string  `json:"asset_id"`
	SourcePath         string  `json:"source_path"`
	SilenceThreshold   float64 `json:"silence_threshold"`    // threshold for silence detection (0.0-1.0)
	MinSilenceDuration float64 `json:"min_silence_duration"` // minimum silence duration in seconds to trim
}

// TrimSilenceOutput is the output from the TrimSilence activity
type TrimSilenceOutput struct {
	NewAssetID  string `json:"new_asset_id,omitempty"` // empty if no new asset was created
	ContentHash string `json:"content_hash"`
	WasTrimmed  bool   `json:"was_trimmed"`           // true if silence was actually trimmed
	NoOp        bool   `json:"no_op"`                 // true if trimmed audio is identical to original
	OutputPath  string `json:"output_path,omitempty"` // path to trimmed audio file if created
}
