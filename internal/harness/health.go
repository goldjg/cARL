package harness

import (
	"fmt"

	"github.com/goldjg/carl/internal/repair"
)

const (
	// PresencePresent indicates the adapter detection file is present on disk.
	PresencePresent = "Present"
	// PresenceMissing indicates the adapter detection file is absent from disk.
	PresenceMissing = "Missing"
	// PresenceUnknown indicates no presence check is available for this adapter.
	PresenceUnknown = "-"

	// SyncSynced indicates all managed adapter files match the canonical source.
	SyncSynced = "Synced"
	// SyncDrifted indicates at least one managed adapter file differs from the canonical source.
	SyncDrifted = "Drifted"
	// SyncMissing indicates at least one managed adapter file is absent from disk.
	SyncMissing = "Missing"
	// SyncUnknown indicates no sync-health check is available for this adapter.
	SyncUnknown = "-"
)

// AdapterHealth captures the presence and sync health of a harness adapter.
type AdapterHealth struct {
	Adapter      Adapter
	Presence     string
	Sync         string
	MissingFiles []string
	DriftedFiles []string
}

// Active reports whether the adapter detection file is present.
func (h AdapterHealth) Active() bool { return h.Presence == PresencePresent }

// Healthy reports whether all managed adapter files are present and canonical.
func (h AdapterHealth) Healthy() bool { return h.Sync == SyncSynced }

// Inspect classifies all known harness adapters in rootDir.
// Presence is based on DetectionFile. Sync health compares AdapterFiles to the
// canonical embedded SourceFile using the same byte-for-byte comparison model
// as runtime artefact drift detection.
func Inspect(rootDir string, arts Artifacts) ([]AdapterHealth, error) {
	result := make([]AdapterHealth, 0, len(knownAdapters))
	for _, a := range knownAdapters {
		health := AdapterHealth{
			Adapter:  a,
			Presence: PresenceUnknown,
			Sync:     SyncUnknown,
		}
		if a.DetectionFile != "" {
			if isDetected(a, rootDir) {
				health.Presence = PresencePresent
			} else {
				health.Presence = PresenceMissing
			}
		}
		if len(a.AdapterFiles) > 0 && a.SourceFile != "" {
			for _, f := range a.AdapterFiles {
				fileMissing, fileDrifted, err := repair.CompareFile(rootDir, f, a.SourceFile, arts)
				if err != nil {
					return nil, fmt.Errorf("inspect harness %q: %w", a.ID, err)
				}
				if fileMissing {
					health.MissingFiles = append(health.MissingFiles, f)
					continue
				}
				if fileDrifted {
					health.DriftedFiles = append(health.DriftedFiles, f)
				}
			}
			switch {
			case len(health.MissingFiles) > 0:
				health.Sync = SyncMissing
			case len(health.DriftedFiles) > 0:
				health.Sync = SyncDrifted
			default:
				health.Sync = SyncSynced
			}
		}
		result = append(result, health)
	}
	return result, nil
}

// Summary aggregates harness health counts for presentation in other commands.
type Summary struct {
	Active  int
	Missing int
	Drifted int
	Healthy int
}

// Summarize derives aggregate counts from adapter health results.
func Summarize(health []AdapterHealth) Summary {
	var summary Summary
	for _, h := range health {
		if h.Active() {
			summary.Active++
		}
		switch h.Sync {
		case SyncMissing:
			summary.Missing++
		case SyncDrifted:
			summary.Drifted++
		case SyncSynced:
			summary.Healthy++
		}
	}
	return summary
}
