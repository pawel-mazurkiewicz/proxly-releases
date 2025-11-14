# Implementation Plan: Safari Profile Support

**Branch**: `003-safari-profile` | **Date**: 2025-11-04 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/003-safari-profile/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

**Primary Requirement**: Enable Proxly to route URLs to specific Safari profiles based on user-defined rules, using macOS Accessibility API for window detection and keyboard shortcuts for profile window creation, while maintaining an orphaned profile model that preserves backward compatibility with existing rules.

**Technical Approach**: Extend existing BrowserProfile model to support Safari profiles with manual name/order configuration. Implement Safari-specific window detection using Accessibility API to match window titles against the pattern "PROFILENAME — PAGE TITLE". Use keyboard shortcuts (Option+Cmd+Shift+[1-9]) for profile window creation when target window doesn't exist. Implement orphaned profile lifecycle where editing creates new entries while preserving originals for rule references. Integrate with existing BrowserOrderService, BrowserProfileDetector, and BrowserProfileSelector while keeping Safari logic isolated for future ProxlyHelper migration.

## Technical Context

**Language/Version**: Swift 6.0+
**Primary Dependencies**:
- SwiftUI, AppKit (for UI)
- Combine (for reactive state)
- macOS Accessibility API (AXUIElement, AXObserver for window enumeration and title reading)
- AppleScript/NSAppleScript (for Safari tab creation and URL navigation)
- CGEvent (for keyboard shortcut simulation: Option+Cmd+Shift+[1-9], Cmd+T)

**Storage**: UserDefaults.standard (Safari profile configuration persisted with existing rules)
**Testing**: XCTest (unit tests for validation, window detection logic, orphaned profile lifecycle)
**Target Platform**: macOS 14.0+
**Project Type**: Desktop Mac app (menu bar accessory) - extending existing Proxly application
**Performance Goals**:
- 2 seconds for existing profile window routing (p95)
- 4 seconds for new profile window creation (p90)
- Minimal overhead from window enumeration (<50ms)

**Constraints**:
- No official Safari API for profile control (reverse-engineered window title patterns)
- Maximum 9 Safari profiles (keyboard shortcut limitation)
- Window title pattern "PROFILENAME — PAGE TITLE" assumes em dash (U+2014) in English locale
- Accessibility API permissions required from user
- Future Mac App Store version will require moving Accessibility logic to ProxlyHelper

**Scale/Scope**:
- Up to 9 actively configured Safari profiles per user
- Unlimited orphaned profiles (legacy entries)
- Profile configuration syncs via existing CloudKit infrastructure

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

This feature MUST comply with Proxly Constitution v1.0.0. Check all applicable principles:

### I. macOS Native Integration
- [x] Uses native Apple frameworks (SwiftUI for UI, Accessibility API for window detection, CGEvent for keyboard simulation)
- [x] Maintains macOS 14.0+ compatibility (Accessibility API stable since macOS 10.9, enhanced in recent versions)
- [x] Respects accessibility standards (VoiceOver labels for profile configuration UI, keyboard navigation)
- [x] Follows system integration patterns (Accessibility permissions flow, keyboard shortcut conventions)

**Note**: Uses reverse-engineered window title patterns due to lack of official Safari profile API, but this is the only viable approach.

### II. URL Processing Correctness
- [x] Pattern matching logic is stateless (window detection logic will be stateless functions)
- [x] Rule evaluation follows priority order (Safari profile routing integrates with existing URLProcessor priority system)
- [x] Behavior is deterministic (same profile configuration + same Safari windows = same routing result)
- [x] User input validation with helpful error messages (em dash validation, duplicate order detection, 9-profile limit)

**Note**: Adds profile-specific routing after browser selection, doesn't modify core URLProcessingEngine.

### III. Data Persistence & Sync Integrity
- [x] Uses UserDefaults.standard with JSON encoding (Safari profiles stored alongside existing browser profiles)
- [x] CloudKit sync uses last-write-wins conflict resolution (Safari profile config syncs via existing AppSettings)
- [x] Schema migrations maintain backward compatibility (orphaned profile model preserves rule references)
- [x] Sync triggering is debounced/throttled (uses existing PersistenceManager sync patterns)

**Note**: Orphaned profile model is a form of append-only data structure that prevents breaking rule references.

### IV. Graceful Degradation
- [x] Feature works when ProxlyHelper unavailable (direct Accessibility API in standalone version, delegates to helper in Mac App Store version)
- [x] Handles missing permissions gracefully (Accessibility permission errors caught, user shown guidance to System Settings)
- [x] Provides user-friendly error messages (permission errors, profile window creation timeouts, Safari not responding)
- [x] Never crashes on external dependency failure (2-second timeouts, fallback to default Safari opening)

**Note**: Falls back to opening Safari without profile specification when window detection or creation fails.

### V. Privacy & Sandboxing
- [x] Operates within sandbox constraints (Mac App Store version delegates Accessibility operations to ProxlyHelper)
- [x] Minimizes URL logging/storage (only logs routing attempts/failures for debugging, no URL content stored)
- [x] Requests minimal permissions (Accessibility permissions required, clearly explained to user)
- [⚠️] Uses audit tokens for opener detection (not applicable to Safari profile routing, but preserved in overall system)

**Note**: Accessibility API requires extra permissions, but this is unavoidable for Safari profile window detection. Implementation is isolated for clean ProxlyHelper migration path.

### VI. Performance & Responsiveness
- [x] URL processing <50ms p95 (window enumeration via Accessibility API is fast, typically <20ms for 5-10 windows)
- [x] Menu bar interaction <100ms (profile configuration UI uses standard SwiftUI patterns)
- [x] Rule evaluation <10ms for typical rule sets (profile matching adds negligible overhead after browser selection)
- [x] Non-blocking async operations (window detection, keyboard shortcuts, profile window creation all async)
- [⚠️] Regex patterns cached appropriately (not applicable to Safari profile matching, which uses simple string prefix matching)

**Note**: Keyboard shortcut simulation and window creation can take up to 2 seconds, but this is acceptable for cold-start scenarios.

### VII. Localization & Accessibility
- [x] All strings externalized to Localizable.strings (profile configuration UI, validation errors, permission guidance)
- [x] Supports all 7 languages (EN, DE, ES, FR, PL, NL, IT) (new strings added to all localization files)
- [x] VoiceOver labels and hints provided (profile name input, order selector, orphaned profile indicators)
- [x] Keyboard navigation fully supported (profile configuration form, profile selector in rules)
- [x] Sufficient color contrast in all UI (orphaned profiles grayed out but still legible, standard SwiftUI contrast)

**Note**: Window title pattern matching assumes English locale em dash (U+2014). Other locales may fail and fall back to default Safari (documented limitation).

### Testing Requirements
- [x] Unit tests for stateless logic (profile name validation, em dash detection, duplicate order checking, orphaned profile lifecycle)
- [x] Integration tests for multi-component flows (profile configuration → rule creation → URL routing → window detection)
- [x] Manual UI testing in Light and Dark mode (profile configuration UI, orphaned profile visual distinction)
- [x] Migration tests if schema changes (orphaned profile model adds new fields to BrowserProfile, requires migration test)

**Note**: Accessibility API and keyboard shortcut testing difficult to automate. Manual testing plan required.

### Anti-Patterns Avoided
- [x] No state stored in URLProcessingEngine (Safari profile logic in separate service classes)
- [x] No hard-coded browser bundle IDs (Safari bundle ID used from existing BrowserProfileDetector)
- [x] No blocking operations on main thread (all Accessibility API calls async, keyboard shortcuts on background)
- [x] No force-unwraps outside test code (optional chaining for window title parsing, profile lookups)

**Additional Considerations**:
- Accessibility API usage is inherently fragile (relies on window title patterns)
- Keyboard shortcut simulation may conflict with user's custom shortcuts
- Em dash character assumption may break in non-English locales
- 2-second timeout may be too short on slower hardware

## Project Structure

### Documentation (this feature)

```text
specs/003-safari-profile/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
browser-chooser/proxly/
├── Proxly/
│   ├── Models/
│   │   ├── BrowserProfile.swift           (EXTEND: Add Safari profile fields, orphaned state)
│   │   └── SafariProfile.swift            (NEW: Safari-specific profile model)
│   ├── Services/
│   │   ├── BrowserOrderService.swift      (EXTEND: Support Safari profile ordering)
│   │   ├── BrowserProfileDetector.swift   (EXTEND: Manual Safari profile input handling)
│   │   ├── SafariProfileManager.swift     (NEW: Safari profile CRUD, orphaned lifecycle)
│   │   ├── SafariWindowDetector.swift     (NEW: Accessibility API window enumeration and title matching)
│   │   ├── SafariProfileLauncher.swift    (NEW: Profile window creation, keyboard shortcuts, URL navigation)
│   │   ├── BrowserLauncher.swift          (EXTEND: Delegate Safari profile launches to SafariProfileLauncher)
│   │   └── PersistenceManager.swift       (EXTEND: Persist Safari profiles with orphaned state)
│   ├── UI/
│   │   ├── BrowserProfileSelector.swift   (EXTEND: Show active + orphaned Safari profiles with visual distinction)
│   │   ├── SafariProfileConfigView.swift  (NEW: Safari profile configuration UI)
│   │   └── Settings/
│   │       └── BrowserSettingsView.swift  (EXTEND: Add Safari profile configuration section)
│   └── Localizable resources (de, en, es, fr, it, nl, pl)
│       └── Localizable.strings            (EXTEND: Add Safari profile strings)
├── ProxlyHelper/
│   ├── SafariProfileHelper.swift          (NEW: Accessibility API operations for Mac App Store version)
│   └── HelperURLRouter.swift              (EXTEND: Route Safari profile commands)
├── ProxlyTests/
│   ├── SafariProfileManagerTests.swift    (NEW: Profile CRUD, orphaned lifecycle tests)
│   ├── SafariWindowDetectorTests.swift    (NEW: Window title pattern matching tests)
│   ├── SafariProfileValidationTests.swift (NEW: Em dash validation, duplicate order tests)
│   └── OrphanedProfileTests.swift         (NEW: Orphaned profile lifecycle and selector tests)
└── Proxly.xcodeproj/
```

**Structure Decision**:
- Safari-specific logic isolated in dedicated files (SafariProfile.swift, SafariProfileManager.swift, SafariWindowDetector.swift, SafariProfileLauncher.swift)
- Extends existing components minimally (BrowserProfile, BrowserLauncher, BrowserProfileSelector)
- Clean separation for future ProxlyHelper migration (all Accessibility/keyboard logic in dedicated classes)
- Follows existing Proxly architecture patterns (Services for business logic, UI for SwiftUI views, Models for data)

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| Accessibility API fragility | No official Safari profile API exists. Window title pattern matching is the only viable approach to detect profile windows. | Direct Safari API (doesn't exist), AppleScript alone (can't detect which profile window is open), Browser extension (can't create profile windows) |
| Orphaned profile model | Prevents breaking existing rules when users edit profiles. Simpler "edit in place" would invalidate rule references and confuse users. | Edit in place (breaks rules), Force rule update (too complex UX), Block edits (too restrictive), Delete and recreate (loses history) |
| Keyboard shortcut simulation | Safari profiles can only be opened via keyboard shortcuts Option+Cmd+Shift+[1-9]. No programmatic API available. | Menu bar interaction (too unreliable), AppleScript (can't target specific profile), Direct app launch (doesn't open profile windows) |

## Phase 0: Research & Analysis

**Status**: ✅ Complete

### Research Topics

#### 1. macOS Accessibility API for Window Detection

**Decision**: Use `AXUIElementCreateApplication` + `AXUIElementCopyAttributeValue` with `kAXWindowsAttribute` to enumerate Safari windows, then read `kAXTitleAttribute` for pattern matching.

**Rationale**:
- Accessibility API is the only way to read window titles programmatically
- Stable API since macOS 10.9, well-documented
- Used by existing macOS window managers (Rectangle, Magnet, etc.)
- Provides synchronous access to window metadata

**Alternatives Considered**:
- **AppleScript window enumeration**: Too slow (100-500ms), less reliable, harder to parse
- **CGWindowListCopyWindowInfo**: Doesn't provide window titles, only window IDs and bounds
- **Scripting Bridge**: Deprecated, unreliable with sandboxing

**Implementation Pattern**:
```swift
// Pseudo-code for window enumeration
func enumerateSafariWindows() -> [AXUIElement] {
    let safariApp = AXUIElementCreateApplication(safariPID)
    var windowsRef: CFTypeRef?
    AXUIElementCopyAttributeValue(safariApp, kAXWindowsAttribute, &windowsRef)
    return (windowsRef as? [AXUIElement]) ?? []
}

func getWindowTitle(_ window: AXUIElement) -> String? {
    var titleRef: CFTypeRef?
    AXUIElementCopyAttributeValue(window, kAXTitleAttribute, &titleRef)
    return titleRef as? String
}
```

**Best Practices**:
- Check `AXIsProcessTrusted()` before Accessibility operations
- Cache Safari PID lookup to avoid repeated NSWorkspace calls
- Use `AXObserver` for window creation events (future optimization)
- Handle `kAXErrorAPIDisabled` with user-friendly permission guidance

**Performance Characteristics**:
- Window enumeration: 5-20ms for 5-10 windows
- Title reading: <1ms per window
- Total overhead: <50ms for typical use case

**References**:
- Apple Accessibility API docs: https://developer.apple.com/documentation/applicationservices/axuielement
- Example implementations: Rectangle window manager, Hammerspoon automation

#### 2. Keyboard Shortcut Simulation for Profile Window Creation

**Decision**: Use `CGEventCreateKeyboardEvent` + `CGEventPost` to simulate Option+Cmd+Shift+[1-9] keyboard shortcuts for profile window creation.

**Rationale**:
- Only programmatic way to trigger Safari profile window creation
- CGEvent is lower-level and more reliable than NSEvent simulation
- Works outside sandbox (ProxlyHelper) and in standalone version
- Used by existing automation tools (Keyboard Maestro, BetterTouchTool)

**Alternatives Considered**:
- **AppleScript key down**: Less reliable, doesn't support modifier combinations well
- **NSEvent posting**: Higher-level but less control, doesn't work reliably for system shortcuts
- **Accessibility API AXPress**: Doesn't support arbitrary keyboard shortcuts
- **Menu bar automation**: Safari profiles not accessible via menu bar

**Implementation Pattern**:
```swift
// Pseudo-code for keyboard shortcut simulation
func triggerProfileShortcut(order: Int) {
    let keyCode = CGKeyCode(/* 1-9 number key codes */)
    let flags: CGEventFlags = [.maskShift, .maskCommand, .maskAlternate]

    let keyDown = CGEvent(keyboardEventSource: nil, virtualKey: keyCode, keyDown: true)
    keyDown?.flags = flags
    keyDown?.post(tap: .cghidEventTap)

    let keyUp = CGEvent(keyboardEventSource: nil, virtualKey: keyCode, keyDown: false)
    keyUp?.flags = flags
    keyUp?.post(tap: .cghidEventTap)
}
```

**Best Practices**:
- Activate Safari first using `NSWorkspace.shared.launchApplication`
- Wait 100-200ms after activation before sending keyboard events
- Send both key down and key up events with proper flags
- Handle case where Safari not running (launch first)
- Implement 2-second timeout for window creation detection

**Key Code Mappings** (US keyboard layout):
- 1: 18, 2: 19, 3: 20, 4: 21, 5: 23
- 6: 22, 7: 26, 8: 28, 9: 25

**Caveats**:
- May conflict with user's custom keyboard shortcuts
- Requires Accessibility permissions (same as window detection)
- Keyboard layout dependent (assumes US/standard layout)

**References**:
- CGEvent documentation: https://developer.apple.com/documentation/coregraphics/cgevent
- Virtual key codes: https://eastmanreference.com/complete-list-of-applescript-key-codes

#### 3. Orphaned Profile Data Model

**Decision**: Add `isOrphaned: Bool` field to Safari profile model, exclude orphaned profiles from active limits/validation, show in UI with visual distinction.

**Rationale**:
- Preserves backward compatibility with existing rules
- Prevents user confusion when editing profiles referenced by rules
- Allows gradual migration (users can update rules at their pace)
- Follows append-only data pattern (safer than mutation)

**Alternatives Considered**:
- **Edit in place**: Breaks rule references, requires cascade updates
- **Force rule update on edit**: Complex UX, surprising behavior
- **Block edits if referenced**: Too restrictive, poor UX
- **Delete and recreate**: Loses history, breaks rules

**Data Model Extension**:
```swift
// Pseudo-code for Safari profile model
struct SafariProfile: Codable, Identifiable {
    let id: UUID
    var name: String              // Max 50 chars, no em dash
    var order: Int                // 1-9, maps to keyboard shortcut
    var isOrphaned: Bool = false  // True if replaced by edited version
    var createdAt: Date
    var modifiedAt: Date
}
```

**Lifecycle Rules**:
1. **Create**: New profile with `isOrphaned = false`
2. **Edit**: Create new profile (isOrphaned = false), mark original as `isOrphaned = true`
3. **Delete**: Remove profile (active or orphaned)
4. **Validation**: Only non-orphaned profiles checked for duplicate orders, 9-profile limit
5. **UI Display**: Active profiles shown first (sorted by order), orphaned profiles shown after (grayed out, labeled "(legacy)")

**Migration Strategy**:
- Existing browser profiles unaffected (no Safari profiles exist yet)
- Future versions can add "bulk update rules" feature to migrate from orphaned to new profiles

**References**:
- Append-only data patterns: Event sourcing, CQRS
- Similar patterns: Git (commit history), Notion (page history)

#### 4. Em Dash Character Detection and Validation

**Decision**: Use Unicode scalar comparison to detect em dash (U+2014) in profile names, reject during validation with clear error message.

**Rationale**:
- Em dash is the critical separator character in window title pattern
- Preventing it in profile names eliminates pattern matching ambiguity
- Other special characters safe (only appear after em dash in page title)

**Implementation Pattern**:
```swift
// Pseudo-code for em dash validation
func containsEmDash(_ string: String) -> Bool {
    return string.unicodeScalars.contains { $0 == "\u{2014}" }
}

func validateProfileName(_ name: String) -> ValidationResult {
    if name.isEmpty {
        return .failure("Profile name cannot be empty")
    }
    if name.count > 50 {
        return .failure("Profile name cannot exceed 50 characters")
    }
    if containsEmDash(name) {
        return .failure("Profile name cannot contain the em dash character (—)")
    }
    return .success
}
```

**Character Considerations**:
- **Em dash (U+2014)**: PROHIBITED (separator character)
- **En dash (U+2013)**: ALLOWED (different character)
- **Hyphen (U+002D)**: ALLOWED (common in names)
- **Emoji, special chars**: ALLOWED (only appear in page title part)

**Localization Notes**:
- Em dash (U+2014) is standard in English locales
- Other locales may use different separators (e.g., spaced hyphen, colon)
- Initial implementation targets English em dash only
- Future versions can add locale-specific pattern detection

**References**:
- Unicode em dash: U+2014 (—)
- Swift Unicode handling: String.unicodeScalars

#### 5. Accessibility Permission Handling

**Decision**: Check permissions on-error (when Accessibility API calls fail), show user-friendly error with direct link to System Settings, fall back to default Safari.

**Rationale**:
- Avoids performance overhead of pre-checking permissions on every routing attempt
- Provides clear user guidance when permissions missing
- Follows principle of least surprise (only asks when needed)
- Minimizes permission requests (constitution principle V)

**Implementation Pattern**:
```swift
// Pseudo-code for permission handling
func checkAccessibilityPermissions() -> Bool {
    return AXIsProcessTrusted()
}

func handleAccessibilityError() {
    let alert = NSAlert()
    alert.messageText = "Accessibility Permissions Required"
    alert.informativeText = """
    Safari profile routing requires Accessibility permissions to detect Safari windows.

    Please enable permissions in:
    System Settings → Privacy & Security → Accessibility → Proxly
    """
    alert.addButton(withTitle: "Open System Settings")
    alert.addButton(withTitle: "Cancel")

    if alert.runModal() == .alertFirstButtonReturn {
        NSWorkspace.shared.open(URL(string: "x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility")!)
    }
}
```

**Error Detection Strategy**:
- Catch `kAXErrorAPIDisabled` from Accessibility API calls
- Catch `kAXErrorCannotComplete` (permissions revoked mid-session)
- Log all Accessibility errors for debugging
- Show user-facing error only on permission-specific errors

**Fallback Behavior**:
1. Accessibility error detected → Show permission guidance alert
2. User dismisses or opens System Settings → Fall back to default Safari
3. Next routing attempt → Check permissions again (may be granted)

**Permission Request Flow**:
- Initial app launch: Don't request Accessibility (too intrusive)
- First Safari profile routing: Request if needed
- Profile configuration UI: Show warning if permissions not granted

**References**:
- AXIsProcessTrusted(): https://developer.apple.com/documentation/applicationservices/1460720-axisprocesstrusted
- System Settings deep links: x-apple.systempreferences:com.apple.preference.security

### Technology Stack Finalization

**Core Technologies** (from Technical Context):
- **Swift 6.0+**: Primary language
- **SwiftUI**: Profile configuration UI
- **AppKit**: Accessibility API bridge, keyboard event simulation
- **Combine**: Reactive state management for profile changes
- **Accessibility API**: Window enumeration and title reading
- **CGEvent**: Keyboard shortcut simulation
- **AppleScript**: Safari tab creation and URL navigation (existing pattern)

**Testing Stack**:
- **XCTest**: Unit and integration tests
- **Manual testing**: Accessibility API operations, keyboard shortcuts

**Performance Profiling**:
- Instruments Time Profiler: Measure window enumeration overhead
- XCTest performance tests: Validate <50ms routing overhead
- Manual testing: Verify 2-second and 4-second success criteria

## Phase 1: Design & Contracts

**Status**: Ready to execute

### Data Model

See: [data-model.md](./data-model.md) (to be generated)

**Key Entities**:
1. **SafariProfile**: Name, order, orphaned state, timestamps
2. **BrowserProfile extension**: Safari-specific fields
3. **SafariWindowInfo**: Window reference, title, profile match result

### API Contracts

See: [contracts/](./contracts/) (to be generated)

**Internal Service Contracts**:
1. **SafariProfileManager**: CRUD operations, orphaned lifecycle
2. **SafariWindowDetector**: Window enumeration, pattern matching
3. **SafariProfileLauncher**: Window creation, keyboard shortcuts, URL navigation
4. **BrowserLauncher extension**: Safari profile delegation

**Note**: No external API contracts (internal macOS app only)

### Quick Start

See: [quickstart.md](./quickstart.md) (to be generated)

**Developer Onboarding**:
1. Clone repo and checkout `003-safari-profile` branch
2. Install Xcode 16+ with Swift 6.0 support
3. Grant Accessibility permissions to Xcode for testing
4. Run unit tests: `xcodebuild test -project Proxly.xcodeproj -scheme Proxly -destination 'platform=macOS'`
5. Manual test: Configure Safari profile, create rule, test routing

## Next Steps

After `/speckit.plan` completes:
1. Review generated `research.md`, `data-model.md`, `quickstart.md`, `contracts/`
2. Run `/speckit.tasks` to generate dependency-ordered task breakdown
3. Begin implementation following task sequence

**Implementation Order** (aligned with user story priorities):
- Phase 1 (P1): Safari profile configuration (SafariProfileManager, UI)
- Phase 2 (P2): Rule integration (BrowserProfileSelector extension, persistence)
- Phase 3 (P3): Existing window routing (SafariWindowDetector, profile matching)
- Phase 4 (P4): New window creation (SafariProfileLauncher, keyboard shortcuts)
