// Package testhelpers provides reusable mock implementations and test utilities
// for testing vibe-dash CLI and TUI components.
//
// This package consolidates common test infrastructure that was previously
// duplicated across multiple test files. It includes:
//
//   - MockRepository: Configurable mock for ports.ProjectRepository
//   - MockDetector: Mock for methodology detection
//   - MockDirectoryManager: Mock for directory operations
//   - ExecuteCommand: Helper for running cobra commands in tests
//
// # MockRepository Usage
//
// Basic usage:
//
//	mock := testhelpers.NewMockRepository()
//	mock.SetSaveError(errors.New("db full"))
//	cli.SetRepository(mock)
//
// Builder pattern (for tests with preset data):
//
//	projects := []*domain.Project{...}
//	mock := testhelpers.NewMockRepository().WithProjects(projects)
//	cli.SetRepository(mock)
//
// Error injection:
//
//	mock := testhelpers.NewMockRepository()
//	mock.SetSaveError(errors.New("save failed"))
//	mock.SetDeleteError(errors.New("delete failed"))
//	mock.SetFindError(errors.New("find failed"))
//	mock.SetFindAllError(errors.New("findall failed"))
//	mock.SetResetError(errors.New("reset failed"))
//
// Call tracking for assertions:
//
//	mock := testhelpers.NewMockRepository()
//	// ... run tests ...
//	if len(mock.SaveCalls()) != 1 {
//	    t.Error("expected one save call")
//	}
//
// # MockDetector Usage
//
// Basic usage:
//
//	detector := testhelpers.NewMockDetector()
//	detector.SetResult(&domain.DetectionResult{...})
//	cli.SetDetectionService(detector)
//
// Error injection:
//
//	detector := testhelpers.NewMockDetector()
//	detector.SetError(errors.New("detection failed"))
//
// # ExecuteCommand Usage
//
// IMPORTANT: Caller must reset command flags before calling ExecuteCommand.
//
//	cli.ResetAddFlags()  // Caller responsibility
//	output, err := testhelpers.ExecuteCommand(
//	    cli.NewRootCmd,
//	    cli.RegisterAddCommand,
//	    "add",
//	    []string{"."},
//	)
//
// For commands requiring stdin input:
//
//	cli.ResetRemoveFlags()
//	output, err := testhelpers.ExecuteCommandWithInput(
//	    cli.NewRootCmd,
//	    cli.RegisterRemoveCommand,
//	    "remove",
//	    []string{"project-name"},
//	    "y\n",
//	)
package testhelpers
