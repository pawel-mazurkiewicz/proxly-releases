# Quick Start: Safari Profile Support

**Feature**: Safari Profile Support
**Branch**: `003-safari-profile`
**Target Audience**: Developers implementing or testing the feature

## Prerequisites

### Development Environment

- **macOS**: 14.0 or later (Sonoma+)
- **Xcode**: 16.0 or later
- **Swift**: 6.0 or later
- **Git**: For branch checkout and version control

### System Configuration

- **Accessibility Permissions**: Grant to Xcode and/or Proxly app for testing
  - Path: System Settings → Privacy & Security → Accessibility → Enable "Xcode" and "Proxly"
- **Safari Profiles**: Configure at least 2 profiles in Safari for testing
  - Safari → Safari Settings → Profiles → Create profiles named "Work" and "Personal"
  - Assign keyboard shortcuts: Work = Option+Cmd+Shift+1, Personal = Option+Cmd+Shift+2

### Knowledge Requirements

- Swift 6.0 concurrency (`async/await`, `@MainActor`)
- SwiftUI basics (views, state management, `@Published` properties)
- macOS Accessibility API familiarity (helpful but not required)
- Proxly codebase architecture (see `CLAUDE.md`)

## Getting Started

### 1. Clone and Checkout

```bash
# Clone repository (if not already cloned)
git clone https://github.com/pawel-mazurkiewicz/proxly.git
cd proxly

# Checkout feature branch
git checkout 003-safari-profile

# Verify branch
git branch --show-current
# Output: 003-safari-profile
```

### 2. Open Project in Xcode

```bash
open Proxly.xcodeproj
```

**Project Structure** (relevant to this feature):
```
Proxly/
├── Models/
│   ├── BrowserProfile.swift       (EXTEND: Add safariProfileId field)
│   └── SafariProfile.swift        (NEW: Safari profile model)
├── Services/
│   ├── SafariProfileManager.swift  (NEW: Profile CRUD operations)
│   ├── SafariWindowDetector.swift  (NEW: Accessibility API window detection)
│   ├── SafariProfileLauncher.swift (NEW: Profile window creation, keyboard shortcuts)
│   └── BrowserLauncher.swift       (EXTEND: Delegate to SafariProfileLauncher)
├── UI/
│   ├── SafariProfileConfigView.swift (NEW: Profile configuration UI)
│   └── BrowserProfileSelector.swift  (EXTEND: Show active + orphaned profiles)
└── ProxlyTests/
    ├── SafariProfileManagerTests.swift  (NEW: Profile CRUD tests)
    └── SafariWindowDetectorTests.swift  (NEW: Pattern matching tests)
```

### 3. Install Dependencies

**No external dependencies required** for this feature. Proxly uses native frameworks:
- SwiftUI, AppKit, Combine (built-in)
- Accessibility API (macOS system framework)
- CGEvent (macOS system framework)

**Verify dependencies**:
- Xcode → Proxly target → Build Phases → Link Binary With Libraries
- Should include: `ApplicationServices.framework` (for Accessibility API)

### 4. Build and Run

#### Build the Project

```bash
# Command line build
xcodebuild -project Proxly.xcodeproj -scheme Proxly -destination 'platform=macOS' build

# Or use Xcode: Product → Build (Cmd+B)
```

#### Run the App

```bash
# Command line run
xcodebuild -project Proxly.xcodeproj -scheme Proxly -destination 'platform=macOS' run

# Or use Xcode: Product → Run (Cmd+R)
```

**Expected Behavior**:
- Proxly menu bar icon appears in top-right corner
- No immediate changes visible (Safari profile UI not yet implemented)
- Check Console.app for log messages: `com.proxly.Proxly`

### 5. Run Unit Tests

```bash
# Run all tests
xcodebuild test -project Proxly.xcodeproj -scheme Proxly -destination 'platform=macOS'

# Run specific test class
xcodebuild test -project Proxly.xcodeproj -scheme Proxly -destination 'platform=macOS' -only-testing:ProxlyTests/SafariProfileManagerTests

# Or use Xcode: Product → Test (Cmd+U)
```

**Expected Test Results**:
- All existing tests pass (baseline)
- New tests for Safari profile feature (as implemented)
- Test coverage report: Xcode → Report Navigator → Coverage

## Development Workflow

### Phase 1: Safari Profile Configuration (P1)

**Implementation Order**:
1. Create `SafariProfile.swift` model with `Codable` conformance
2. Extend `AppSettings` to include `safariProfiles: [SafariProfile]` field
3. Implement `SafariProfileManager` service (CRUD operations, validation)
4. Create `SafariProfileConfigView` SwiftUI view
5. Integrate config view into Settings → Browser Settings
6. Write unit tests for validation (name, order, em dash, orphaned lifecycle)

**Testing Checklist**:
- [ ] Create profile with valid name/order → Success
- [ ] Create profile with em dash → Validation error
- [ ] Create profile with duplicate order → Validation error
- [ ] Create 10th active profile → Limit error
- [ ] Edit profile → Original orphaned, new created
- [ ] Delete profile → Removed from storage
- [ ] Restart app → Profiles persist

**Manual Testing**:
1. Open Proxly Settings
2. Navigate to Browser Settings → Safari Profiles
3. Add profile "Work" with order 1 → Verify saved
4. Add profile "Personal" with order 2 → Verify saved
5. Edit "Work" to "Work2" → Verify original marked orphaned
6. Delete "Personal" → Verify removed
7. Restart Proxly → Verify "Work2" and orphaned "Work" both present

### Phase 2: Rule Integration (P2)

**Implementation Order**:
1. Extend `BrowserProfile` with `safariProfileId: UUID?` field
2. Update `PersistenceManager` to persist Safari profiles
3. Extend `BrowserProfileSelector` to show active + orphaned Safari profiles
4. Add visual distinction for orphaned profiles (gray out, "(legacy)" label)
5. Write integration tests for rule persistence with Safari profiles

**Testing Checklist**:
- [ ] Create rule with Safari profile → Profile reference saved
- [ ] Edit profile → Rule still references original (orphaned) profile
- [ ] Delete orphaned profile → Rule remains valid
- [ ] Profile selector shows active profiles first, orphaned after
- [ ] Orphaned profiles visually distinguished (grayed out)

**Manual Testing**:
1. Create domain rule for `github.com` → Select Safari → Select "Work" profile → Save
2. Open rule editor → Verify "Work" profile selected
3. Edit "Work" profile to "Work2"
4. Open rule editor → Verify still shows "Work" (orphaned, grayed out)
5. Create new rule → Verify profile selector shows "Work2" (active) and "Work" (orphaned, grayed out)

### Phase 3: Existing Window Routing (P3)

**Implementation Order**:
1. Implement `SafariWindowDetector` service (Accessibility API window enumeration)
2. Implement pattern matching logic (`extractProfileName`, `matchesProfilePattern`)
3. Implement `findWindow(forProfile:)` and `bringWindowToFront(_:)`
4. Extend `BrowserLauncher` to delegate Safari profile routing to new service
5. Add Accessibility permission checking and error handling
6. Write unit tests for pattern matching, integration tests for window detection

**Testing Checklist**:
- [ ] Pattern matching: "Work — Apple" → extracts "Work"
- [ ] Pattern matching: "Safari" → returns nil
- [ ] Window enumeration: Multiple Safari windows → All returned
- [ ] Find window: Target "Work" profile → Correct window found
- [ ] Bring to front: Window activated and raised
- [ ] Accessibility denied → User shown guidance dialog

**Manual Testing**:
1. Open Safari with "Work" profile window (Option+Cmd+Shift+1)
2. Navigate to a page (e.g., apple.com) → Verify title shows "Work — Apple"
3. Create Proxly rule: `apple.com` → Safari → "Work" profile
4. Click URL for `apple.com` → Verify existing "Work" window comes to front and opens in new tab
5. Open Safari "Personal" profile window (Option+Cmd+Shift+2)
6. Click URL for `apple.com` → Verify "Work" window activated (not "Personal")
7. Revoke Accessibility permissions → Trigger routing → Verify guidance dialog shown

### Phase 4: New Window Creation (P4)

**Implementation Order**:
1. Implement `SafariProfileLauncher.createProfileWindow(order:)` (keyboard shortcut simulation)
2. Add `triggerProfileShortcut(order:)` using `CGEvent` API
3. Implement 2-second timeout and retry logic for window detection
4. Add fallback to default Safari on timeout
5. Write unit tests for keyboard shortcut generation, integration tests for window creation

**Testing Checklist**:
- [ ] No Safari running → Launch Safari and create profile window
- [ ] Safari running, no "Work" window → Create "Work" window via shortcut
- [ ] Wait for window creation → Detect new window within 2 seconds
- [ ] Window creation fails → Fallback to default Safari
- [ ] Safari unresponsive → Timeout and fallback

**Manual Testing**:
1. Quit Safari completely
2. Create Proxly rule: `github.com` → Safari → "Work" profile
3. Click URL for `github.com` → Verify:
   - Safari launches
   - Keyboard shortcut Option+Cmd+Shift+1 triggered
   - "Work" profile window created
   - URL opens in new window
   - Total time < 4 seconds (p90 target)
4. Close "Work" window (keep Safari running)
5. Click URL for `github.com` → Verify new "Work" window created (don't need to launch Safari)
6. Wait > 2 seconds for window creation (simulate slow system) → Verify fallback to default Safari

## Debugging Tips

### Enable Verbose Logging

```swift
// In SafariProfileManager.swift
let logger = Logger(subsystem: "com.proxly.Proxly", category: "SafariProfile")
logger.debug("Profile created: \(profile.name) (order: \(profile.order))")
logger.info("Profile edited: \(originalId) -> \(newProfile.id)")
logger.error("Failed to save profile: \(error)")
```

**View Logs**:
```bash
# Follow Proxly logs in real-time
log stream --predicate 'subsystem == "com.proxly.Proxly"' --level debug

# Filter for Safari profile logs
log stream --predicate 'subsystem == "com.proxly.Proxly" AND category == "SafariProfile"' --level debug
```

### Debug Accessibility API Issues

**Check Permissions**:
```swift
import ApplicationServices

// In SafariWindowDetector
let trusted = AXIsProcessTrusted()
print("Accessibility permissions granted: \(trusted)")
```

**Enumerate Windows Manually**:
```swift
// In SafariWindowDetector.enumerateWindows()
let windows = try await enumerateWindows()
for window in windows {
    print("Window: \(window.title)")
    print("  Profile: \(window.profileName ?? "nil")")
    print("  Page: \(window.pageTitle ?? "nil")")
}
```

**Test Pattern Matching**:
```swift
let detector = SafariWindowDetector.shared
print(detector.extractProfileName(from: "Work — Apple"))  // "Work"
print(detector.extractProfileName(from: "Safari"))        // nil
```

### Debug Keyboard Shortcut Simulation

**Verify Key Codes**:
```swift
// Key codes for 1-9 (US keyboard layout)
let keyCodes: [Int: CGKeyCode] = [
    1: 18, 2: 19, 3: 20, 4: 21, 5: 23,
    6: 22, 7: 26, 8: 28, 9: 25
]

// Test shortcut generation
let order = 1
let keyCode = keyCodes[order]!
print("Order \(order) -> Key code \(keyCode)")
```

**Monitor Keyboard Events**:
```bash
# Use Karabiner EventViewer to see keyboard events
# Download: https://karabiner-elements.pqrs.org/
```

### Common Issues

#### Issue: Accessibility Permissions Denied

**Symptom**: `SafariWindowDetectorError.accessibilityPermissionDenied` thrown

**Solution**:
1. Open System Settings → Privacy & Security → Accessibility
2. Enable checkbox for "Proxly" (or "Xcode" for testing)
3. Restart Proxly
4. Verify: `AXIsProcessTrusted()` returns `true`

#### Issue: Pattern Matching Fails (Non-English Locale)

**Symptom**: `extractProfileName()` returns `nil` for valid Safari windows

**Solution**:
1. Check window title in Console: `log stream --predicate 'category == "SafariProfile"'`
2. Verify em dash character: Should be U+2014, not U+2013 or hyphen
3. If different separator, update pattern matching logic for locale
4. Document limitation in user-facing docs

#### Issue: Keyboard Shortcut Doesn't Create Window

**Symptom**: Safari activated but no profile window created

**Solution**:
1. Verify Safari profile keyboard shortcuts: Safari Settings → Profiles
2. Check if user has custom keyboard shortcut conflicting with Option+Cmd+Shift+[1-9]
3. Increase activation delay: Wait 200-300ms after `NSWorkspace.launchApplication`
4. Test manually: Press Option+Cmd+Shift+1 → Should create profile window

#### Issue: Window Detection Timeout

**Symptom**: Profile window created but not detected within 2 seconds

**Solution**:
1. Increase timeout to 3-4 seconds for slow systems
2. Add retry logic: Re-enumerate windows 2-3 times with 500ms delays
3. Check window title in Console: May not match expected pattern
4. Fallback to default Safari after timeout (acceptable degradation)

## Performance Profiling

### Measure Window Enumeration Overhead

```swift
let start = CFAbsoluteTimeGetCurrent()
let windows = try await SafariWindowDetector.shared.enumerateWindows()
let duration = CFAbsoluteTimeGetCurrent() - start
print("Window enumeration took \(duration * 1000)ms for \(windows.count) windows")
// Target: < 50ms for typical use case (5-10 windows)
```

### Profile URL Routing End-to-End

```swift
// In URLProcessor.processURL
let start = CFAbsoluteTimeGetCurrent()
// ... existing window routing logic ...
let duration = CFAbsoluteTimeGetCurrent() - start
print("Safari profile routing took \(duration * 1000)ms")
// Target: < 2 seconds p95 for existing windows, < 4 seconds p90 for new windows
```

### Use Instruments

```bash
# Profile with Time Profiler
xcodebuild -project Proxly.xcodeproj -scheme Proxly -destination 'platform=macOS' \
  -derivedDataPath ./DerivedData \
  build

instruments -t "Time Profiler" ./DerivedData/Build/Products/Debug/Proxly.app

# Profile with Allocations (check for memory leaks)
instruments -t "Allocations" ./DerivedData/Build/Products/Debug/Proxly.app
```

**Key Metrics**:
- Window enumeration: 5-20ms (baseline)
- Pattern matching: <1ms per window
- Keyboard shortcut simulation: 100-200ms
- Window creation detection: 0-2000ms (Safari response time)

## Next Steps

After completing development and testing:

1. **Code Review**: Create pull request for `003-safari-profile` → `main`
2. **User Testing**: Beta test with 5-10 users (different Safari configurations)
3. **Documentation**: Update user-facing docs with Safari profile instructions
4. **Localization**: Add Safari profile strings to all 7 language `.lproj` files
5. **Release**: Merge to `main` and include in next Proxly release

## Resources

### Documentation

- **Feature Spec**: [spec.md](./spec.md)
- **Implementation Plan**: [plan.md](./plan.md)
- **Data Model**: [data-model.md](./data-model.md)
- **Service Contracts**: [contracts/](./contracts/)
- **Proxly Architecture**: `/CLAUDE.md` in repo root
- **Proxly Constitution**: `/.specify/memory/constitution.md`

### Apple Documentation

- **Accessibility API**: https://developer.apple.com/documentation/applicationservices/axuielement
- **CGEvent API**: https://developer.apple.com/documentation/coregraphics/cgevent
- **NSWorkspace**: https://developer.apple.com/documentation/appkit/nsworkspace
- **SwiftUI**: https://developer.apple.com/documentation/swiftui

### Community Resources

- **Accessibility API Examples**: Rectangle window manager (open source)
- **Keyboard Shortcut Simulation**: Hammerspoon automation (open source)
- **Swift Concurrency**: Swift.org documentation

## Getting Help

**Internal**:
- Review specification documents in `/specs/003-safari-profile/`
- Check CLAUDE.md for Proxly architecture patterns
- Review existing browser profile code (Chrome/Firefox profiles)

**External**:
- Apple Developer Forums: https://developer.apple.com/forums/
- Stack Overflow: Tag with `macos`, `accessibility-api`, `swift`
- Proxly GitHub Issues: Report bugs or request features
