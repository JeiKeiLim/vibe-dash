package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestGenerateID(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{"simple path", "/home/user/project"},
		{"nested path", "/home/user/projects/myapp"},
		{"root path", "/project"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := GenerateID(tt.path)

			// ID should be 16 hex characters
			if len(id) != 16 {
				t.Errorf("GenerateID() = %q, want length 16, got %d", id, len(id))
			}

			// ID should be deterministic
			id2 := GenerateID(tt.path)
			if id != id2 {
				t.Errorf("GenerateID() not deterministic: %q != %q", id, id2)
			}

			// ID should only contain hex characters
			for _, c := range id {
				if !strings.ContainsRune("0123456789abcdef", c) {
					t.Errorf("GenerateID() contains non-hex character: %c", c)
				}
			}
		})
	}
}

func TestGenerateID_DifferentPaths(t *testing.T) {
	id1 := GenerateID("/path/one")
	id2 := GenerateID("/path/two")

	if id1 == id2 {
		t.Errorf("GenerateID() should produce different IDs for different paths")
	}
}

func TestNewProject(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		pname   string
		wantErr error
	}{
		{
			name:    "valid absolute path",
			path:    "/home/user/myproject",
			pname:   "myproject",
			wantErr: nil,
		},
		{
			name:    "valid path with custom name",
			path:    "/home/user/myproject",
			pname:   "custom-name",
			wantErr: nil,
		},
		{
			name:    "valid path with empty name derives from path",
			path:    "/home/user/myproject",
			pname:   "",
			wantErr: nil,
		},
		{
			name:    "trailing slash path derives name correctly",
			path:    "/home/user/myproject/",
			pname:   "",
			wantErr: nil,
		},
		{
			name:    "root path uses default name",
			path:    "/",
			pname:   "",
			wantErr: nil,
		},
		{
			name:    "empty path",
			path:    "",
			pname:   "test",
			wantErr: ErrPathNotAccessible,
		},
		{
			name:    "relative path",
			path:    "relative/path",
			pname:   "test",
			wantErr: ErrPathNotAccessible,
		},
		{
			name:    "dot path",
			path:    "./myproject",
			pname:   "test",
			wantErr: ErrPathNotAccessible,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project, err := NewProject(tt.path, tt.pname)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("NewProject() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("NewProject() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("NewProject() unexpected error = %v", err)
				return
			}

			// Verify project fields
			if project.Path != tt.path {
				t.Errorf("project.Path = %q, want %q", project.Path, tt.path)
			}

			// If name was empty, should derive from path
			expectedName := tt.pname
			if expectedName == "" {
				// Extract last segment from path, handling trailing slashes
				cleanPath := strings.TrimRight(tt.path, "/")
				if cleanPath == "" {
					expectedName = "root"
				} else {
					parts := strings.Split(cleanPath, "/")
					expectedName = parts[len(parts)-1]
				}
			}
			if project.Name != expectedName {
				t.Errorf("project.Name = %q, want %q", project.Name, expectedName)
			}

			// Verify ID is generated
			if project.ID == "" {
				t.Error("project.ID should not be empty")
			}
			if len(project.ID) != 16 {
				t.Errorf("project.ID length = %d, want 16", len(project.ID))
			}

			// Verify default state
			if project.State != StateActive {
				t.Errorf("project.State = %v, want StateActive", project.State)
			}

			// Verify timestamps are set
			if project.CreatedAt.IsZero() {
				t.Error("project.CreatedAt should not be zero")
			}
			if project.UpdatedAt.IsZero() {
				t.Error("project.UpdatedAt should not be zero")
			}
			if project.LastActivityAt.IsZero() {
				t.Error("project.LastActivityAt should not be zero")
			}
		})
	}
}

func TestProject_IsHibernated(t *testing.T) {
	tests := []struct {
		name  string
		state ProjectState
		want  bool
	}{
		{"active project", StateActive, false},
		{"hibernated project", StateHibernated, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Project{State: tt.state}
			if got := p.IsHibernated(); got != tt.want {
				t.Errorf("Project.IsHibernated() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProject_IsActive(t *testing.T) {
	tests := []struct {
		name  string
		state ProjectState
		want  bool
	}{
		{"active project", StateActive, true},
		{"hibernated project", StateHibernated, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Project{State: tt.state}
			if got := p.IsActive(); got != tt.want {
				t.Errorf("Project.IsActive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProject_HasDisplayName(t *testing.T) {
	tests := []struct {
		name        string
		displayName string
		want        bool
	}{
		{"empty display name", "", false},
		{"with display name", "My Project", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Project{DisplayName: tt.displayName}
			if got := p.HasDisplayName(); got != tt.want {
				t.Errorf("Project.HasDisplayName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProject_Validate(t *testing.T) {
	tests := []struct {
		name    string
		project *Project
		wantErr bool
	}{
		{
			name: "valid project",
			project: &Project{
				ID:   "abc123def4567890",
				Path: "/home/user/project",
				Name: "project",
			},
			wantErr: false,
		},
		{
			name: "empty path",
			project: &Project{
				ID:   "abc123def4567890",
				Path: "",
				Name: "project",
			},
			wantErr: true,
		},
		{
			name: "relative path",
			project: &Project{
				ID:   "abc123def4567890",
				Path: "relative/path",
				Name: "project",
			},
			wantErr: true,
		},
		{
			name: "empty ID",
			project: &Project{
				ID:   "",
				Path: "/home/user/project",
				Name: "project",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.project.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Project.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewProject_TimestampsAreRecent(t *testing.T) {
	before := time.Now()
	project, err := NewProject("/home/user/project", "test")
	after := time.Now()

	if err != nil {
		t.Fatalf("NewProject() unexpected error = %v", err)
	}

	// All timestamps should be between before and after
	if project.CreatedAt.Before(before) || project.CreatedAt.After(after) {
		t.Errorf("CreatedAt = %v, want between %v and %v", project.CreatedAt, before, after)
	}
	if project.UpdatedAt.Before(before) || project.UpdatedAt.After(after) {
		t.Errorf("UpdatedAt = %v, want between %v and %v", project.UpdatedAt, before, after)
	}
	if project.LastActivityAt.Before(before) || project.LastActivityAt.After(after) {
		t.Errorf("LastActivityAt = %v, want between %v and %v", project.LastActivityAt, before, after)
	}
}

// Story 11.1: Test DaysSinceHibernated helper method
func TestProject_DaysSinceHibernated(t *testing.T) {
	tests := []struct {
		name         string
		hibernatedAt *time.Time
		wantDays     int
	}{
		{
			name:         "nil HibernatedAt returns 0",
			hibernatedAt: nil,
			wantDays:     0,
		},
		{
			name:         "hibernated today returns 0",
			hibernatedAt: func() *time.Time { t := time.Now(); return &t }(),
			wantDays:     0,
		},
		{
			name:         "hibernated 1 day ago returns 1",
			hibernatedAt: func() *time.Time { t := time.Now().Add(-25 * time.Hour); return &t }(),
			wantDays:     1,
		},
		{
			name:         "hibernated 7 days ago returns 7",
			hibernatedAt: func() *time.Time { t := time.Now().Add(-7 * 24 * time.Hour); return &t }(),
			wantDays:     7,
		},
		{
			name:         "hibernated 30 days ago returns 30",
			hibernatedAt: func() *time.Time { t := time.Now().Add(-30 * 24 * time.Hour); return &t }(),
			wantDays:     30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Project{HibernatedAt: tt.hibernatedAt}
			got := p.DaysSinceHibernated()
			if got != tt.wantDays {
				t.Errorf("Project.DaysSinceHibernated() = %v, want %v", got, tt.wantDays)
			}
		})
	}
}
