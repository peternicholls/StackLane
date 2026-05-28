package config

import "path/filepath"

func (m StackMetadata) sharedFilePath(stackHome string) string {
	return filepath.Join(stackHome, m.AssetDir, m.SharedComposeFile)
}

func (m StackMetadata) projectFilePath(stackHome string) string {
	return filepath.Join(stackHome, m.AssetDir, m.ProjectComposeFile)
}

var supportedStacks = map[string]StackMetadata{
	"20i": {
		Kind:               "20i",
		AssetDir:           filepath.Join("stacks", "20i"),
		SharedComposeFile:  "docker-compose.shared.yml",
		ProjectComposeFile: "docker-compose.20i.yml",
		Capabilities: StackCapabilities{
			SharedGateway:   true,
			LocalDNS:        true,
			ManagedTLS:      true,
			ProjectDatabase: true,
			DebugProfile:    true,
		},
		Requirements: []StackRequirement{
			{Name: "gateway", Scope: "shared-runtime", Description: "Shared reverse proxy and shared Docker network must be present."},
			{Name: "local-dns", Scope: "host", Description: "Local DNS resolution must send supported suffixes to StageServe."},
			{Name: "mkcert", Scope: "host", Description: "Local certificate generation is required for HTTPS routes on supported suffixes."},
		},
		Compatibility: StackCompatibility{
			SupportedOS:       []string{"darwin", "linux"},
			SupportedSuffixes: []string{"test", "dev", "develop"},
			SupportedProfiles: []string{"debug"},
		},
	},
}

func lookupStackDefinition(stackKind string) (StackMetadata, bool) {
	def, ok := supportedStacks[normalizeStackKind(stackKind)]
	return def, ok
}

func stackCatalogRootExists(stackHome string) bool {
	for _, def := range supportedStacks {
		if _, err := filepath.Abs(def.sharedFilePath(stackHome)); err != nil {
			continue
		}
		if _, err := osStat(def.sharedFilePath(stackHome)); err == nil {
			return true
		}
	}
	return false
}
