# Detection Test Fixtures

This directory contains golden path test fixtures for validating detection accuracy across all methodology detectors (Speckit, BMAD, etc.).

## 95% Accuracy Requirement

Detection accuracy is a **launch blocker** per PRD NFR-R3. The formula:

```
accuracy = correct_detections / total_fixtures * 100
19/20 = 95% ✅ Pass
18/20 = 90% ❌ Blocked
```

Run `make test-accuracy` to verify.

## Fixture Catalog

| Fixture | Expected Stage | shouldDetect | Purpose |
|---------|----------------|--------------|---------|
| speckit-stage-specify | Specify | true | Standard specify stage |
| speckit-stage-plan | Plan | true | Standard plan stage |
| speckit-stage-tasks | Tasks | true | Standard tasks stage |
| speckit-stage-implement | Implement | true | Standard implement stage |
| speckit-uncertain | Unknown | true | Ambiguous artifacts |
| speckit-dotspecify-marker | Plan | true | .specify/ marker |
| speckit-dotspeckit-marker | Specify | true | .speckit/ marker |
| no-method-detected | Unknown | false | No Speckit markers |
| empty-project | Unknown | false | Empty directory |
| speckit-stage-specify-nested | Unknown | true | Nested directory structure (detector limitation) |
| speckit-stage-plan-with-drafts | Plan | true | Extra draft files |
| speckit-stage-tasks-partial | Tasks | true | Incomplete tasks |
| speckit-stage-implement-complete | Implement | true | Full workflow |
| speckit-multiple-features | Plan | true | Multiple feature dirs |
| speckit-no-spec-subdirs | Unknown | true | Empty specs/ |
| speckit-hidden-files | Unknown | true | Hidden files only |
| speckit-mixed-markers | Specify | true | Both specs/ and .speckit/ |
| speckit-empty-spec-dir | Unknown | true | Empty subdirectory |
| speckit-non-standard-names | Specify | true | Non-numbered dir name |
| speckit-readme-only | Unknown | true | README but no spec.md |

### BMAD Fixtures

| Fixture | Expected Stage | shouldDetect | Purpose |
|---------|----------------|--------------|---------|
| bmad-v6-complete | Implement | true | Full .bmad structure with sprint-status.yaml (epic in-progress) |
| bmad-v6-minimal | Unknown | true | Just .bmad/bmm/config.yaml - no sprint-status or artifacts |
| bmad-v6-no-config | Unknown | true | .bmad folder but no config.yaml |
| bmad-v6-mid-sprint | Implement | true | sprint-status.yaml with one epic done, one in-progress |
| bmad-v6-all-done | Implement | true | All epics marked done |
| bmad-v6-artifacts-only | Implement | true | Has epics.md but no sprint-status - falls back to artifact detection |
| bmad-v4-not-supported | - | false | .bmad-core folder (v4 structure) - should not detect |

## Naming Convention

- `{method}-stage-{stage}` - Standard stage fixtures
- `{method}-{scenario}` - Edge case fixtures

## Adding New Fixtures

1. Create directory following naming convention
2. Add required files per structure
3. Update TestDetectionAccuracy in detector_test.go
4. Update this README
5. Run `make test-accuracy` to verify

## Edge Case Behaviors

| Edge Case | Expected Behavior | Verified By |
|-----------|-------------------|-------------|
| **Nested directories** | Detector looks **one level deep** under marker dir. `specs/group/001-feature/` returns Unknown (artifacts not found at `specs/group/`). | `speckit-stage-specify-nested` |
| **Multiple spec directories** | **Most recently modified** wins. Detector uses directory mod times. | `speckit-multiple-features` |
| **Hidden files** | Files starting with `.` are **NOT** recognized as artifacts. `.spec.md` ≠ `spec.md` | `speckit-hidden-files` |
| **Mixed markers** | Priority order: `specs/` → `.speckit/` → `.specify/`. First match wins. | `speckit-mixed-markers` |
| **Empty subdirectories** | Returns `StageUnknown` with `ConfidenceUncertain` | `speckit-empty-spec-dir` |
| **Empty marker dir** | `specs/` with no subdirs returns `StageUnknown`, `ConfidenceUncertain` | `speckit-no-spec-subdirs` |
| **Non-standard naming** | Subdirs without `NNN-` prefix still work (e.g., `specs/my-feature/`) | `speckit-non-standard-names` |
