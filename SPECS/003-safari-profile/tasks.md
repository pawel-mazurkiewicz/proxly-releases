# Tasks: Safari Profile Support

**Input**: Design documents from `/specs/003-safari-profile/`
**Prerequisites**: plan.md, spec.md, data-model.md, contracts/

**Tests**: Tests are included based on Proxly Constitution testing requirements (unit tests for stateless logic, integration tests for multi-component flows).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- Include exact file paths in descriptions

## Path Conventions

Proxly uses Swift/Xcode project structure:
- **Models**: `Proxly/Models/`
- **Services**: `Proxly/Services/`
- **UI**: `Proxly/UI/`
- **Tests**: `ProxlyTests/`
- **Localization**: `Proxly/[lang].lproj/Localizable.strings`

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and Safari profile data structures

- [X] T001 Create SafariProfile model struct in `Proxly/Models/SafariProfile.swift` with Codable, Identifiable, fields: id (UUID), name (String), order (Int), isOrphaned (Bool), createdAt (Date), modifiedAt (Date)
- [X] T002 Extend AppSettings model in `Proxly/Models/AppSettings.swift` to add `safariProfiles: [SafariProfile]` field with default empty array
- [X] T003 [P] Extend BrowserProfile struct in `Proxly/Models/BrowserProfile.swift` to add optional `safariProfileId: UUID?` field for Safari profile references

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core Safari profile management service that MUST be complete before ANY user story UI can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T004 Create SafariProfileManager service class in `Proxly/Services/SafariProfileManager.swift` as @MainActor ObservableObject singleton with @Published profiles property
- [X] T005 [P] Implement profile name validation in SafariProfileManager: `validateName(_:)` returning ValidationResult (empty check, 50-char limit, em dash U+2014 detection)
- [X] T006 [P] Implement profile order validation in SafariProfileManager: `validateOrder(_:excludingId:)` returning ValidationResult (1-9 range, active profile duplicate check, excluding specific ID)
- [X] T007 [P] Implement active profile count validation in SafariProfileManager: `validateActiveProfileCount()` returning ValidationResult (max 9 active profiles)
- [X] T008 Implement `createProfile(name:order:)` in SafariProfileManager with full validation chain, create SafariProfile with isOrphaned=false, append to AppSettings, save via PersistenceManager, trigger CloudKit sync
- [X] T009 Implement `getProfile(id:)` in SafariProfileManager returning optional SafariProfile for ID lookup (active or orphaned)
- [X] T010 [P] Implement `getActiveProfiles()` in SafariProfileManager returning filtered array where isOrphaned==false, sorted by order
- [X] T011 [P] Implement `getOrphanedProfiles()` in SafariProfileManager returning filtered array where isOrphaned==true, sorted by modifiedAt descending
- [X] T012 Implement `editProfile(id:newName:newOrder:)` in SafariProfileManager: lookup original, validate new values, mark original isOrphaned=true, create new profile with isOrphaned=false, save, sync
- [X] T013 Implement `deleteProfile(id:)` in SafariProfileManager: remove from AppSettings.safariProfiles array, save, sync, return true if found
- [X] T014 Extend PersistenceManager in `Proxly/Services/PersistenceManager.swift` to persist SafariProfile array in AppSettings with JSON encoding (backward compatible with default empty array)

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Configure Safari Profiles (Priority: P1) 🎯 MVP

**Goal**: Enable users to manually configure Safari profiles with names and keyboard shortcut order numbers, validate inputs, support orphaned profile lifecycle on edits, persist across restarts

**Independent Test**: Open Proxly settings → Navigate to browser configuration → Add Safari profiles "Work" (order 1) and "Personal" (order 2) → Verify saved and displayed → Edit "Work" to "Work2" → Verify original orphaned and new created → Delete "Personal" → Verify removed → Restart Proxly → Verify "Work2" active and "Work" orphaned both persist

### Tests for User Story 1

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T015 [P] [US1] Create SafariProfileManagerTests in `ProxlyTests/SafariProfileManagerTests.swift` with test cases for profile CRUD operations
- [ ] T016 [P] [US1] Write test: create profile with valid name/order → success, profile added to storage
- [ ] T017 [P] [US1] Write test: create profile with empty name → throws SafariProfileError.invalidName
- [ ] T018 [P] [US1] Write test: create profile with em dash in name → throws SafariProfileError.invalidName with message about em dash
- [ ] T019 [P] [US1] Write test: create profile with duplicate active order → throws SafariProfileError.invalidOrder
- [ ] T020 [P] [US1] Write test: create 10th active profile → throws SafariProfileError.activeProfileLimitReached
- [ ] T021 [P] [US1] Write test: edit profile → original marked isOrphaned=true, new profile created with isOrphaned=false
- [ ] T022 [P] [US1] Write test: edit with duplicate order to orphaned profile → success (orphaned excluded from validation)
- [ ] T023 [P] [US1] Write test: delete profile → removed from storage, returns true
- [ ] T024 [P] [US1] Write test: delete non-existent profile → returns false
- [ ] T025 [P] [US1] Create SafariProfileValidationTests in `ProxlyTests/SafariProfileValidationTests.swift` for validation logic
- [ ] T026 [P] [US1] Write test: validateName with 51 chars → failure
- [ ] T027 [P] [US1] Write test: validateName with en dash (U+2013) → success (allowed)
- [ ] T028 [P] [US1] Write test: validateName with hyphen → success (allowed)
- [ ] T029 [P] [US1] Write test: validateOrder with value 0 → failure
- [ ] T030 [P] [US1] Write test: validateOrder with value 10 → failure
- [ ] T031 [P] [US1] Create OrphanedProfileTests in `ProxlyTests/OrphanedProfileTests.swift` for orphaned lifecycle
- [ ] T032 [P] [US1] Write test: orphaned profiles not counted toward 9-profile limit
- [ ] T033 [P] [US1] Write test: orphaned profiles not checked for duplicate orders
- [ ] T034 [P] [US1] Write test: getActiveProfiles excludes orphaned profiles
- [ ] T035 [P] [US1] Write test: getOrphanedProfiles excludes active profiles

### Implementation for User Story 1

- [X] T036 [US1] Create SafariProfileConfigView in `Proxly/UI/SafariProfileConfigView.swift` as SwiftUI view with @StateObject SafariProfileManager
- [X] T037 [US1] Add profile list display in SafariProfileConfigView showing active profiles (sorted by order) with name and order columns
- [X] T038 [US1] Add orphaned profiles section in SafariProfileConfigView showing orphaned profiles with visual distinction (grayed out, "(legacy)" label)
- [X] T039 [US1] Implement add profile form in SafariProfileConfigView: TextField for name (max 50 chars), Picker for order (1-9), Save button
- [X] T040 [US1] Add real-time validation in SafariProfileConfigView: display inline error for empty name, em dash detection, duplicate order, 9-profile limit
- [X] T041 [US1] Implement edit profile flow in SafariProfileConfigView: pre-fill form with existing values, call SafariProfileManager.editProfile on save
- [X] T042 [US1] Implement delete profile action in SafariProfileConfigView: confirmation alert if profile referenced by rules (use SafariProfileManager.getRuleReferenceCount), call deleteProfile
- [X] T043 [US1] Integrate SafariProfileConfigView into BrowserSettingsView in `Proxly/UI/Settings/BrowserSettingsView.swift` as new section "Safari Profiles"
- [X] T044 [US1] Add VoiceOver accessibility labels in SafariProfileConfigView: profile name input, order selector, save button, delete button, orphaned profile indicators
- [X] T045 [US1] Add localized strings for Safari profile UI in `Proxly/en.lproj/Localizable.strings`: "safari_profile_title", "add_profile", "profile_name_placeholder", "profile_order", "save_profile", "delete_profile", "orphaned_label", "em_dash_error", "duplicate_order_error", "profile_limit_error"
- [X] T046 [P] [US1] Copy localized strings to `Proxly/de.lproj/Localizable.strings` (German)
- [X] T047 [P] [US1] Copy localized strings to `Proxly/es.lproj/Localizable.strings` (Spanish)
- [X] T048 [P] [US1] Copy localized strings to `Proxly/fr.lproj/Localizable.strings` (French)
- [X] T049 [P] [US1] Copy localized strings to `Proxly/pl.lproj/Localizable.strings` (Polish)
- [X] T050 [P] [US1] Copy localized strings to `Proxly/nl.lproj/Localizable.strings` (Dutch)
- [X] T051 [P] [US1] Copy localized strings to `Proxly/it.lproj/Localizable.strings` (Italian)

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently. Users can configure Safari profiles, see them persist across restarts, and see orphaned profiles when editing.

---

## Phase 4: User Story 2 - Select Safari Profile in Rules (Priority: P2)

**Goal**: Enable users to select specific Safari profiles when creating or editing URL routing rules, display active and orphaned profiles with visual distinction, persist profile selection in rules

**Independent Test**: Create new domain rule for `github.com` → Select Safari as browser → Verify profile selector shows active profiles first (sorted by order), then orphaned profiles (grayed out) → Select "Work" profile → Save rule → Re-open rule editor → Verify "Work" profile selected → Edit "Work" profile to "Work2" → Re-open rule → Verify original "Work" (now orphaned, grayed out) still selected

### Tests for User Story 2

- [ ] T052 [P] [US2] Create integration test in `ProxlyTests/RuleProfileIntegrationTests.swift` for rule creation with Safari profile
- [ ] T053 [P] [US2] Write test: create rule with Safari profile → save → load → profile reference preserved in BrowserTarget.safariProfileId
- [ ] T054 [P] [US2] Write test: edit profile → rule still references original (orphaned) profile by ID
- [ ] T055 [P] [US2] Write test: delete orphaned profile → rule remains valid but routing may fall back
- [ ] T056 [P] [US2] Write test: profile selector shows active profiles before orphaned profiles
- [ ] T057 [P] [US2] Write test: profile selector includes "Any Profile" option

### Implementation for User Story 2

- [X] T058 [US2] Extend BrowserProfileSelector in `Proxly/UI/BrowserProfileSelector.swift` to detect Safari browser selection and show Safari-specific profile picker
- [X] T059 [US2] Add Safari profile list generation in BrowserProfileSelector: fetch active profiles via SafariProfileManager.getActiveProfiles(), fetch orphaned via getOrphanedProfiles()
- [X] T060 [US2] Implement visual distinction for orphaned profiles in BrowserProfileSelector: gray out orphaned profiles, add "(legacy)" suffix to display name
- [X] T061 [US2] Add "Any Profile" option at top of Safari profile selector (maps to nil safariProfileId in BrowserTarget)
- [X] T062 [US2] Bind selected Safari profile to Rule.browserTarget.safariProfileId field (UUID?)
- [X] T063 [US2] Display selected profile name in rule list view when Safari + specific profile selected
- [X] T064 [US2] Add prompt in profile selector when no Safari profiles configured: "Configure Safari profiles in Settings → Browsers"
- [X] T065 [US2] Add VoiceOver accessibility labels in BrowserProfileSelector for Safari profile picker: active profile option, orphaned profile option, "Any Profile" option
- [X] T066 [US2] Add localized strings for profile selector UI in `Proxly/en.lproj/Localizable.strings`: "any_profile", "select_safari_profile", "configure_profiles_prompt", "legacy_profile_suffix"
- [X] T067 [P] [US2] Copy profile selector localized strings to `Proxly/de.lproj/Localizable.strings` (German)
- [X] T068 [P] [US2] Copy profile selector localized strings to `Proxly/es.lproj/Localizable.strings` (Spanish)
- [X] T069 [P] [US2] Copy profile selector localized strings to `Proxly/fr.lproj/Localizable.strings` (French)
- [X] T070 [P] [US2] Copy profile selector localized strings to `Proxly/pl.lproj/Localizable.strings` (Polish)
- [X] T071 [P] [US2] Copy profile selector localized strings to `Proxly/nl.lproj/Localizable.strings` (Dutch)
- [X] T072 [P] [US2] Copy profile selector localized strings to `Proxly/it.lproj/Localizable.strings` (Italian)

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently. Users can configure profiles and select them in rules.

---

## Phase 5: User Story 3 - Route to Existing Safari Profile Window (Priority: P3)

**Goal**: When URL matches rule with Safari profile target and matching profile window exists, bring window to front, create new tab, navigate to URL within 2 seconds

**Independent Test**: Configure "Work" profile (order 1) → Create rule `apple.com` → Safari → "Work" profile → Open Safari with "Work" profile window (Option+Cmd+Shift+1) → Navigate to any page → Trigger URL `apple.com` → Verify "Work" window comes to front → Verify new tab opens with `apple.com` → Verify completes within 2 seconds

### Tests for User Story 3

- [ ] T073 [P] [US3] Create SafariWindowDetectorTests in `ProxlyTests/SafariWindowDetectorTests.swift` for pattern matching logic
- [ ] T074 [P] [US3] Write test: extractProfileName("Work — Apple") → returns "Work"
- [ ] T075 [P] [US3] Write test: extractProfileName("Personal — GitHub — Pull Requests") → returns "Personal"
- [ ] T076 [P] [US3] Write test: extractProfileName("Safari") → returns nil
- [ ] T077 [P] [US3] Write test: extractProfileName("Work — ") → returns "Work" (empty page title)
- [ ] T078 [P] [US3] Write test: extractProfileName(" — Page") → returns nil (empty profile name)
- [ ] T079 [P] [US3] Write test: matchesProfilePattern("Work — Page") → returns true
- [ ] T080 [P] [US3] Write test: matchesProfilePattern("Safari") → returns false
- [ ] T081 [P] [US3] Create integration test in `ProxlyTests/WindowDetectionIntegrationTests.swift` (requires Accessibility permissions in test environment)
- [ ] T082 [P] [US3] Write integration test: enumerate Safari windows → verify all windows returned with titles
- [ ] T083 [P] [US3] Write integration test: find window for profile "Work" → verify correct window returned

### Implementation for User Story 3

- [ ] T084 [US3] Create SafariWindowDetector service in `Proxly/Services/SafariWindowDetector.swift` as @MainActor singleton
- [ ] T085 [US3] Implement `checkAccessibilityPermissions()` in SafariWindowDetector using AXIsProcessTrusted() returning Bool
- [ ] T086 [US3] Implement `extractProfileName(from:)` in SafariWindowDetector: split title on first " — " (space + U+2014 + space), return substring before em dash or nil
- [ ] T087 [US3] Implement `matchesProfilePattern(_:)` in SafariWindowDetector: check if title contains " — " pattern
- [ ] T088 [US3] Create SafariWindowInfo struct in SafariWindowDetector.swift with fields: windowElement (AXUIElement), title (String), profileName (String?), pageTitle (String?), pid (pid_t)
- [ ] T089 [US3] Implement `enumerateWindows()` async in SafariWindowDetector: check Safari running via NSWorkspace, check permissions, get Safari PID, create AXUIElement, copy windows attribute, read titles, extract profile names, return [SafariWindowInfo]
- [ ] T090 [US3] Implement `findWindow(forProfile:)` async in SafariWindowDetector: call enumerateWindows(), filter by profileName match, return first match or nil
- [ ] T091 [US3] Implement `bringWindowToFront(_:)` async in SafariWindowDetector: validate windowElement, set kAXMainAttribute to true, optionally activate Safari app
- [ ] T092 [US3] Implement `handleAccessibilityError(_:)` async in SafariWindowDetector: create NSAlert with "Accessibility Permissions Required" message, add "Open System Settings" button, open x-apple.systempreferences URL on click
- [ ] T093 [US3] Extend BrowserLauncher in `Proxly/Services/BrowserLauncher.swift` to detect Safari with profile target: check if browserTarget has safariProfileId, lookup profile via SafariProfileManager
- [ ] T094 [US3] Add Safari profile routing logic in BrowserLauncher: if profile found, call SafariWindowDetector.findWindow(forProfile:), if window exists call bringWindowToFront, then create tab via AppleScript/keyboard shortcut (Cmd+T), navigate URL
- [ ] T095 [US3] Add Accessibility permission error handling in BrowserLauncher Safari routing: catch SafariWindowDetectorError.accessibilityPermissionDenied, call handleAccessibilityError, fall back to default Safari opening
- [ ] T096 [US3] Add performance logging in BrowserLauncher for Safari profile routing: log start time, window detection time, total routing time, verify <2 seconds target
- [ ] T097 [US3] Add fallback logic in BrowserLauncher: if findWindow returns nil (no matching window), fall back to new window creation (handled in User Story 4) or default Safari
- [ ] T098 [US3] Add localized strings for Accessibility errors in `Proxly/en.lproj/Localizable.strings`: "accessibility_required_title", "accessibility_required_message", "open_system_settings", "accessibility_denied_fallback"
- [ ] T099 [P] [US3] Copy Accessibility error strings to `Proxly/de.lproj/Localizable.strings` (German)
- [ ] T100 [P] [US3] Copy Accessibility error strings to `Proxly/es.lproj/Localizable.strings` (Spanish)
- [ ] T101 [P] [US3] Copy Accessibility error strings to `Proxly/fr.lproj/Localizable.strings` (French)
- [ ] T102 [P] [US3] Copy Accessibility error strings to `Proxly/pl.lproj/Localizable.strings` (Polish)
- [ ] T103 [P] [US3] Copy Accessibility error strings to `Proxly/nl.lproj/Localizable.strings` (Dutch)
- [ ] T104 [P] [US3] Copy Accessibility error strings to `Proxly/it.lproj/Localizable.strings` (Italian)

**Checkpoint**: All user stories 1-3 should now be independently functional. Users can configure profiles, select in rules, and route to existing profile windows.

---

## Phase 6: User Story 4 - Create New Safari Profile Window (Priority: P4)

**Goal**: When URL matches rule with Safari profile target but no matching window exists, launch Safari (if needed), trigger keyboard shortcut to create profile window, wait up to 2 seconds for window creation, navigate to URL, or fall back to default Safari on timeout

**Independent Test**: Quit Safari completely → Create rule `github.com` → Safari → "Work" profile → Trigger URL `github.com` → Verify Safari launches → Verify keyboard shortcut Option+Cmd+Shift+1 triggered → Verify "Work" profile window created → Verify URL opens in new window → Verify completes within 4 seconds

### Tests for User Story 4

- [ ] T105 [P] [US4] Create SafariProfileLauncherTests in `ProxlyTests/SafariProfileLauncherTests.swift` for keyboard shortcut generation
- [ ] T106 [P] [US4] Write test: keyboard shortcut key code mapping for order 1 → key code 18
- [ ] T107 [P] [US4] Write test: keyboard shortcut key code mapping for order 9 → key code 25
- [ ] T108 [P] [US4] Write test: keyboard shortcut flags include maskShift, maskCommand, maskAlternate
- [ ] T109 [P] [US4] Create integration test in `ProxlyTests/WindowCreationIntegrationTests.swift` (requires Safari and manual verification)
- [ ] T110 [P] [US4] Write integration test: create profile window → wait for detection → verify window appears with correct profile name

### Implementation for User Story 4

- [ ] T111 [US4] Create SafariProfileLauncher service in `Proxly/Services/SafariProfileLauncher.swift` as class with static methods (no state)
- [ ] T112 [US4] Define key code mapping in SafariProfileLauncher: static dictionary [Int: CGKeyCode] mapping order 1-9 to US keyboard layout key codes (1:18, 2:19, 3:20, 4:21, 5:23, 6:22, 7:26, 8:28, 9:25)
- [ ] T113 [US4] Implement `triggerProfileShortcut(order:)` in SafariProfileLauncher: create CGEvent with key code and flags (maskShift, maskCommand, maskAlternate), post key down and key up events
- [ ] T114 [US4] Implement `launchSafari()` async in SafariProfileLauncher: use NSWorkspace.shared.launchApplication with Safari bundle ID, wait 200ms for launch, return success/failure
- [ ] T115 [US4] Implement `createProfileWindow(order:)` async in SafariProfileLauncher: launch Safari if not running, wait for ready, call triggerProfileShortcut, wait up to 2 seconds polling for window creation via SafariWindowDetector.enumerateWindows
- [ ] T116 [US4] Implement timeout handling in createProfileWindow: if 2 seconds elapsed without matching window detected, return failure result
- [ ] T117 [US4] Implement fallback logic in createProfileWindow: on failure, return error indicating fallback to default Safari needed
- [ ] T118 [US4] Extend BrowserLauncher Safari routing to call SafariProfileLauncher.createProfileWindow when findWindow returns nil (no existing window)
- [ ] T119 [US4] Add retry logic in BrowserLauncher: if window creation succeeds, verify window exists via SafariWindowDetector.findWindow, then proceed with bringWindowToFront and URL navigation
- [ ] T120 [US4] Add fallback in BrowserLauncher: if createProfileWindow fails or times out, open URL in default Safari without profile specification via NSWorkspace
- [ ] T121 [US4] Add performance logging in BrowserLauncher for new window creation: log keyboard shortcut trigger time, window detection polling time, total time, verify <4 seconds target
- [ ] T122 [US4] Add localized strings for window creation errors in `Proxly/en.lproj/Localizable.strings`: "profile_window_creation_failed", "profile_window_timeout", "falling_back_to_default_safari"
- [ ] T123 [P] [US4] Copy window creation error strings to `Proxly/de.lproj/Localizable.strings` (German)
- [ ] T124 [P] [US4] Copy window creation error strings to `Proxly/es.lproj/Localizable.strings` (Spanish)
- [ ] T125 [P] [US4] Copy window creation error strings to `Proxly/fr.lproj/Localizable.strings` (French)
- [ ] T126 [P] [US4] Copy window creation error strings to `Proxly/pl.lproj/Localizable.strings` (Polish)
- [ ] T127 [P] [US4] Copy window creation error strings to `Proxly/nl.lproj/Localizable.strings` (Dutch)
- [ ] T128 [P] [US4] Copy window creation error strings to `Proxly/it.lproj/Localizable.strings` (Italian)

**Checkpoint**: All user stories should now be independently functional. Complete end-to-end Safari profile routing works for both existing and new profile windows.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories and ProxlyHelper migration preparation

- [ ] T129 [P] Create SafariProfileHelper in `ProxlyHelper/SafariProfileHelper.swift` with Accessibility API operations (window detection, keyboard shortcuts) for Mac App Store version
- [ ] T130 [P] Extend HelperURLRouter in `ProxlyHelper/HelperURLRouter.swift` to handle Safari profile commands via proxly-helper:// URL scheme
- [ ] T131 [P] Add IPC communication pattern between main app and ProxlyHelper for Safari profile operations: define URL scheme format, handle responses via proxly://helper URLs
- [ ] T132 Code review: Verify Safari-specific logic isolated in dedicated files (SafariProfile.swift, SafariProfileManager.swift, SafariWindowDetector.swift, SafariProfileLauncher.swift)
- [ ] T133 Code review: Verify no state stored in window detection or keyboard shortcut logic (stateless functions only)
- [ ] T134 Code review: Verify all Accessibility API calls have error handling with user-friendly fallbacks
- [ ] T135 Code review: Verify VoiceOver labels added to all Safari profile UI elements
- [ ] T136 Code review: Verify all user-facing strings localized in 7 languages
- [ ] T137 Performance testing: Measure window enumeration overhead with 5-10 Safari windows, verify <50ms
- [ ] T138 Performance testing: Measure end-to-end existing window routing, verify <2 seconds p95
- [ ] T139 Performance testing: Measure end-to-end new window creation, verify <4 seconds p90
- [ ] T140 Manual testing: Test Dark mode and Light mode appearance for profile configuration UI and profile selector
- [ ] T141 Manual testing: Test keyboard navigation through profile configuration form
- [ ] T142 Manual testing: Test VoiceOver announcements for profile actions (add, edit, delete, select in rule)
- [ ] T143 Manual testing: Test with multiple Safari windows for different profiles, verify correct window targeting
- [ ] T144 Manual testing: Test fallback scenarios (Accessibility denied, Safari not running, window creation timeout)
- [ ] T145 Update quickstart.md with Safari profile testing instructions if needed
- [ ] T146 Update CLAUDE.md if Safari profile patterns introduce new architecture conventions

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup (Phase 1) completion - BLOCKS all user stories
- **User Stories (Phase 3-6)**: All depend on Foundational (Phase 2) completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3 → P4)
- **Polish (Phase 7)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) AND User Story 1 completion (needs SafariProfileManager working for profile selector data)
- **User Story 3 (P3)**: Can start after Foundational (Phase 2) AND User Story 2 completion (needs BrowserProfileSelector integration and Rule.browserTarget.safariProfileId field)
- **User Story 4 (P4)**: Can start after User Story 3 completion (extends BrowserLauncher Safari routing with window creation fallback)

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Models before services (Phase 1/2 handles this)
- Services before UI (SafariProfileManager before SafariProfileConfigView)
- Core implementation before integration (window detection before BrowserLauncher integration)
- Story complete before moving to next priority

### Parallel Opportunities

- **Setup Phase**: T001, T002, T003 can run in parallel (different files)
- **Foundational Phase**: T005, T006, T007 validation methods can run in parallel; T009, T010, T011 getter methods can run in parallel
- **User Story 1 Tests**: T015-T035 all test tasks can run in parallel (different test files)
- **User Story 1 Localization**: T046-T051 localization tasks can run in parallel (different files)
- **User Story 2 Tests**: T052-T057 test tasks can run in parallel
- **User Story 2 Localization**: T067-T072 localization tasks can run in parallel
- **User Story 3 Tests**: T073-T083 test tasks can run in parallel
- **User Story 3 Localization**: T099-T104 localization tasks can run in parallel
- **User Story 4 Tests**: T105-T110 test tasks can run in parallel
- **User Story 4 Localization**: T123-T128 localization tasks can run in parallel
- **Polish Phase**: T129, T130, T131 ProxlyHelper tasks can run in parallel; T137, T138, T139 performance tests can run in parallel; T140-T144 manual tests can run in parallel

---

## Parallel Example: User Story 1

```bash
# Launch all tests for User Story 1 together (write tests first):
Task T015: "Create SafariProfileManagerTests in ProxlyTests/SafariProfileManagerTests.swift"
Task T016: "Write test: create profile with valid name/order → success"
Task T017: "Write test: create profile with empty name → throws error"
# ... (all T015-T035 tests in parallel)

# Launch all localization tasks for User Story 1 together:
Task T046: "Copy localized strings to Proxly/de.lproj/Localizable.strings"
Task T047: "Copy localized strings to Proxly/es.lproj/Localizable.strings"
Task T048: "Copy localized strings to Proxly/fr.lproj/Localizable.strings"
# ... (all T046-T051 localizations in parallel)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T003)
2. Complete Phase 2: Foundational (T004-T014) - CRITICAL - blocks all stories
3. Complete Phase 3: User Story 1 (T015-T051)
4. **STOP and VALIDATE**: Test User Story 1 independently
   - Configure profiles in Proxly settings
   - Verify validation (em dash, duplicate orders, 9-profile limit)
   - Edit profile and verify orphaned lifecycle
   - Restart app and verify persistence
5. Deploy/demo if ready (users can configure profiles, foundation for routing)

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Deploy/Demo (MVP! Users can configure profiles)
3. Add User Story 2 → Test independently → Deploy/Demo (Users can select profiles in rules)
4. Add User Story 3 → Test independently → Deploy/Demo (Existing window routing works)
5. Add User Story 4 → Test independently → Deploy/Demo (New window creation works, complete feature)
6. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers (NOT RECOMMENDED for this feature due to tight service coupling):

1. Team completes Setup + Foundational together (T001-T014)
2. Once Foundational is done:
   - Developer A: User Story 1 (T015-T051)
   - Wait for US1 complete before starting US2
   - Developer A continues: User Story 2 (T052-T072)
   - Developer B can start: User Story 3 (T073-T104) in parallel with US2
   - Wait for US3 complete before starting US4
   - Developer A or B: User Story 4 (T105-T128)
3. Stories integrate via shared SafariProfileManager and BrowserLauncher

**RECOMMENDATION**: Sequential implementation (P1 → P2 → P3 → P4) due to dependencies.

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify tests fail before implementing (TDD approach)
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- **Manual testing required**: Accessibility API and keyboard shortcuts difficult to automate, manual testing plan in quickstart.md
- **ProxlyHelper preparation**: Phase 7 includes ProxlyHelper tasks for future Mac App Store migration
- **Localization**: All 7 languages (EN, DE, ES, FR, PL, NL, IT) must be updated per user story
- **Performance targets**: <50ms window enumeration, <2s existing window routing, <4s new window creation
- Avoid: vague tasks, same file conflicts, cross-story dependencies that break independence

---

## Summary

**Total Tasks**: 146 tasks
- Phase 1 (Setup): 3 tasks
- Phase 2 (Foundational): 11 tasks (CRITICAL BLOCKING PHASE)
- Phase 3 (User Story 1 - P1): 36 tasks (21 tests + 15 implementation)
- Phase 4 (User Story 2 - P2): 21 tasks (6 tests + 15 implementation)
- Phase 5 (User Story 3 - P3): 32 tasks (11 tests + 21 implementation)
- Phase 6 (User Story 4 - P4): 24 tasks (6 tests + 18 implementation)
- Phase 7 (Polish): 18 tasks

**Parallel Opportunities**: 76 tasks marked [P] can run in parallel (52%)

**Independent Test Criteria**:
- User Story 1: Configure profiles, verify persistence, test orphaned lifecycle
- User Story 2: Create rules with profile selection, verify persistence
- User Story 3: Route to existing profile window, verify <2s performance
- User Story 4: Create new profile window, verify <4s performance, test fallback

**Suggested MVP Scope**: User Story 1 only (T001-T051, 50 tasks)
- Delivers: Profile configuration UI with full validation and orphaned lifecycle
- Value: Foundation for all routing functionality, users can prepare profiles
- Risk: Low, isolated UI feature with comprehensive tests
