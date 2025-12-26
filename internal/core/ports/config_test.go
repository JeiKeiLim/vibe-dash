package ports_test

import (
	"context"
	"errors"
	"testing"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

func TestNewConfig_Defaults(t *testing.T) {
	config := ports.NewConfig()

	t.Run("HibernationDays defaults to 14", func(t *testing.T) {
		if config.HibernationDays != 14 {
			t.Errorf("HibernationDays = %d, want 14", config.HibernationDays)
		}
	})

	t.Run("RefreshIntervalSeconds defaults to 10", func(t *testing.T) {
		if config.RefreshIntervalSeconds != 10 {
			t.Errorf("RefreshIntervalSeconds = %d, want 10", config.RefreshIntervalSeconds)
		}
	})

	t.Run("RefreshDebounceMs defaults to 200", func(t *testing.T) {
		if config.RefreshDebounceMs != 200 {
			t.Errorf("RefreshDebounceMs = %d, want 200", config.RefreshDebounceMs)
		}
	})

	t.Run("AgentWaitingThresholdMinutes defaults to 10", func(t *testing.T) {
		if config.AgentWaitingThresholdMinutes != 10 {
			t.Errorf("AgentWaitingThresholdMinutes = %d, want 10", config.AgentWaitingThresholdMinutes)
		}
	})

	t.Run("Projects map is initialized", func(t *testing.T) {
		if config.Projects == nil {
			t.Error("Projects map is nil, want initialized")
		}
	})

	t.Run("DetailLayout defaults to vertical", func(t *testing.T) {
		if config.DetailLayout != "vertical" {
			t.Errorf("DetailLayout = %q, want %q", config.DetailLayout, "vertical")
		}
	})
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *ports.Config
		wantErr bool
	}{
		{
			name:    "valid defaults",
			config:  ports.NewConfig(),
			wantErr: false,
		},
		{
			name: "negative HibernationDays is invalid",
			config: &ports.Config{
				StorageVersion:               2,
				HibernationDays:              -1,
				RefreshIntervalSeconds:       10,
				RefreshDebounceMs:            200,
				AgentWaitingThresholdMinutes: 10,
				Projects:                     make(map[string]ports.ProjectConfig),
			},
			wantErr: true,
		},
		{
			name: "zero HibernationDays is valid (disabled)",
			config: &ports.Config{
				StorageVersion:               2,
				HibernationDays:              0,
				RefreshIntervalSeconds:       10,
				RefreshDebounceMs:            200,
				AgentWaitingThresholdMinutes: 10,
				DetailLayout:                 "vertical",
				Projects:                     make(map[string]ports.ProjectConfig),
			},
			wantErr: false,
		},
		{
			name: "zero RefreshIntervalSeconds is invalid",
			config: &ports.Config{
				StorageVersion:               2,
				HibernationDays:              14,
				RefreshIntervalSeconds:       0,
				RefreshDebounceMs:            200,
				AgentWaitingThresholdMinutes: 10,
				Projects:                     make(map[string]ports.ProjectConfig),
			},
			wantErr: true,
		},
		{
			name: "negative RefreshIntervalSeconds is invalid",
			config: &ports.Config{
				StorageVersion:               2,
				HibernationDays:              14,
				RefreshIntervalSeconds:       -5,
				RefreshDebounceMs:            200,
				AgentWaitingThresholdMinutes: 10,
				Projects:                     make(map[string]ports.ProjectConfig),
			},
			wantErr: true,
		},
		{
			name: "zero RefreshDebounceMs is invalid",
			config: &ports.Config{
				StorageVersion:               2,
				HibernationDays:              14,
				RefreshIntervalSeconds:       10,
				RefreshDebounceMs:            0,
				AgentWaitingThresholdMinutes: 10,
				Projects:                     make(map[string]ports.ProjectConfig),
			},
			wantErr: true,
		},
		{
			name: "negative AgentWaitingThresholdMinutes is invalid",
			config: &ports.Config{
				StorageVersion:               2,
				HibernationDays:              14,
				RefreshIntervalSeconds:       10,
				RefreshDebounceMs:            200,
				AgentWaitingThresholdMinutes: -1,
				Projects:                     make(map[string]ports.ProjectConfig),
			},
			wantErr: true,
		},
		{
			name: "zero AgentWaitingThresholdMinutes is valid (disabled)",
			config: &ports.Config{
				StorageVersion:               2,
				HibernationDays:              14,
				RefreshIntervalSeconds:       10,
				RefreshDebounceMs:            200,
				AgentWaitingThresholdMinutes: 0,
				DetailLayout:                 "vertical",
				Projects:                     make(map[string]ports.ProjectConfig),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			// Verify errors wrap domain.ErrConfigInvalid
			if err != nil && !errors.Is(err, domain.ErrConfigInvalid) {
				t.Errorf("Validate() error should wrap domain.ErrConfigInvalid, got %v", err)
			}
		})
	}
}

func TestConfig_Validate_ProjectConfigOverrides(t *testing.T) {
	negativeVal := -5
	zeroVal := 0
	positiveVal := 7

	tests := []struct {
		name    string
		config  *ports.Config
		wantErr bool
	}{
		{
			name: "negative project HibernationDays is invalid",
			config: func() *ports.Config {
				c := ports.NewConfig()
				c.Projects["test-project"] = ports.ProjectConfig{
					Path:            "/test/path",
					HibernationDays: &negativeVal,
				}
				return c
			}(),
			wantErr: true,
		},
		{
			name: "zero project HibernationDays is valid (disabled)",
			config: func() *ports.Config {
				c := ports.NewConfig()
				c.Projects["test-project"] = ports.ProjectConfig{
					Path:            "/test/path",
					HibernationDays: &zeroVal,
				}
				return c
			}(),
			wantErr: false,
		},
		{
			name: "positive project HibernationDays is valid",
			config: func() *ports.Config {
				c := ports.NewConfig()
				c.Projects["test-project"] = ports.ProjectConfig{
					Path:            "/test/path",
					HibernationDays: &positiveVal,
				}
				return c
			}(),
			wantErr: false,
		},
		{
			name: "negative project AgentWaitingThresholdMinutes is invalid",
			config: func() *ports.Config {
				c := ports.NewConfig()
				c.Projects["test-project"] = ports.ProjectConfig{
					Path:                         "/test/path",
					AgentWaitingThresholdMinutes: &negativeVal,
				}
				return c
			}(),
			wantErr: true,
		},
		{
			name: "zero project AgentWaitingThresholdMinutes is valid (disabled)",
			config: func() *ports.Config {
				c := ports.NewConfig()
				c.Projects["test-project"] = ports.ProjectConfig{
					Path:                         "/test/path",
					AgentWaitingThresholdMinutes: &zeroVal,
				}
				return c
			}(),
			wantErr: false,
		},
		{
			name: "nil Projects map is valid",
			config: &ports.Config{
				StorageVersion:               2,
				HibernationDays:              14,
				RefreshIntervalSeconds:       10,
				RefreshDebounceMs:            200,
				AgentWaitingThresholdMinutes: 10,
				DetailLayout:                 "vertical",
				Projects:                     nil,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !errors.Is(err, domain.ErrConfigInvalid) {
				t.Errorf("Validate() error should wrap domain.ErrConfigInvalid, got %v", err)
			}
		})
	}
}

func TestConfig_GetProjectConfig(t *testing.T) {
	config := ports.NewConfig()

	// Add a project config
	config.Projects["project-123"] = ports.ProjectConfig{
		Path:        "/test/path",
		DisplayName: "Test Project",
		IsFavorite:  true,
	}

	t.Run("returns project config when exists", func(t *testing.T) {
		pc, ok := config.GetProjectConfig("project-123")
		if !ok {
			t.Fatal("GetProjectConfig() ok = false, want true")
		}
		if pc.DisplayName != "Test Project" {
			t.Errorf("GetProjectConfig().DisplayName = %q, want %q", pc.DisplayName, "Test Project")
		}
	})

	t.Run("returns false when project not found", func(t *testing.T) {
		_, ok := config.GetProjectConfig("nonexistent")
		if ok {
			t.Error("GetProjectConfig() ok = true for nonexistent, want false")
		}
	})
}

func TestConfig_GetEffectiveHibernationDays(t *testing.T) {
	config := ports.NewConfig()
	config.HibernationDays = 14

	// Project without override
	config.Projects["no-override"] = ports.ProjectConfig{
		Path: "/test/no-override",
	}

	// Project with override
	overrideDays := 7
	config.Projects["with-override"] = ports.ProjectConfig{
		Path:            "/test/with-override",
		HibernationDays: &overrideDays,
	}

	t.Run("returns global when no project override", func(t *testing.T) {
		days := config.GetEffectiveHibernationDays("no-override")
		if days != 14 {
			t.Errorf("GetEffectiveHibernationDays() = %d, want 14", days)
		}
	})

	t.Run("returns project override when set", func(t *testing.T) {
		days := config.GetEffectiveHibernationDays("with-override")
		if days != 7 {
			t.Errorf("GetEffectiveHibernationDays() = %d, want 7", days)
		}
	})

	t.Run("returns global when project not found", func(t *testing.T) {
		days := config.GetEffectiveHibernationDays("nonexistent")
		if days != 14 {
			t.Errorf("GetEffectiveHibernationDays() = %d, want 14", days)
		}
	})
}

func TestConfig_GetEffectiveWaitingThreshold(t *testing.T) {
	config := ports.NewConfig()
	config.AgentWaitingThresholdMinutes = 10

	// Project without override
	config.Projects["no-override"] = ports.ProjectConfig{
		Path: "/test/no-override",
	}

	// Project with override
	overrideMinutes := 5
	config.Projects["with-override"] = ports.ProjectConfig{
		Path:                         "/test/with-override",
		AgentWaitingThresholdMinutes: &overrideMinutes,
	}

	t.Run("returns global when no project override", func(t *testing.T) {
		minutes := config.GetEffectiveWaitingThreshold("no-override")
		if minutes != 10 {
			t.Errorf("GetEffectiveWaitingThreshold() = %d, want 10", minutes)
		}
	})

	t.Run("returns project override when set", func(t *testing.T) {
		minutes := config.GetEffectiveWaitingThreshold("with-override")
		if minutes != 5 {
			t.Errorf("GetEffectiveWaitingThreshold() = %d, want 5", minutes)
		}
	})

	t.Run("returns global when project not found", func(t *testing.T) {
		minutes := config.GetEffectiveWaitingThreshold("nonexistent")
		if minutes != 10 {
			t.Errorf("GetEffectiveWaitingThreshold() = %d, want 10", minutes)
		}
	})
}

func TestProjectConfig_Struct(t *testing.T) {
	overrideDays := 21
	overrideMinutes := 15

	pc := ports.ProjectConfig{
		Path:                         "/test/path",
		DisplayName:                  "My Project",
		IsFavorite:                   true,
		HibernationDays:              &overrideDays,
		AgentWaitingThresholdMinutes: &overrideMinutes,
	}

	if pc.Path != "/test/path" {
		t.Errorf("ProjectConfig.Path = %q, want %q", pc.Path, "/test/path")
	}
	if pc.DisplayName != "My Project" {
		t.Errorf("ProjectConfig.DisplayName = %q, want %q", pc.DisplayName, "My Project")
	}
	if !pc.IsFavorite {
		t.Error("ProjectConfig.IsFavorite = false, want true")
	}
	if *pc.HibernationDays != 21 {
		t.Errorf("ProjectConfig.HibernationDays = %d, want 21", *pc.HibernationDays)
	}
	if *pc.AgentWaitingThresholdMinutes != 15 {
		t.Errorf("ProjectConfig.AgentWaitingThresholdMinutes = %d, want 15", *pc.AgentWaitingThresholdMinutes)
	}
}

// mockConfigLoader verifies interface compliance at compile time
type mockConfigLoader struct {
	config *ports.Config
}

func (m *mockConfigLoader) Load(ctx context.Context) (*ports.Config, error) {
	// Check context cancellation (interface contract)
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if m.config == nil {
		return ports.NewConfig(), nil
	}
	return m.config, nil
}

func (m *mockConfigLoader) Save(ctx context.Context, config *ports.Config) error {
	// Check context cancellation (interface contract)
	if err := ctx.Err(); err != nil {
		return err
	}
	m.config = config
	return nil
}

// Compile-time interface compliance check
var _ ports.ConfigLoader = (*mockConfigLoader)(nil)

// Tests for Story 3.5.4: Master Config as Path Index

func TestNewConfig_StorageVersionDefault(t *testing.T) {
	config := ports.NewConfig()

	if config.StorageVersion != 2 {
		t.Errorf("StorageVersion = %d, want 2", config.StorageVersion)
	}
}

func TestConfig_Validate_StorageVersion(t *testing.T) {
	tests := []struct {
		name           string
		storageVersion int
		wantErr        bool
	}{
		{
			name:           "storage version 2 is valid",
			storageVersion: 2,
			wantErr:        false,
		},
		{
			name:           "storage version 0 is invalid",
			storageVersion: 0,
			wantErr:        true,
		},
		{
			name:           "storage version 1 is invalid",
			storageVersion: 1,
			wantErr:        true,
		},
		{
			name:           "storage version 3 is invalid",
			storageVersion: 3,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ports.NewConfig()
			config.StorageVersion = tt.storageVersion
			err := config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !errors.Is(err, domain.ErrConfigInvalid) {
				t.Errorf("Validate() error should wrap domain.ErrConfigInvalid, got %v", err)
			}
		})
	}
}

func TestProjectConfig_DirectoryName(t *testing.T) {
	pc := ports.ProjectConfig{
		Path:          "/home/user/api-service",
		DirectoryName: "api-service",
		DisplayName:   "API Service",
		IsFavorite:    true,
	}

	if pc.DirectoryName != "api-service" {
		t.Errorf("ProjectConfig.DirectoryName = %q, want %q", pc.DirectoryName, "api-service")
	}
}

// Story 8.6: DetailLayout validation tests
func TestConfig_Validate_DetailLayout(t *testing.T) {
	tests := []struct {
		name         string
		detailLayout string
		wantErr      bool
	}{
		{
			name:         "vertical is valid",
			detailLayout: "vertical",
			wantErr:      false,
		},
		{
			name:         "horizontal is valid",
			detailLayout: "horizontal",
			wantErr:      false,
		},
		{
			name:         "empty string is invalid",
			detailLayout: "",
			wantErr:      true,
		},
		{
			name:         "diagonal is invalid",
			detailLayout: "diagonal",
			wantErr:      true,
		},
		{
			name:         "VERTICAL (uppercase) is invalid",
			detailLayout: "VERTICAL",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ports.NewConfig()
			config.DetailLayout = tt.detailLayout
			err := config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !errors.Is(err, domain.ErrConfigInvalid) {
				t.Errorf("Validate() error should wrap domain.ErrConfigInvalid, got %v", err)
			}
		})
	}
}

func TestConfigLoader_InterfaceCompliance(t *testing.T) {
	var loader ports.ConfigLoader = &mockConfigLoader{}
	ctx := context.Background()

	t.Run("Load accepts context and returns Config", func(t *testing.T) {
		config, err := loader.Load(ctx)
		if err != nil {
			t.Fatalf("Load() error = %v, want nil", err)
		}
		if config == nil {
			t.Fatal("Load() returned nil config")
		}
	})

	t.Run("Save accepts context and Config", func(t *testing.T) {
		config := ports.NewConfig()
		config.HibernationDays = 7

		err := loader.Save(ctx, config)
		if err != nil {
			t.Fatalf("Save() error = %v, want nil", err)
		}

		// Verify saved
		loaded, _ := loader.Load(ctx)
		if loaded.HibernationDays != 7 {
			t.Errorf("Save() did not persist config, HibernationDays = %d, want 7", loaded.HibernationDays)
		}
	})

	t.Run("Load respects context cancellation", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		freshLoader := &mockConfigLoader{}
		_, err := freshLoader.Load(cancelCtx)
		if err == nil {
			t.Error("Load() with cancelled context should return error")
		}
	})

	t.Run("Save respects context cancellation", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		freshLoader := &mockConfigLoader{}
		err := freshLoader.Save(cancelCtx, ports.NewConfig())
		if err == nil {
			t.Error("Save() with cancelled context should return error")
		}
	})
}

// Story 3.5.4: Path Index Lookup Methods Tests

// TestConfig_GetDirForPath tests ProjectPathLookup interface compliance (Subtask 3.1, 6.4)
func TestConfig_GetDirForPath(t *testing.T) {
	tests := []struct {
		name     string
		projects map[string]ports.ProjectConfig
		path     string
		wantDir  string
	}{
		{
			name: "existing project returns directory name",
			projects: map[string]ports.ProjectConfig{
				"api-service": {Path: "/home/user/api-service", DirectoryName: "api-service"},
			},
			path:    "/home/user/api-service",
			wantDir: "api-service",
		},
		{
			name:     "non-existent path returns empty string",
			projects: map[string]ports.ProjectConfig{},
			path:     "/non/existent",
			wantDir:  "",
		},
		{
			name: "disambiguated directory name",
			projects: map[string]ports.ProjectConfig{
				"client-b-api-service": {Path: "/home/user/client-b/api-service", DirectoryName: "client-b-api-service"},
			},
			path:    "/home/user/client-b/api-service",
			wantDir: "client-b-api-service",
		},
		{
			name:     "nil projects returns empty string",
			projects: nil,
			path:     "/any/path",
			wantDir:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &ports.Config{StorageVersion: 2, Projects: tt.projects}
			got := cfg.GetDirForPath(tt.path)
			if got != tt.wantDir {
				t.Errorf("GetDirForPath() = %q, want %q", got, tt.wantDir)
			}
		})
	}
}

// TestConfig_GetDirectoryName tests GetDirectoryName method (Subtask 3.2, 6.5)
func TestConfig_GetDirectoryName(t *testing.T) {
	tests := []struct {
		name      string
		projects  map[string]ports.ProjectConfig
		path      string
		wantDir   string
		wantFound bool
	}{
		{
			name: "existing project",
			projects: map[string]ports.ProjectConfig{
				"api-service": {Path: "/home/user/api-service", DirectoryName: "api-service"},
			},
			path:      "/home/user/api-service",
			wantDir:   "api-service",
			wantFound: true,
		},
		{
			name:      "non-existent project",
			projects:  map[string]ports.ProjectConfig{},
			path:      "/non/existent",
			wantDir:   "",
			wantFound: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &ports.Config{StorageVersion: 2, Projects: tt.projects}
			gotDir, gotFound := cfg.GetDirectoryName(tt.path)
			if gotDir != tt.wantDir {
				t.Errorf("GetDirectoryName() dir = %q, want %q", gotDir, tt.wantDir)
			}
			if gotFound != tt.wantFound {
				t.Errorf("GetDirectoryName() found = %v, want %v", gotFound, tt.wantFound)
			}
		})
	}
}

// TestConfig_GetProjectPath tests GetProjectPath method (Subtask 3.3, 6.6)
func TestConfig_GetProjectPath(t *testing.T) {
	tests := []struct {
		name      string
		projects  map[string]ports.ProjectConfig
		dirName   string
		wantPath  string
		wantFound bool
	}{
		{
			name: "existing project",
			projects: map[string]ports.ProjectConfig{
				"api-service": {Path: "/home/user/api-service", DirectoryName: "api-service"},
			},
			dirName:   "api-service",
			wantPath:  "/home/user/api-service",
			wantFound: true,
		},
		{
			name:      "non-existent directory name",
			projects:  map[string]ports.ProjectConfig{},
			dirName:   "nonexistent",
			wantPath:  "",
			wantFound: false,
		},
		{
			name:      "nil projects",
			projects:  nil,
			dirName:   "any",
			wantPath:  "",
			wantFound: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &ports.Config{StorageVersion: 2, Projects: tt.projects}
			gotPath, gotFound := cfg.GetProjectPath(tt.dirName)
			if gotPath != tt.wantPath {
				t.Errorf("GetProjectPath() path = %q, want %q", gotPath, tt.wantPath)
			}
			if gotFound != tt.wantFound {
				t.Errorf("GetProjectPath() found = %v, want %v", gotFound, tt.wantFound)
			}
		})
	}
}

// TestConfig_SetProjectEntry tests SetProjectEntry method (Subtask 3.4, 6.7)
func TestConfig_SetProjectEntry(t *testing.T) {
	t.Run("adds new project to nil map", func(t *testing.T) {
		cfg := &ports.Config{StorageVersion: 2, Projects: nil}
		cfg.SetProjectEntry("api-service", "/path/to/api", "API Service", true)

		if cfg.Projects == nil {
			t.Fatal("Projects map should be initialized")
		}

		pc, ok := cfg.Projects["api-service"]
		if !ok {
			t.Fatal("api-service not found in projects")
		}
		if pc.Path != "/path/to/api" {
			t.Errorf("Path = %q, want %q", pc.Path, "/path/to/api")
		}
		if pc.DirectoryName != "api-service" {
			t.Errorf("DirectoryName = %q, want %q", pc.DirectoryName, "api-service")
		}
		if pc.DisplayName != "API Service" {
			t.Errorf("DisplayName = %q, want %q", pc.DisplayName, "API Service")
		}
		if !pc.IsFavorite {
			t.Error("IsFavorite = false, want true")
		}
	})

	t.Run("updates existing project", func(t *testing.T) {
		cfg := &ports.Config{
			StorageVersion: 2,
			Projects: map[string]ports.ProjectConfig{
				"api-service": {Path: "/old/path", DirectoryName: "api-service"},
			},
		}
		cfg.SetProjectEntry("api-service", "/new/path", "New Name", false)

		pc := cfg.Projects["api-service"]
		if pc.Path != "/new/path" {
			t.Errorf("Path = %q, want %q", pc.Path, "/new/path")
		}
		if pc.DisplayName != "New Name" {
			t.Errorf("DisplayName = %q, want %q", pc.DisplayName, "New Name")
		}
	})
}

// TestConfig_RemoveProject tests RemoveProject method (Subtask 3.5, 6.8)
func TestConfig_RemoveProject(t *testing.T) {
	t.Run("removes existing project", func(t *testing.T) {
		cfg := &ports.Config{
			StorageVersion: 2,
			Projects: map[string]ports.ProjectConfig{
				"api-service": {Path: "/path/to/api", DirectoryName: "api-service"},
			},
		}
		removed := cfg.RemoveProject("api-service")

		if !removed {
			t.Error("RemoveProject() = false, want true")
		}
		if _, ok := cfg.Projects["api-service"]; ok {
			t.Error("api-service should be removed from projects")
		}
	})

	t.Run("returns false for non-existent project", func(t *testing.T) {
		cfg := &ports.Config{StorageVersion: 2, Projects: map[string]ports.ProjectConfig{}}
		removed := cfg.RemoveProject("nonexistent")

		if removed {
			t.Error("RemoveProject() = true for non-existent, want false")
		}
	})

	t.Run("handles nil projects", func(t *testing.T) {
		cfg := &ports.Config{StorageVersion: 2, Projects: nil}
		removed := cfg.RemoveProject("any")

		if removed {
			t.Error("RemoveProject() = true for nil projects, want false")
		}
	})
}

// Compile-time interface compliance check (Subtask 3.6)
var _ ports.ProjectPathLookup = (*ports.Config)(nil)
