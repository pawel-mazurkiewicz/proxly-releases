# Implementation Plan: Customizable Browser & Profile Ordering

**Branch**: `002-browser-order-shortcuts` | **Date**: 2025-11-03 | **Spec**: [spec.md](./spec.md)

## Summary

**Primary Requirement**: Enable users to define a custom ordering for browsers and browser profiles that persists across app restarts and browser refresh operations, ensuring consistent keyboard shortcuts (1-9) in the browser selection panel regardless of system-detected browser changes.

**Technical Approach**: Extend the existing `AppSettings` model with a `BrowserOrderState` structure that stores user-defined ordering. The `BrowserDetector` service will be enhanced to respect custom ordering when returning browsers, merging detected browsers with stored order (appending new browsers to the end). The `BrowserManagementView` will be extended with SwiftUI drag-and-drop capabilities for reordering. All UI surfaces (`SelectionPanel`, `ProfileSelectionView`, `RuleEditView`) will use a centralized ordering service to ensure consistency.

## Technical Context

**Language/Version**: Swift 6.0+
**Primary Dependencies**: SwiftUI, Combine, AppKit (for drag-and-drop support)
**Storage**: UserDefaults.standard (via PersistenceManager)
**Testing**: SwiftTest
**Target Platform**: macOS 14.0+
**Project Type**: Desktop Mac app (menu bar accessory)
**Performance Goals**: <100ms UI updates on reorder, sub-50ms URL processing (unchanged)
**Constraints**: Must integrate with existing browser visibility system without modification
**Scale/Scope**: Support up to 50 browsers in custom order per user

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

This feature MUST comply with Proxly Constitution v1.0.0. Check all applicable principles:

### I. macOS Native Integration
- [x] Uses native Apple frameworks (SwiftUI, AppKit, Combine) - SwiftUI for UI, AppKit's NSItemProvider for drag-and-drop
- [x] Maintains macOS 14.0+ compatibility - Using SwiftUI features available in macOS 14+
- [x] Respects accessibility standards (VoiceOver, keyboard navigation) - Will add accessibility labels for reorder handles and position indicators
- [x] Follows Focus Mode integration patterns if applicable - N/A (not touching Focus Mode functionality)

### II. URL Processing Correctness
- [x] Pattern matching logic is stateless (no state in URLProcessingEngine) - Not modifying URLProcessingEngine
- [x] Rule evaluation follows priority order (Domain → Time → Focus) - Not modifying rule evaluation logic
- [x] Behavior is deterministic and well-tested - Ordering is deterministic (explicit user-defined order)
- [x] User input validation with helpful error messages - Drag-and-drop is visual, minimal validation needed

### III. Data Persistence & Sync Integrity
- [x] Uses UserDefaults.standard with JSON encoding - Will store BrowserOrderState in AppSettings via PersistenceManager
- [ ] CloudKit sync uses last-write-wins conflict resolution - **NEEDS RESEARCH**: Determine if browser order should sync via CloudKit or remain device-local
- [x] Schema migrations maintain backward compatibility - Will provide default migration (alphabetical order for existing users)
- [x] Sync triggering is debounced/throttled - Will use existing PersistenceManager debouncing for AppSettings

### IV. Graceful Degradation
- [x] Feature works when ProxlyHelper unavailable (falls back to NSWorkspace) - Not dependent on ProxlyHelper
- [x] Handles missing permissions gracefully - No permissions required for this feature
- [x] Provides user-friendly error messages - Drag-and-drop failures will show toast notifications
- [x] Never crashes on external dependency failure - Robust handling of missing browsers (remove from order)

### V. Privacy & Sandboxing
- [x] Operates within sandbox constraints (Mac App Store version) - No sandbox-restricted operations required
- [x] Minimizes URL logging/storage - Feature doesn't touch URLs
- [x] Requests minimal permissions - No additional permissions needed
- [x] Uses audit tokens for opener detection - N/A (not touching opener detection)

### VI. Performance & Responsiveness
- [x] URL processing <50ms p95 - Not modifying URL processing (ordering applied at UI level)
- [x] Menu bar interaction <100ms - Not modifying menu bar
- [x] Rule evaluation <10ms for typical rule sets - Not modifying rule evaluation
- [x] Non-blocking async operations - Reordering updates will be synchronous (user-initiated, immediate feedback required)
- [x] Regex patterns cached appropriately - N/A (not touching pattern matching)

### VII. Localization & Accessibility
- [x] All strings externalized to Localizable.strings - Will add keys for "Reset Order", position indicators, etc.
- [x] Supports all 7 languages (EN, DE, ES, FR, PL, NL, IT) - Will provide translations for new strings
- [x] VoiceOver labels and hints provided - Will add labels for drag handles, position numbers, reset button
- [x] Keyboard navigation fully supported - Will implement keyboard-based reordering (arrow keys + modifier)
- [x] Sufficient color contrast in all UI - Will use existing AppColors palette

### Testing Requirements
- [x] Unit tests for stateless logic - Will test BrowserOrderState merge logic, migration
- [x] Integration tests for multi-component flows - Will test ordering respected across UI surfaces
- [x] Manual UI testing in Light and Dark mode - Required for drag-and-drop visual feedback
- [x] Migration tests if schema changes - Required for AppSettings migration

### Anti-Patterns Avoided
- [x] No state stored in URLProcessingEngine - Not touching URLProcessingEngine
- [x] No hard-coded browser bundle IDs - Using existing Browser.bundleID
- [x] No blocking operations on main thread - All operations are lightweight (array manipulation)
- [x] No force-unwraps outside test code - Will use safe optional unwrapping

**Gate Status**: ⚠️ CONDITIONAL PASS - Requires Phase 0 research to determine CloudKit sync strategy

## Project Structure

### Documentation (this feature)

```text
specs/002-browser-order-shortcuts/
├── plan.md              # This file
├── research.md          # Phase 0 output (CloudKit sync decision)
├── data-model.md        # Phase 1 output (BrowserOrderState structure)
├── quickstart.md        # Phase 1 output (developer setup guide)
├── contracts/           # Phase 1 output (N/A - internal feature)
├── checklists/
│   └── requirements.md  # Already created during /speckit.specify
└── tasks.md             # Phase 2 output (from /speckit.tasks - NOT YET CREATED)
```

### Source Code (repository root)

```text
browser-chooser/proxly/
├── Proxly/
│   ├── Models/
│   │   ├── AppSettings.swift          (MODIFY: Add browserOrderState property)
│   │   └── BrowserOrderState.swift    (NEW: Data model for browser ordering)
│   ├── Services/
│   │   ├── BrowserDetector.swift      (MODIFY: Add ordering merge logic)
│   │   ├── PersistenceManager.swift   (MODIFY: Handle browserOrderState migration)
│   │   └── BrowserOrderService.swift  (NEW: Centralized ordering logic)
│   ├── UI/
│   │   ├── BrowserManagementView.swift   (MODIFY: Add drag-and-drop reordering)
│   │   ├── SelectionPanel.swift          (MODIFY: Respect custom order)
│   │   ├── ProfileSelectionView.swift    (MODIFY: Respect profile order)
│   │   └── RuleEditView.swift            (MODIFY: Respect order in browser picker)
│   └── Localizable resources (de, en, es, fr, it, nl, pl)/
│       └── Localizable.strings        (MODIFY: Add ordering-related strings)
├── ProxlyTests/
│   ├── BrowserOrderStateTests.swift   (NEW: Unit tests for order merging)
│   ├── BrowserOrderServiceTests.swift (NEW: Unit tests for ordering service)
│   └── BrowserManagementViewTests.swift (NEW: UI tests for drag-and-drop)
└── Proxly.xcodeproj/
```

**Structure Decision**: Following existing Proxly architecture. New `BrowserOrderService` provides centralized ordering logic to avoid duplicating merge logic across UI components.

## Phase 0: Research & Technical Decisions

**Goal**: Resolve all NEEDS CLARIFICATION items and establish technical approach.

### Research Tasks

1. **CloudKit Sync Strategy for Browser Order**
   - **Question**: Should browser order sync across devices via CloudKit or remain device-local?
   - **Research needed**:
     - Review existing CloudKit sync implementation in CloudKitSyncManager
     - Evaluate: Do users want the same browser order on all devices, or device-specific orders?
     - Consider: MacBook Pro might have different browsers than Mac Mini
     - Decision impacts: Whether BrowserOrderState is part of synced AppSettings or a separate local-only structure
   - **Decision criteria**: User value vs. complexity trade-off

2. **SwiftUI Drag-and-Drop Implementation Pattern**
   - **Question**: What's the best approach for drag-and-drop in a ScrollView with LazyVStack?
   - **Research needed**:
     - SwiftUI's `.onDrag()` and `.onDrop()` modifiers on macOS 14+
     - Alternatives: Custom gesture recognizers vs. native drag-and-drop
     - Visual feedback patterns (drag preview, drop indicators, animation)
   - **Decision criteria**: Native feel, accessibility support, macOS 14+ compatibility

3. **Profile Ordering Data Model**
   - **Question**: Should profile order be nested under browser entries or flattened?
   - **Research needed**:
     - Option A: `BrowserOrderEntry { bundleID, order, profileOrder: [String] }` (nested)
     - Option B: Flat list of `OrderEntry { bundleID, profileName?, order }` (flat)
     - Trade-offs: Nested is more intuitive but harder to reorder across browsers; flat is flexible but more complex UI
   - **Decision criteria**: P1 vs P2 implementation complexity, UI ergonomics

4. **Migration Strategy for Existing Users**
   - **Question**: How do we initialize custom order for users upgrading from a version without this feature?
   - **Research needed**:
     - Current default ordering in BrowserDetector (alphabetical by display name?)
     - Whether to preserve current "random" order or impose alphabetical
     - One-time migration flag in AppSettings to avoid re-running
   - **Decision criteria**: Least surprising behavior for existing users

### Expected Outputs

`research.md` will contain:
- **Decision 1**: CloudKit sync strategy (sync order OR device-local only)
- **Decision 2**: SwiftUI drag-and-drop implementation pattern with code examples
- **Decision 3**: Profile ordering data model (nested OR flat)
- **Decision 4**: Migration approach with step-by-step plan

## Phase 1: Design & Contracts

**Prerequisites**: `research.md` complete with all decisions finalized

### Design Artifacts

1. **data-model.md**: Entity definitions
   - `BrowserOrderState` structure (based on Research Decision 3)
   - `BrowserOrderEntry` structure
   - Validation rules (e.g., no duplicate bundleIDs, valid displayOrder indices)
   - State transitions (e.g., browser uninstalled → remove from order)

2. **contracts/** (Internal only - no external API)
   - N/A for this feature (purely internal to Proxly)

3. **quickstart.md**: Developer guide
   - How to test drag-and-drop locally
   - How to simulate browser install/uninstall scenarios
   - How to verify ordering across UI surfaces
   - Manual test checklist for regression testing

### Agent Context Update

After Phase 1 design complete:
```bash
.specify/scripts/bash/update-agent-context.sh claude
```

This will update `.claude/CLAUDE.md` with:
- New technologies: SwiftUI drag-and-drop APIs (if not already documented)
- New data models: BrowserOrderState, BrowserOrderEntry
- New services: BrowserOrderService

## Complexity Tracking

> No Constitution violations requiring justification.

**Rationale**: This feature operates entirely within existing architectural patterns:
- Uses established PersistenceManager for storage
- Extends existing BrowserManagementView UI
- Follows existing service layer pattern (BrowserOrderService similar to BrowserDetector)
- No new external dependencies
- No sandbox-restricted operations

## Next Steps

1. Run `/speckit.plan` to execute Phase 0 and Phase 1 (this command generates research.md, data-model.md, quickstart.md)
2. Review generated artifacts for technical soundness
3. Run `/speckit.tasks` to generate implementation tasks from completed plan
4. Begin implementation following task order