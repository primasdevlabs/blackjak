package protocol

// Well-known host identifiers reported by IDE host adapters.
const (
	HostVSCode   = "vscode"
	HostCursor   = "cursor"
	HostWindsurf = "windsurf"
	HostVSCodium = "vscodium"
	HostTheia    = "theia"
	HostCLI      = "cli"
	HostUnknown  = "unknown"
)

// HostCapabilities describes which IDE features a host adapter supports.
// The agent must treat every capability as optional and degrade gracefully
// when a capability is false (e.g. Theia without SCM support).
type HostCapabilities struct {
	Webview     bool `json:"webview"`
	EditorTabs  bool `json:"editorTabs"`
	FileWatcher bool `json:"fileWatcher"`
	Terminal    bool `json:"terminal"`
	SCM         bool `json:"scm"`
	DiffEditor  bool `json:"diffEditor"`
	Secrets     bool `json:"secrets"`
}

// HostInfo identifies the IDE host driving this agent session. It is sent by
// the host adapter in a host.hello message right after connecting.
type HostInfo struct {
	Name             string           `json:"name"`                       // e.g. "vscode", "cursor", "theia"
	DisplayName      string           `json:"displayName,omitempty"`      // e.g. "Visual Studio Code"
	Version          string           `json:"version,omitempty"`          // host application version
	APIVersion       string           `json:"apiVersion,omitempty"`       // supported VS Code extension API level
	ExtensionVersion string           `json:"extensionVersion,omitempty"` // installed extension version
	Capabilities     HostCapabilities `json:"capabilities"`
}
