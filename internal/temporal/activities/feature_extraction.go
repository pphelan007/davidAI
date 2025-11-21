package activities

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/go-audio/wav"
	"github.com/google/uuid"
	"github.com/pphelan007/davidAI/internal/database"
	"go.temporal.io/sdk/activity"
)

// ComputeSNR computes the Signal-to-Noise Ratio (SNR) of an audio file in dB.
// SNR is calculated as 10 * log10(signal_power / noise_power).
// Signal power is computed from the RMS of all samples.
// Noise power is estimated from samples below a threshold or from silent segments.
func (ac *ActivitiesClient) ComputeSNR(ctx context.Context, input ComputeSNRInput) (*ComputeSNROutput, error) {
	// Default noise threshold if not provided
	noiseThreshold := input.NoiseThreshold
	if noiseThreshold == 0 {
		noiseThreshold = 0.01 // Default 1% threshold
	}

	// Open and decode the audio file
	// Convert to absolute path to avoid working directory issues
	filePath := input.FilePath
	if !filepath.IsAbs(filePath) {
		// Try to get absolute path
		absPath, err := filepath.Abs(filePath)
		if err == nil {
			filePath = absPath
		}
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open audio file (path: %s, original: %s): %w", filePath, input.FilePath, err)
	}
	defer file.Close()

	// Get file info for better error messages
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat audio file: %w", err)
	}
	fileSize := fileInfo.Size()

	// Ensure file is at beginning (should be, but just to be safe)
	if _, err := file.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("failed to seek to beginning of file: %w", err)
	}

	// Create decoder - use exact same pattern as TrimSilence
	decoder := wav.NewDecoder(file)
	if !decoder.IsValidFile() {
		return nil, fmt.Errorf("file is not a valid WAV file (file: %s, size: %d bytes)", filePath, fileSize)
	}

	format := decoder.Format()
	channels := int(format.NumChannels)

	// Read all audio samples using FullPCMBuffer which allocates the buffer for us
	buf, err := decoder.FullPCMBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to decode audio (file: %s, size: %d bytes): %w", filePath, fileSize, err)
	}
	samples := buf.Data

	if len(samples) == 0 {
		return nil, fmt.Errorf("audio file contains no samples (file: %s, size: %d bytes, sample rate: %d, channels: %d). "+
			"Please verify the file is a valid PCM WAV file", filePath, fileSize, format.SampleRate, format.NumChannels)
	}

	// Convert threshold to sample value (assuming 16-bit audio, range -32768 to 32767)
	maxSampleValue := 32767.0
	thresholdValue := int(noiseThreshold * maxSampleValue)

	// Calculate signal power (mean of squares) and RMS
	var signalSumSquared float64
	for _, sample := range samples {
		sampleFloat := float64(sample)
		signalSumSquared += sampleFloat * sampleFloat
	}

	signalPower := 0.0
	signalRMS := 0.0
	if len(samples) > 0 {
		// Signal power is the mean of squares
		signalPower = signalSumSquared / float64(len(samples))
		// RMS is the square root of the mean of squares
		if signalPower > 0 {
			signalRMS = math.Sqrt(signalPower)
		}
	}

	// Calculate noise power
	var noiseSamples []int
	if input.UseSilentSegments {
		// Estimate noise from silent segments (consecutive samples below threshold)
		// For simplicity, we'll use all samples below threshold
		for i := 0; i < len(samples); i += channels {
			isSilent := true
			for ch := 0; ch < channels && i+ch < len(samples); ch++ {
				absValue := samples[i+ch]
				if absValue < 0 {
					absValue = -absValue
				}
				if absValue > thresholdValue {
					isSilent = false
					break
				}
			}
			if isSilent {
				for ch := 0; ch < channels && i+ch < len(samples); ch++ {
					noiseSamples = append(noiseSamples, samples[i+ch])
				}
			}
		}
	} else {
		// Use all samples below threshold as noise
		for _, sample := range samples {
			absValue := sample
			if absValue < 0 {
				absValue = -absValue
			}
			if absValue <= thresholdValue {
				noiseSamples = append(noiseSamples, sample)
			}
		}
	}

	// Calculate noise power (mean of squares) and RMS
	noisePower := 0.0
	noiseRMS := 0.0
	if len(noiseSamples) > 0 {
		var noiseSumSquared float64
		for _, sample := range noiseSamples {
			sampleFloat := float64(sample)
			noiseSumSquared += sampleFloat * sampleFloat
		}
		// Noise power is the mean of squares
		noisePower = noiseSumSquared / float64(len(noiseSamples))
		// RMS is the square root of the mean of squares
		if noisePower > 0 {
			noiseRMS = math.Sqrt(noisePower)
		}
	} else {
		// If no noise samples found, use a very small value to avoid division by zero
		// This represents the quantization noise floor (1 LSB)
		noisePower = 1.0
		noiseRMS = 1.0
	}

	// Calculate SNR in dB: 10 * log10(signal_power / noise_power)
	snr := 0.0
	if noisePower > 0 && signalPower > 0 {
		ratio := signalPower / noisePower
		if ratio > 0 {
			snr = 10.0 * math.Log10(ratio)
		}
	} else if signalPower > 0 {
		// If noise power is effectively zero, SNR is very high
		// Cap it at a reasonable maximum (e.g., 120 dB)
		snr = 120.0
	}

	output := &ComputeSNROutput{
		SNR:         snr,
		SignalPower: signalPower,
		NoisePower:  noisePower,
		SignalRMS:   signalRMS,
		NoiseRMS:    noiseRMS,
	}

	// Store feature in database if asset ID is provided and db client is available
	if input.AssetID != "" && ac.dbClient != nil {
		featureID := uuid.New().String()

		// Prepare feature data
		featureData := map[string]interface{}{
			"snr":          snr,
			"signal_power": signalPower,
			"noise_power":  noisePower,
			"signal_rms":   signalRMS,
			"noise_rms":    noiseRMS,
		}

		// Prepare computation parameters
		computationParams := map[string]interface{}{
			"noise_threshold":     noiseThreshold,
			"use_silent_segments": input.UseSilentSegments,
		}

		dbFeature := &database.Feature{
			ID:                featureID,
			AssetID:           input.AssetID,
			FeatureType:       "snr",
			FeatureData:       featureData,
			ComputationParams: computationParams,
			ComputedAt:        time.Now(),
		}

		if err := ac.dbClient.InsertFeature(dbFeature); err != nil {
			// Log error but don't fail the activity
			activity.GetLogger(ctx).Error("Failed to insert feature into database", "error", err)
		}
	}

	return output, nil
}
