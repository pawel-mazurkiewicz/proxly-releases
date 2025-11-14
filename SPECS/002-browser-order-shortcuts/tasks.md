# Tasks: Customizable Browser & Profile Ordering with Keyboard Shortcuts

**Feature**: 002-browser-order-shortcuts
**Input**: Design documents from `/specs/002-browser-order-shortcuts/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md

**Tests**: Test tasks are included only where explicitly needed for validation. Manual testing guidance is provided in quickstart.md.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Include exact file paths in descriptions

## Path Conventions

All paths are relative to repository root: `/Users/pawelma/code/browser-chooser/proxly/`

- **Models**: `Proxly/Models/`
- **Services**: `Proxly/Services/`
- **UI**: `Proxly/UI/`
- **Tests**: `ProxlyTests/`
- **Localizable**: `Proxly/[language].lproj/Localizable.strings` (de, en, es, fr, it, nl, pl)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create data model and migration infrastructure

- [X] T001 [P] Create BrowserOrderState.swift model in Proxly/Models/BrowserOrderState.swift
- [X] T002 [P] Create BrowserOrderEntry struct within BrowserOrderState.swift with Transferable conformance
- [X] T003 Add migration to version 6 in Proxly/Services/MigrationManager.swift (alphabetical initialization)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core service layer that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T004 Create BrowserOrderService.swift singleton in Proxly/Services/BrowserOrderService.swift
- [X] T005 Implement loadOrder() method in BrowserOrderService (reads from PersistenceManager)
- [X] T006 Implement saveOrder() method in BrowserOrderService (writes to PersistenceManager)
- [X] T007 Implement getOrderedBrowsers() method in BrowserOrderService (merges detected browsers with custom order)
- [X] T008 Add saveBrowserOrder() method to Proxly/Services/PersistenceManager.swift (device-local UserDefaults)
- [X] T009 Add loadBrowserOrder() method to Proxly/Services/PersistenceManager.swift (JSON decoding with graceful defaults)
- [X] T010 Update BrowserDetector.swift to use BrowserOrderService.getOrderedBrowsers() for ordering

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Custom Browser Ordering (Priority: P1) 🎯 MVP

**Goal**: Enable users to drag-and-drop reorder browsers in the Manage Browsers popup, with custom order persisting across app restarts and browser refresh operations. Keyboard shortcuts 1-9 map to custom order.

**Independent Test**:
1. Open Manage Browsers popup
2. Drag Safari to top position
3. Close popup and trigger URL
4. Press key "1" in browser selection panel
5. Verify Safari opens (test passes if Safari opens, not original first browser)
6. Restart app and verify order persists

### Implementation for User Story 1

- [X] T011 [P] [US1] Add browserOrderState property to Proxly/Models/AppSettings.swift (SKIPPED - not applicable, browser order stored separately per research)
- [X] T012 [US1] Extend BrowserManagementView.swift with drag-and-drop modifiers (.draggable() and .dropDestination()) on browser cards
- [X] T013 [US1] Implement moveBrowser(dragged:droppedOn:) helper method in BrowserManagementView.swift (updates availableBrowsers array)
- [X] T014 [US1] Add visual position indicators (1, 2, 3...) to browser cards in BrowserManagementView.swift
- [X] T015 [US1] Add "Reset Order" button to BrowserManagementView.swift (calls BrowserOrderService.resetToAlphabetical())
- [X] T016 [US1] Update SelectionPanel.swift to use BrowserOrderService.getOrderedBrowsers() for display order
- [X] T017 [US1] Update keyboard shortcut mapping in SelectionPanel.swift to respect custom order (filter by visibility first)
- [X] T018 [US1] Update RuleEditView.swift browser picker to use BrowserOrderService.getOrderedBrowsers()
- [X] T019 [US1] Implement handleBrowserUninstall() in BrowserOrderService (removes entries, renumbers remaining)
- [X] T020 [US1] Implement handleNewBrowserDetected() in BrowserOrderService (appends to end of order)

### Localization for User Story 1

- [X] T021 [P] [US1] Add localization keys to Proxly/en.lproj/Localizable.strings (reset_order_button, position_indicator)
- [X] T022 [P] [US1] Add German translations to Proxly/de.lproj/Localizable.strings
- [X] T023 [P] [US1] Add Spanish translations to Proxly/es.lproj/Localizable.strings
- [X] T024 [P] [US1] Add French translations to Proxly/fr.lproj/Localizable.strings
- [X] T025 [P] [US1] Add Polish translations to Proxly/pl.lproj/Localizable.strings
- [X] T026 [P] [US1] Add Dutch translations to Proxly/nl.lproj/Localizable.strings
- [X] T027 [P] [US1] Add Italian translations to Proxly/it.lproj/Localizable.strings

### Accessibility for User Story 1

- [X] T028 [US1] Add .accessibilityLabel() to browser cards with position info in BrowserManagementView.swift
- [X] T029 [US1] Add .accessibilityAdjustableAction() to browser cards for keyboard-based reordering in BrowserManagementView.swift
- [X] T030 [US1] Add .accessibilityHint() to drag handles with reorder instructions in BrowserManagementView.swift

### Tests for User Story 1

- [X] T031 [P] [US1] Create BrowserOrderStateTests.swift in ProxlyTests/ (Codable roundtrip, validation)
- [X] T032 [P] [US1] Create BrowserOrderServiceTests.swift in ProxlyTests/ (merge logic, append, remove, renumber)
- [X] T033 [P] [US1] Add browser order migration test to ProxlyTests/MigrationManagerTests.swift (verify alphabetical init)

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently. Users can reorder browsers, order persists, keyboard shortcuts work correctly.

---

## Phase 4: User Story 2 - Custom Profile Ordering (Priority: P2)

**Goal**: Enable power users to reorder browser profiles at a granular level and interleave profiles with browsers (e.g., Chrome Work → Safari → Chrome Personal → Firefox). Keyboard shortcuts 1-9 map to flat list of browser+profile combinations.

**Independent Test**:
1. Enable Profile Features in Settings
2. Open Manage Browsers popup
3. Expand Chrome browser (verify Work and Personal profiles shown)
4. Drag to create order: Chrome Work (1), Safari (2), Chrome Personal (3), Firefox (4)
5. Close popup and trigger URL
6. Press key "1": Verify Chrome opens with Work profile
7. Trigger URL again, press key "3": Verify Chrome opens with Personal profile

### Implementation for User Story 2

- [X] T034 [US2] Extend BrowserOrderEntry to support optional profileName field in BrowserOrderState.swift (already in model)
- [X] T035 [US2] Implement profile expansion/collapse UI in BrowserManagementView.swift (chevron icon, expandable list)
- [X] T036 [US2] Add drag-and-drop support for profile entries in BrowserManagementView.swift (.draggable() on profile cards)
- [X] T037 [US2] Implement moveProfile(dragged:droppedOn:) helper method in BrowserManagementView.swift (updates order with profileName)
- [X] T038 [US2] Update getOrderedBrowsers() in BrowserOrderService to include profile-level entries when profileFeaturesEnabled
- [X] T039 [US2] Update keyboard shortcut mapping in SelectionPanel.swift to handle browser+profile combinations
- [X] T040 [US2] Update ProfileSelectionView.swift to respect custom profile order when displaying profiles
- [X] T041 [US2] Implement handleProfileDeleted() in BrowserOrderService (removes profile entry, renumbers)
- [X] T042 [US2] Add profile order validation in BrowserOrderService (orphaned profiles removed on load)

### Localization for User Story 2

- [X] T043 [P] [US2] Add profile ordering localization keys to Proxly/en.lproj/Localizable.strings (expand_profiles, collapse_profiles)
- [X] T044 [P] [US2] Add German profile translations to Proxly/de.lproj/Localizable.strings
- [X] T045 [P] [US2] Add Spanish profile translations to Proxly/es.lproj/Localizable.strings
- [X] T046 [P] [US2] Add French profile translations to Proxly/fr.lproj/Localizable.strings
- [X] T047 [P] [US2] Add Polish profile translations to Proxly/pl.lproj/Localizable.strings
- [X] T048 [P] [US2] Add Dutch profile translations to Proxly/nl.lproj/Localizable.strings
- [X] T049 [P] [US2] Add Italian profile translations to Proxly/it.lproj/Localizable.strings

### Accessibility for User Story 2

- [X] T050 [US2] Add .accessibilityLabel() to profile cards with position info in BrowserManagementView.swift
- [X] T051 [US2] Add .accessibilityAdjustableAction() to profile cards for keyboard-based reordering in BrowserManagementView.swift
- [X] T052 [US2] Update .accessibilityHint() to explain profile interleaving in BrowserManagementView.swift

### Tests for User Story 2

- [X] T053 [P] [US2] Add profile ordering tests to BrowserOrderServiceTests.swift (interleaving, profile deletion)
- [X] T054 [P] [US2] Add profile validation tests to BrowserOrderStateTests.swift (orphaned profile removal)

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently. Users can reorder browsers at browser level (US1) or profile level (US2), with full interleaving support.

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [X] T055 [P] Add comprehensive inline documentation to BrowserOrderService.swift
- [X] T056 [P] Add comprehensive inline documentation to BrowserOrderState.swift
- [X] T057 Run quickstart.md validation scenarios (all 7 scenarios from quickstart.md)
- [X] T058 Performance validation: Verify reordering UI updates < 100ms (manual testing with Xcode Instruments)
- [X] T059 Performance validation: Verify persistence operations < 5ms (run BrowserOrderServiceTests performance tests)
- [X] T060 Verify all 7 languages display correctly in Manage Browsers popup (manual testing)
- [X] T061 VoiceOver validation: Test all accessibility scenarios from quickstart.md
- [X] T062 [P] Code review: Verify all force-unwraps avoided (grep for `!` in new files)
- [X] T063 [P] Code review: Verify proper error handling in persistence layer
- [X] T064 Final integration test: Verify ordering works across all UI surfaces (SelectionPanel, ProfileSelectionView, RuleEditView)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3+)**: All depend on Foundational phase completion
  - User Story 1 (P1): Can start after Foundational - No dependencies on other stories
  - User Story 2 (P2): Can start after Foundational - Builds on US1 but independently testable
- **Polish (Phase 5)**: Depends on User Story 1 at minimum (MVP), ideally both stories

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
  - MVP scope: Complete this story for initial release
  - Fully functional without User Story 2
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - Extends US1 but independently testable
  - Optional enhancement: Can be shipped separately from US1
  - Requires profileFeaturesEnabled setting to activate

### Within Each User Story

**User Story 1**:
1. T011-T020: Implementation tasks (T012-T015 can run in parallel, then T016-T020 sequential)
2. T021-T027: Localization tasks (all parallel)
3. T028-T030: Accessibility tasks (sequential after implementation)
4. T031-T033: Tests (all parallel)

**User Story 2**:
1. T034-T042: Implementation tasks (T035-T037 sequential, T038-T042 sequential)
2. T043-T049: Localization tasks (all parallel)
3. T050-T052: Accessibility tasks (sequential after implementation)
4. T053-T054: Tests (all parallel)

### Parallel Opportunities

- **Phase 1**: All tasks (T001, T002, T003) can run in parallel
- **Phase 2**: T004-T007 sequential (service methods depend on each other), but T008-T009 can run in parallel with T004-T007
- **User Story 1**:
  - T011 + T012 can run in parallel (different files)
  - T021-T027 all parallel (different language files)
  - T031-T033 all parallel (different test files)
- **User Story 2**:
  - T043-T049 all parallel (different language files)
  - T053-T054 all parallel (different test files)
- **Polish Phase**:
  - T055, T056, T062, T063 all parallel (documentation and code review)

---

## Parallel Example: User Story 1

```bash
# Launch all implementation tasks that can run in parallel:
Task T011: "Add browserOrderState property to Proxly/Models/AppSettings.swift"
Task T012: "Extend BrowserManagementView.swift with drag-and-drop modifiers"

# Launch all localization tasks together:
Task T021: "Add localization keys to Proxly/en.lproj/Localizable.strings"
Task T022: "Add German translations to Proxly/de.lproj/Localizable.strings"
Task T023: "Add Spanish translations to Proxly/es.lproj/Localizable.strings"
Task T024: "Add French translations to Proxly/fr.lproj/Localizable.strings"
Task T025: "Add Polish translations to Proxly/pl.lproj/Localizable.strings"
Task T026: "Add Dutch translations to Proxly/nl.lproj/Localizable.strings"
Task T027: "Add Italian translations to Proxly/it.lproj/Localizable.strings"

# Launch all test tasks together:
Task T031: "Create BrowserOrderStateTests.swift in ProxlyTests/"
Task T032: "Create BrowserOrderServiceTests.swift in ProxlyTests/"
Task T033: "Add browser order migration test to ProxlyTests/MigrationManagerTests.swift"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T003) - ~30 minutes
2. Complete Phase 2: Foundational (T004-T010) - ~2 hours (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1 (T011-T033) - ~6-8 hours
   - Implementation: T011-T020
   - Localization: T021-T027
   - Accessibility: T028-T030
   - Tests: T031-T033
4. **STOP and VALIDATE**: Test User Story 1 independently using quickstart.md Scenarios 1-4
5. Deploy/demo if ready

**Estimated total for MVP**: 8-10 hours of focused development

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready (~2.5 hours)
2. Add User Story 1 → Test independently → Deploy/Demo (MVP! +6-8 hours)
3. Add User Story 2 → Test independently → Deploy/Demo (+4-6 hours)
4. Polish Phase → Final validation → Production release (+2-3 hours)

**Total estimated time**: 14-20 hours for complete feature

### Parallel Team Strategy

With multiple developers:

1. **Team completes Setup + Foundational together** (2.5 hours)
2. Once Foundational is done:
   - **Developer A**: User Story 1 implementation (T011-T020)
   - **Developer B**: User Story 1 localization (T021-T027)
   - **Developer C**: User Story 2 implementation (T034-T042)
3. Stories complete and integrate independently

**Parallel completion time**: ~4-5 hours with 3 developers (vs 14-20 hours solo)

---

## Task Summary

- **Total Tasks**: 64
- **Phase 1 (Setup)**: 3 tasks
- **Phase 2 (Foundational)**: 7 tasks (BLOCKING)
- **Phase 3 (User Story 1)**: 23 tasks
  - Implementation: 10 tasks
  - Localization: 7 tasks
  - Accessibility: 3 tasks
  - Tests: 3 tasks
- **Phase 4 (User Story 2)**: 21 tasks
  - Implementation: 9 tasks
  - Localization: 7 tasks
  - Accessibility: 3 tasks
  - Tests: 2 tasks
- **Phase 5 (Polish)**: 10 tasks

### Parallel Opportunities Identified

- **Phase 1**: 3 tasks can run in parallel (T001, T002, T003)
- **Phase 2**: 2 parallel groups (T004-T007 sequential, T008-T009 parallel with group 1)
- **User Story 1**: 9 parallel tasks (T011+T012, T021-T027 all parallel, T031-T033 all parallel)
- **User Story 2**: 9 parallel tasks (T043-T049 all parallel, T053-T054 parallel)
- **Polish**: 4 parallel tasks (T055, T056, T062, T063)

**Total parallel opportunities**: 27 tasks (42% of all tasks)

### Independent Test Criteria

**User Story 1**:
- ✅ Drag-and-drop reordering works in Manage Browsers popup
- ✅ Custom order persists across app restarts
- ✅ Keyboard shortcuts 1-9 map to custom order
- ✅ "Refresh Browser List" preserves custom order
- ✅ Reset button restores alphabetical order

**User Story 2**:
- ✅ Profiles can be expanded under browsers
- ✅ Profiles can be dragged to interleave with browsers
- ✅ Keyboard shortcuts map to browser+profile combinations
- ✅ Profile order persists when Profile Features enabled

### Suggested MVP Scope

**MVP = User Story 1 ONLY**

This delivers immediate value:
- ✅ Solves core problem: consistent keyboard shortcuts
- ✅ Order persists across refresh operations
- ✅ Fully functional without profile complexity
- ✅ Can be validated independently
- ✅ Estimated 8-10 hours of development

**User Story 2 can ship as enhancement** in subsequent release once MVP validated.

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Avoid: vague tasks, same file conflicts, cross-story dependencies that break independence
- Manual testing scenarios detailed in quickstart.md (7 comprehensive scenarios)
- All localization follows existing Proxly pattern (7 languages: EN, DE, ES, FR, PL, NL, IT)
- Accessibility follows macOS Human Interface Guidelines (VoiceOver, keyboard navigation)
