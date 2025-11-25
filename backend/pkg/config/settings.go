package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Settings represents application configuration
type Settings struct {
	// Download settings
	DefaultDownloadPath    string `json:"default_download_path"`
	MaxConcurrentDownloads int    `json:"max_concurrent_downloads"`
	DefaultConnections     int    `json:"default_connections"`
	GlobalSpeedLimit       int64  `json:"global_speed_limit"`  // bytes per second, 0 = unlimited
	DefaultSpeedLimit      int64  `json:"default_speed_limit"` // bytes per second, 0 = unlimited
	MinPartSize            int64  `json:"min_part_size"`       // minimum size for multi-part downloads
	MaxRetries             int    `json:"max_retries"`
	RetryDelay             int    `json:"retry_delay"` // seconds

	// Network settings
	UserAgent         string `json:"user_agent"`
	ConnectionTimeout int    `json:"connection_timeout"` // seconds
	ReadTimeout       int    `json:"read_timeout"`       // seconds

	// UI settings
	Theme             string `json:"theme"` // light, dark, auto
	Language          string `json:"language"`
	ShowNotifications bool   `json:"show_notifications"`
	ShowTrayIcon      bool   `json:"show_tray_icon"`
	MinimizeToTray    bool   `json:"minimize_to_tray"`
	StartMinimized    bool   `json:"start_minimized"`

	// Storage settings
	DataDirectory  string `json:"data_directory"`
	AutoBackup     bool   `json:"auto_backup"`
	BackupInterval int    `json:"backup_interval"` // hours
	MaxBackups     int    `json:"max_backups"`

	// Advanced settings
	EnableLogging   bool   `json:"enable_logging"`
	LogLevel        string `json:"log_level"` // debug, info, warn, error
	LogDirectory    string `json:"log_directory"`
	CheckForUpdates bool   `json:"check_for_updates"`
	SendUsageStats  bool   `json:"send_usage_stats"`
}

// NewDefaultSettings creates settings with default values
func NewDefaultSettings() *Settings {
	homeDir, _ := os.UserHomeDir()

	return &Settings{
		// Download defaults
		DefaultDownloadPath:    filepath.Join(homeDir, "Downloads"),
		MaxConcurrentDownloads: 3,
		DefaultConnections:     8,
		GlobalSpeedLimit:       0,
		DefaultSpeedLimit:      0,
		MinPartSize:            1024 * 1024, // 1MB
		MaxRetries:             3,
		RetryDelay:             30,

		// Network defaults
		UserAgent:         "Download Manager 1.0",
		ConnectionTimeout: 30,
		ReadTimeout:       30,

		// UI defaults
		Theme:             "auto",
		Language:          "en",
		ShowNotifications: true,
		ShowTrayIcon:      true,
		MinimizeToTray:    false,
		StartMinimized:    false,

		// Storage defaults
		DataDirectory:  getDefaultDataDirectory(),
		AutoBackup:     true,
		BackupInterval: 24,
		MaxBackups:     7,

		// Advanced defaults
		EnableLogging:   true,
		LogLevel:        "info",
		LogDirectory:    filepath.Join(getDefaultDataDirectory(), "logs"),
		CheckForUpdates: true,
		SendUsageStats:  false,
	}
}

// getDefaultDataDirectory returns the default data directory based on OS
func getDefaultDataDirectory() string {
	homeDir, _ := os.UserHomeDir()

	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData != "" {
			return filepath.Join(appData, "DownloadManager")
		}
		return filepath.Join(homeDir, "AppData", "Roaming", "DownloadManager")
	case "darwin":
		return filepath.Join(homeDir, "Library", "Application Support", "DownloadManager")
	default: // linux and others
		configDir := os.Getenv("XDG_CONFIG_HOME")
		if configDir != "" {
			return filepath.Join(configDir, "download-manager")
		}
		return filepath.Join(homeDir, ".config", "download-manager")
	}
}

// Validate validates the settings and returns any errors
func (s *Settings) Validate() []string {
	var errors []string

	// Validate download path
	if s.DefaultDownloadPath == "" {
		errors = append(errors, "default download path cannot be empty")
	} else {
		if err := os.MkdirAll(s.DefaultDownloadPath, 0755); err != nil {
			errors = append(errors, "cannot create default download path: "+err.Error())
		}
	}

	// Validate concurrent downloads
	if s.MaxConcurrentDownloads < 1 {
		errors = append(errors, "max concurrent downloads must be at least 1")
	}
	if s.MaxConcurrentDownloads > 20 {
		errors = append(errors, "max concurrent downloads should not exceed 20")
	}

	// Validate default connections
	if s.DefaultConnections < 1 {
		errors = append(errors, "default connections must be at least 1")
	}
	if s.DefaultConnections > 32 {
		errors = append(errors, "default connections should not exceed 32")
	}

	// Validate speed limits
	if s.GlobalSpeedLimit < 0 {
		errors = append(errors, "global speed limit cannot be negative")
	}
	if s.DefaultSpeedLimit < 0 {
		errors = append(errors, "default speed limit cannot be negative")
	}

	// Validate part size
	if s.MinPartSize < 1024 {
		errors = append(errors, "minimum part size must be at least 1KB")
	}

	// Validate retries
	if s.MaxRetries < 0 {
		errors = append(errors, "max retries cannot be negative")
	}
	if s.RetryDelay < 1 {
		errors = append(errors, "retry delay must be at least 1 second")
	}

	// Validate timeouts
	if s.ConnectionTimeout < 1 {
		errors = append(errors, "connection timeout must be at least 1 second")
	}
	if s.ReadTimeout < 1 {
		errors = append(errors, "read timeout must be at least 1 second")
	}

	// Validate theme
	validThemes := []string{"light", "dark", "auto"}
	themeValid := false
	for _, theme := range validThemes {
		if s.Theme == theme {
			themeValid = true
			break
		}
	}
	if !themeValid {
		errors = append(errors, "theme must be one of: light, dark, auto")
	}

	// Validate log level
	validLogLevels := []string{"debug", "info", "warn", "error"}
	logLevelValid := false
	for _, level := range validLogLevels {
		if s.LogLevel == level {
			logLevelValid = true
			break
		}
	}
	if !logLevelValid {
		errors = append(errors, "log level must be one of: debug, info, warn, error")
	}

	// Validate data directory
	if s.DataDirectory == "" {
		errors = append(errors, "data directory cannot be empty")
	} else {
		if err := os.MkdirAll(s.DataDirectory, 0755); err != nil {
			errors = append(errors, "cannot create data directory: "+err.Error())
		}
	}

	// Validate backup settings
	if s.BackupInterval < 1 {
		errors = append(errors, "backup interval must be at least 1 hour")
	}
	if s.MaxBackups < 1 {
		errors = append(errors, "max backups must be at least 1")
	}

	return errors
}

// ApplyDefaults applies default values for any missing settings
func (s *Settings) ApplyDefaults() {
	defaults := NewDefaultSettings()

	if s.DefaultDownloadPath == "" {
		s.DefaultDownloadPath = defaults.DefaultDownloadPath
	}
	if s.MaxConcurrentDownloads == 0 {
		s.MaxConcurrentDownloads = defaults.MaxConcurrentDownloads
	}
	if s.DefaultConnections == 0 {
		s.DefaultConnections = defaults.DefaultConnections
	}
	if s.MinPartSize == 0 {
		s.MinPartSize = defaults.MinPartSize
	}
	if s.MaxRetries == 0 {
		s.MaxRetries = defaults.MaxRetries
	}
	if s.RetryDelay == 0 {
		s.RetryDelay = defaults.RetryDelay
	}
	if s.UserAgent == "" {
		s.UserAgent = defaults.UserAgent
	}
	if s.ConnectionTimeout == 0 {
		s.ConnectionTimeout = defaults.ConnectionTimeout
	}
	if s.ReadTimeout == 0 {
		s.ReadTimeout = defaults.ReadTimeout
	}
	if s.Theme == "" {
		s.Theme = defaults.Theme
	}
	if s.Language == "" {
		s.Language = defaults.Language
	}
	if s.DataDirectory == "" {
		s.DataDirectory = defaults.DataDirectory
	}
	if s.BackupInterval == 0 {
		s.BackupInterval = defaults.BackupInterval
	}
	if s.MaxBackups == 0 {
		s.MaxBackups = defaults.MaxBackups
	}
	if s.LogLevel == "" {
		s.LogLevel = defaults.LogLevel
	}
	if s.LogDirectory == "" {
		s.LogDirectory = defaults.LogDirectory
	}
}

// GetSpeedLimitFormatted returns formatted speed limit string
func (s *Settings) GetSpeedLimitFormatted(limit int64) string {
	if limit == 0 {
		return "Unlimited"
	}

	if limit < 1024 {
		return fmt.Sprintf("%d B/s", limit)
	} else if limit < 1024*1024 {
		return fmt.Sprintf("%.1f KB/s", float64(limit)/1024)
	} else {
		return fmt.Sprintf("%.2f MB/s", float64(limit)/(1024*1024))
	}
}

// GetDownloadPathWithSubfolder returns download path with optional subfolder
func (s *Settings) GetDownloadPathWithSubfolder(subfolder string) string {
	if subfolder == "" {
		return s.DefaultDownloadPath
	}
	return filepath.Join(s.DefaultDownloadPath, subfolder)
}
