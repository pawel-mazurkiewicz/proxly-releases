# Implementation Plan: Custom URL Transformations

**Branch**: `001-custom-url-transform` | **Date**: 2025-10-31 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/001-custom-url-transform/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

**Primary Requirement**: Enable users to create custom URL transformations using template-based or JavaScript-based logic. Transformations apply arbitrary modifications to URLs before opening them (e.g., strip tracking parameters, redirect to alternative frontends, modify URL structure for native app deeplinking). System must provide real-time testing, validation, and CloudKit sync support.

**Technical Approach**: Extend existing Rule system with CustomTransformation capability. Implement two transformation modes: (1) Template-based using placeholder variables for URL components, (2) JavaScript-based using sandboxed JavaScriptCore execution. Transformations integrate into existing URLProcessor flow after rule matching but before browser launching. Add UI for creating/editing/testing transformations with real-time preview. Persist transformations with CloudKit sync using existing infrastructure.

## Technical Context

**Language/Version**: Swift 6.0+
**Primary Dependencies**: SwiftUI, Combine, JavaScriptCore, CloudKit, Sparkle (Standalone version)
**Storage**: UserDefaults (local), CloudKit (sync)
**Testing**: SwiftTest
**Target Platform**: macOS 14.0+
**Project Type**: Desktop Mac app (menu bar accessory)
**Performance Goals**:
- URL transformation: <100ms p95
- JavaScript execution: <100ms p95 (1s timeout enforced)
- Template substitution: <10ms p95
- Real-time preview: <100ms response
**Constraints**:
- JavaScript sandbox: No network, file system, or system API access
- CloudKit quota limits (must debounce sync)
**Scale/Scope**:
- Expected 1-5 custom transformations per user
- Support up to 100 transformations technically

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

This feature MUST comply with Proxly Constitution v1.0.0. Check all applicable principles:

### I. macOS Native Integration
- [x] Uses native Apple frameworks (SwiftUI, AppKit, Combine, JavaScriptCore)
- [x] Maintains macOS 14.0+ compatibility
- [x] Respects accessibility standards (VoiceOver, keyboard navigation) - UI will include labels/hints
- [N/A] Follows Focus Mode integration patterns if applicable - No Focus Mode interaction needed

### II. URL Processing Correctness
- [x] Pattern matching logic is stateless (no state in URLProcessingEngine) - Transformation engine will be stateless
- [x] Rule evaluation follows priority order (Domain → Time → Focus) - Transformations apply after rule matching
- [x] Behavior is deterministic and well-tested - Template expansion is deterministic; JS timeout ensures determinism
- [x] User input validation with helpful error messages - Real-time validation with syntax checking

### III. Data Persistence & Sync Integrity
- [x] Uses UserDefaults.standard with JSON encoding
- [x] CloudKit sync uses last-write-wins conflict resolution - Reuse existing CloudKitSyncManager
- [x] Schema migrations maintain backward compatibility - New optional fields on Rule model
- [x] Sync triggering is debounced/throttled - Reuse existing sync infrastructure

### IV. Graceful Degradation
- [x] Feature works when ProxlyHelper unavailable (falls back to NSWorkspace) - No ProxlyHelper dependency
- [x] Handles missing permissions gracefully - No special permissions needed
- [x] Provides user-friendly error messages - Validation errors shown inline; transformation failures fallback
- [x] Never crashes on external dependency failure - JS sandbox catches all errors; timeout prevents hangs

### V. Privacy & Sandboxing
- [x] Operates within sandbox constraints (Mac App Store version) - JavaScriptCore runs in-process, sandboxed
- [x] Minimizes URL logging/storage - Only transformed URLs logged for debugging (user can disable)
- [x] Requests minimal permissions - No new permissions required
- [N/A] Uses audit tokens for opener detection - Not needed for transformations

### VI. Performance & Responsiveness
- [x] URL processing <50ms p95 - Template transformations <10ms; JS transformations <100ms target
- [N/A] Menu bar interaction <100ms - No menu bar changes
- [x] Rule evaluation <10ms for typical rule sets - Transformation adds minimal overhead
- [x] Non-blocking async operations - Transformation runs synchronously but is fast enough
- [x] Regex patterns cached appropriately - RESOLVED: Cache regex pattern for template parsing, not results

### VII. Localization & Accessibility
- [x] All strings externalized to Localizable.strings - UI strings will be externalized
- [x] Supports all 7 languages (EN, DE, ES, FR, PL, NL, IT) - Translations will be added
- [x] VoiceOver labels and hints provided - Transformation editor UI will include accessibility
- [x] Keyboard navigation fully supported - Standard SwiftUI keyboard navigation
- [x] Sufficient color contrast in all UI - Follow existing Proxly design patterns

### Testing Requirements
- [x] Unit tests for stateless logic - Template parser, JS executor, validators
- [x] Integration tests for multi-component flows - End-to-end transformation tests
- [x] Manual UI testing in Light and Dark mode - Transformation editor will be tested in both modes
- [x] Migration tests if schema changes - Test Rule model with/without transformation fields

### Anti-Patterns Avoided
- [x] No state stored in URLProcessingEngine - Create separate stateless TransformationEngine
- [N/A] No hard-coded browser bundle IDs - Not applicable to transformations
- [x] No blocking operations on main thread - JS execution is fast (<100ms) with timeout
- [x] No force-unwraps outside test code - Use safe optional unwrapping throughout

**Initial Assessment**: ✅ PASS with 1 clarification needed (regex caching strategy)

**Post-Design Re-evaluation**: ✅ PASS - All clarifications resolved

- Regex caching: Cache the compiled regex pattern (NSRegularExpression), not template evaluation results
- All constitutional requirements satisfied
- No complexity violations requiring justification
- Ready for implementation phase

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)
<!--
  ACTION REQUIRED: Replace the placeholder tree below with the concrete layout
  for this feature. Delete unused options and expand the chosen structure with
  real paths (e.g., apps/admin, packages/something). The delivered plan must
  not include Option labels.
-->

```text
# Single project
browser-chooser/
├── proxly/
│   ├── Proxly/                    (Main SwiftUI app - common code)
│   │   ├── Assets.xcassets/
│   │   ├── Models/
│   │   ├── Services/
│   │   ├── UI/
│   │   ├── Utilities/
│   │   ├── Localizable resources (de, en, es, fr, it, nl, pl)
│   │   └── ProxlyApp.swift
│   ├── ProxlyHelper/              (Helper app - only for Mac App Store version of the app)
│   ├── ProxlyTests/               (Test suite)
│   ├── SafariExtension/
│   ├── proxly-license-server/     (License Server Go backend - only for standalone app)
│   │   ├── cmd/
│   │   ├── internal/
│   │   ├── api/
│   │   ├── pkg/
│   │   ├── scripts/
│   │   └── docker/
│   ├── Proxly.xcodeproj/ (Proxly XCode project)
│   └── docs/
```

**Structure Decision**: It's more or less following standard SwiftUI project structure.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |
