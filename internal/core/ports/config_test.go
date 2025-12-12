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
				HibernationDays:              0,
				RefreshIntervalSeconds:       10,
				RefreshDebounceMs:            200,
				AgentWaitingThresholdMinutes: 10,
				Projects:                     make(map[string]ports.ProjectConfig),
			},
			wantErr: false,
		},
		{
			name: "zero RefreshIntervalSeconds is invalid",
			config: &ports.Config{
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
				HibernationDays:              14,
				RefreshIntervalSeconds:       10,
				RefreshDebounceMs:            200,
				AgentWaitingThresholdMinutes: 0,
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
				HibernationDays:              14,
				RefreshIntervalSeconds:       10,
				RefreshDebounceMs:            200,
				AgentWaitingThresholdMinutes: 10,
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
