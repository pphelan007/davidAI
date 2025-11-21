package activities

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/google/uuid"
)

// IngestRawAudio ingests a raw audio file, registers it as an asset,
// computes its content hash, and extracts basic metadata.
func (ac *ActivitiesClient) IngestRawAudio(ctx context.Context, input IngestRawAudioInput) (*IngestRawAudioOutput, error) {
	// Read the audio file
	file, err := os.Open(input.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open audio file: %w", err)
	}
	defer file.Close()

	// Compute content hash
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return nil, fmt.Errorf("failed to compute hash: %w", err)
	}
	contentHash := hex.EncodeToString(hash.Sum(nil))

	// Reset file pointer for reading metadata
	if _, err := file.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("failed to seek file: %w", err)
	}

	// Decode WAV file to extract metadata
	decoder := wav.NewDecoder(file)
	if !decoder.IsValidFile() {
		return nil, fmt.Errorf("file is not a valid WAV file")
	}

	format := decoder.Format()
	sampleRate := int(format.SampleRate)
	channels := int(format.NumChannels)

	// Read all samples to calculate duration
	buf := &audio.IntBuffer{
		Format: format,
		Data:   make([]int, 0),
	}

	_, err = decoder.PCMBuffer(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode audio for metadata: %w", err)
	}

	// Calculate duration
	var duration float64
	if sampleRate > 0 && len(buf.Data) > 0 {
		// Duration = (number of samples) / (sample rate * channels)
		duration = float64(len(buf.Data)) / float64(sampleRate*channels)
	}

	// Generate asset ID
	assetID := uuid.New().String()

	// Create asset info
	asset := AssetInfo{
		AssetID:     assetID,
		FilePath:    input.FilePath,
		ContentHash: contentHash,
		Metadata: AudioMetadata{
			SampleRate: sampleRate,
			Duration:   duration,
			Channels:   channels,
		},
	}

	// TODO: Store asset in database/storage system
	// For now, we just return the asset info

	return &IngestRawAudioOutput{
		Asset: asset,
	}, nil
}

// TrimSilence trims silence from the beginning and end of an audio file,
// computes the content hash of the trimmed audio, and stores it as a new asset
// if it differs from the original.
func (ac *ActivitiesClient) TrimSilence(ctx context.Context, input TrimSilenceInput) (*TrimSilenceOutput, error) {
	// Default values if not provided
	silenceThreshold := input.SilenceThreshold
	if silenceThreshold == 0 {
		silenceThreshold = 0.01 // Default 1% threshold
	}
	minSilenceDuration := input.MinSilenceDuration
	if minSilenceDuration == 0 {
		minSilenceDuration = 0.1 // Default 100ms minimum silence
	}

	// Open and decode the source audio file
	file, err := os.Open(input.SourcePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open source audio file: %w", err)
	}
	defer file.Close()

	decoder := wav.NewDecoder(file)
	if !decoder.IsValidFile() {
		return nil, fmt.Errorf("file is not a valid WAV file")
	}

	format := decoder.Format()
	sampleRate := int(format.SampleRate)
	channels := int(format.NumChannels)

	// Read all audio samples
	var samples []int
	buf := &audio.IntBuffer{
		Format: format,
		Data:   make([]int, 0),
	}

	_, err = decoder.PCMBuffer(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode audio: %w", err)
	}
	samples = buf.Data

	// Find start and end of non-silent audio
	startIdx, endIdx := findNonSilentRange(samples, channels, silenceThreshold, sampleRate, minSilenceDuration)

	// Check if trimming is needed
	if startIdx == 0 && endIdx == len(samples) {
		// No trimming needed - audio has no leading/trailing silence
		// Compute hash of original file
		hash := sha256.New()
		if _, err := file.Seek(0, 0); err != nil {
			return nil, fmt.Errorf("failed to seek file: %w", err)
		}
		if _, err := io.Copy(hash, file); err != nil {
			return nil, fmt.Errorf("failed to compute hash: %w", err)
		}
		contentHash := hex.EncodeToString(hash.Sum(nil))

		return &TrimSilenceOutput{
			ContentHash: contentHash,
			WasTrimmed:  false,
			NoOp:        true,
		}, nil
	}

	// Trim the samples
	trimmedSamples := samples[startIdx:endIdx]

	// Create output file path
	outputDir := filepath.Dir(input.SourcePath)
	outputPath := filepath.Join(outputDir, fmt.Sprintf("trimmed_%s_%s.wav", input.AssetID, time.Now().Format("20060102_150405")))

	// Write trimmed audio to new file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	// Create encoder - use 16-bit depth as default
	bitDepth := 16
	encoder := wav.NewEncoder(outputFile, sampleRate, bitDepth, channels, 1) // 1 = PCM encoding
	trimmedBuf := &audio.IntBuffer{
		Format: format,
		Data:   trimmedSamples,
	}

	if err := encoder.Write(trimmedBuf); err != nil {
		return nil, fmt.Errorf("failed to encode trimmed audio: %w", err)
	}

	if err := encoder.Close(); err != nil {
		return nil, fmt.Errorf("failed to close encoder: %w", err)
	}

	// Compute hash of trimmed file
	hash := sha256.New()
	if _, err := outputFile.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("failed to seek output file: %w", err)
	}
	if _, err := io.Copy(hash, outputFile); err != nil {
		return nil, fmt.Errorf("failed to compute hash: %w", err)
	}
	contentHash := hex.EncodeToString(hash.Sum(nil))

	// Compare with original hash
	originalHash := ""
	if _, err := file.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("failed to seek original file: %w", err)
	}
	origHash := sha256.New()
	if _, err := io.Copy(origHash, file); err != nil {
		return nil, fmt.Errorf("failed to compute original hash: %w", err)
	}
	originalHash = hex.EncodeToString(origHash.Sum(nil))

	output := &TrimSilenceOutput{
		ContentHash: contentHash,
		WasTrimmed:  true,
		NoOp:        false,
		OutputPath:  outputPath,
	}

	// If hashes are different, create new asset
	if contentHash != originalHash {
		newAssetID := uuid.New().String()
		output.NewAssetID = newAssetID
		// TODO: Register new asset in storage system
	} else {
		// Hashes are identical (shouldn't happen if we trimmed, but handle it)
		output.NoOp = true
		// Clean up the output file since it's identical
		os.Remove(outputPath)
		output.OutputPath = ""
	}

	return output, nil
}

// findNonSilentRange finds the start and end indices of non-silent audio
func findNonSilentRange(samples []int, channels int, threshold float64, sampleRate int, minSilenceDuration float64) (int, int) {
	if len(samples) == 0 {
		return 0, 0
	}

	// Convert threshold to sample value (assuming 16-bit audio, range -32768 to 32767)
	maxSampleValue := 32767.0
	thresholdValue := int(threshold * maxSampleValue)

	// Minimum samples of silence to consider
	minSilenceSamples := int(float64(sampleRate) * minSilenceDuration)

	// Find start (skip leading silence)
	startIdx := 0
	for i := 0; i < len(samples)-channels; i += channels {
		// Check if any channel in this frame is above threshold
		isSilent := true
		for ch := 0; ch < channels; ch++ {
			if i+ch < len(samples) {
				absValue := samples[i+ch]
				if absValue < 0 {
					absValue = -absValue
				}
				if absValue > thresholdValue {
					isSilent = false
					break
				}
			}
		}
		if !isSilent {
			startIdx = i
			break
		}
	}

	// Find end (skip trailing silence)
	endIdx := len(samples)
	silenceCount := 0
	for i := len(samples) - channels; i >= startIdx; i -= channels {
		// Check if any channel in this frame is above threshold
		isSilent := true
		for ch := 0; ch < channels; ch++ {
			if i+ch >= 0 && i+ch < len(samples) {
				absValue := samples[i+ch]
				if absValue < 0 {
					absValue = -absValue
				}
				if absValue > thresholdValue {
					isSilent = false
					break
				}
			}
		}
		if !isSilent {
			endIdx = i + channels
			break
		}
		silenceCount++
		if silenceCount*channels >= minSilenceSamples {
			// We've found enough consecutive silence
			endIdx = i + channels
			break
		}
	}

	// Ensure we have valid indices
	if endIdx <= startIdx {
		return 0, len(samples)
	}

	return startIdx, endIdx
}
