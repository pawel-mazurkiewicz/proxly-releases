# Service Contract: SafariWindowDetector

**Feature**: Safari Profile Support
**Service**: SafariWindowDetector
**Responsibility**: Enumerate Safari windows via Accessibility API and match window titles against Safari profile names

## Overview

SafariWindowDetector provides window enumeration and pattern matching capabilities for detecting Safari profile windows. It uses macOS Accessibility API to read window titles and matches them against the pattern "PROFILENAME — PAGE TITLE" to identify which profile window corresponds to a given Safari profile.

## Interface Definition

```swift
@MainActor
final class SafariWindowDetector {
    // Singleton instance
    static let shared = SafariWindowDetector()

    // MARK: - Window Enumeration

    /// Enumerates all Safari windows and extracts profile information
    /// - Returns: Array of window info with extracted profile names
    /// - Throws: SafariWindowDetectorError if Accessibility API fails
    func enumerateWindows() async throws -> [SafariWindowInfo]

    /// Finds Safari window matching specific profile name
    /// - Parameter profileName: Target profile name to match
    /// - Returns: Window info if matching window found, nil otherwise
    /// - Throws: SafariWindowDetectorError if Accessibility API fails
    func findWindow(forProfile profileName: String) async throws -> SafariWindowInfo?

    /// Brings Safari window to front using Accessibility API
    /// - Parameter windowElement: AXUIElement window reference
    /// - Throws: SafariWindowDetectorError if activation fails
    func bringWindowToFront(_ windowElement: AXUIElement) async throws

    // MARK: - Permission Checking

    /// Checks if Accessibility permissions are granted
    /// - Returns: True if permissions granted, false otherwise
    func checkAccessibilityPermissions() -> Bool

    /// Shows user-friendly error dialog with guidance to System Settings
    /// - Parameter error: The Accessibility error that occurred
    func handleAccessibilityError(_ error: SafariWindowDetectorError) async

    // MARK: - Pattern Matching

    /// Extracts profile name from Safari window title
    /// - Parameter title: Complete window title from Accessibility API
    /// - Returns: Profile name if pattern matches, nil otherwise
    func extractProfileName(from title: String) -> String?

    /// Checks if window title matches Safari profile pattern
    /// - Parameter title: Window title to check
    /// - Returns: True if title contains em dash separator (U+2014)
    func matchesProfilePattern(_ title: String) -> Bool
}
```

## Data Types

### SafariWindowInfo

```swift
struct SafariWindowInfo {
    /// Accessibility API window reference
    let windowElement: AXUIElement

    /// Complete window title from Accessibility API
    let title: String

    /// Extracted profile name (substring before first " — ")
    let profileName: String?

    /// Extracted page title (substring after first " — ")
    let pageTitle: String?

    /// Safari process ID
    let pid: pid_t
}
```

### SafariWindowDetectorError

```swift
enum SafariWindowDetectorError: LocalizedError {
    case accessibilityPermissionDenied
    case safariNotRunning
    case windowEnumerationFailed(AXError)
    case windowActivationFailed(AXError)
    case invalidWindowElement

    var errorDescription: String? {
        switch self {
        case .accessibilityPermissionDenied:
            return "Accessibility permissions required. Enable in System Settings → Privacy & Security → Accessibility."
        case .safariNotRunning:
            return "Safari is not running"
        case .windowEnumerationFailed(let error):
            return "Failed to enumerate Safari windows: \(error)"
        case .windowActivationFailed(let error):
            return "Failed to bring Safari window to front: \(error)"
        case .invalidWindowElement:
            return "Invalid window reference"
        }
    }

    var recoverySuggestion: String? {
        switch self {
        case .accessibilityPermissionDenied:
            return "Grant Accessibility permissions to Proxly in System Settings"
        case .safariNotRunning:
            return "Safari will be launched automatically"
        case .windowEnumerationFailed, .windowActivationFailed:
            return "Try again or restart Safari"
        case .invalidWindowElement:
            return "Window may have been closed"
        }
    }
}
```

## Method Specifications

### enumerateWindows()

**Purpose**: Enumerate all Safari windows and extract profile information from titles

**Preconditions**:
- Safari is running (checked via `NSWorkspace.runningApplications`)
- Accessibility permissions granted (checked via `AXIsProcessTrusted()`)

**Process**:
1. Check if Safari is running using `NSWorkspace.shared.runningApplications`
2. If not running, throw `SafariWindowDetectorError.safariNotRunning`
3. Check Accessibility permissions using `AXIsProcessTrusted()`
4. If not granted, throw `SafariWindowDetectorError.accessibilityPermissionDenied`
5. Get Safari PID from running applications
6. Create `AXUIElement` for Safari app using `AXUIElementCreateApplication(pid)`
7. Get windows array using `AXUIElementCopyAttributeValue` with `kAXWindowsAttribute`
8. If enumeration fails, throw `SafariWindowDetectorError.windowEnumerationFailed(error)`
9. For each window element:
   - Read window title using `AXUIElementCopyAttributeValue` with `kAXTitleAttribute`
   - Extract profile name using `extractProfileName(from:)`
   - Extract page title (text after first em dash)
   - Create `SafariWindowInfo` struct
10. Return array of `SafariWindowInfo`

**Postconditions**:
- Returns array of all Safari windows with extracted profile information
- Empty array if Safari has no windows
- Windows ordered by Accessibility API enumeration order (typically front-to-back)

**Errors**:
- `SafariWindowDetectorError.safariNotRunning`: Safari not in running applications
- `SafariWindowDetectorError.accessibilityPermissionDenied`: Permissions not granted
- `SafariWindowDetectorError.windowEnumerationFailed`: Accessibility API error

**Performance**: 5-20ms for 5-10 windows (Accessibility API overhead)

**Thread Safety**: `@MainActor` isolated, safe for concurrent calls (serialized)

### findWindow(forProfile:)

**Purpose**: Find specific Safari window matching target profile name

**Preconditions**:
- Safari is running
- Accessibility permissions granted

**Process**:
1. Call `enumerateWindows()` to get all windows
2. Filter windows where `profileName == targetProfileName` (case-sensitive)
3. Return first matching window or `nil`

**Postconditions**:
- Returns first matching window (if multiple windows with same profile name exist)
- Returns `nil` if no matching window found

**Errors**: Same as `enumerateWindows()`

**Performance**: O(n) where n = Safari window count

### bringWindowToFront(_:)

**Purpose**: Activate and bring Safari window to front using Accessibility API

**Preconditions**:
- `windowElement` is valid `AXUIElement` reference
- Accessibility permissions granted

**Process**:
1. Validate `windowElement` is not nil/invalid
2. Raise window using `AXUIElementSetAttributeValue` with `kAXMainAttribute = true`
3. If raise fails, throw `SafariWindowDetectorError.windowActivationFailed(error)`
4. Optionally activate Safari app using `NSWorkspace.shared.launchApplication` with `activateIgnoringOtherApps: true`

**Postconditions**:
- Safari window is frontmost window on screen
- Safari app is active application
- Window receives keyboard focus

**Errors**:
- `SafariWindowDetectorError.invalidWindowElement`: Window element is invalid/nil
- `SafariWindowDetectorError.windowActivationFailed`: Accessibility API error

**Performance**: <5ms (Accessibility API overhead)

### checkAccessibilityPermissions()

**Purpose**: Check if Accessibility permissions are granted

**Process**:
1. Call `AXIsProcessTrusted()`
2. Return boolean result

**Postconditions**:
- Returns `true` if permissions granted
- Returns `false` if permissions denied or not yet requested

**Errors**: None (pure check, no exceptions)

**Performance**: <1ms (system call)

### handleAccessibilityError(_:)

**Purpose**: Show user-friendly error dialog with guidance to enable Accessibility permissions

**Preconditions**:
- Called on main thread (UI operation)

**Process**:
1. Create `NSAlert` with appropriate message for error type
2. Add "Open System Settings" button
3. Add "Cancel" button
4. Show modal dialog
5. If user clicks "Open System Settings":
   - Open URL `x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility`
   - This deep-links to Accessibility pane in System Settings

**Postconditions**:
- User informed about permission requirement
- Optional: System Settings opened to correct pane
- No state changes in application

**Errors**: None (UI operation, cannot fail)

**Performance**: Synchronous UI operation (blocks until user responds)

### extractProfileName(from:)

**Purpose**: Extract profile name from Safari window title using pattern matching

**Pattern**: `"PROFILENAME — PAGE TITLE"`
- Split on first occurrence of ` — ` (space, em dash U+2014, space)
- Profile name is substring before first ` — `
- Page title is substring after first ` — `

**Process**:
1. Search for first occurrence of ` — ` (space + U+2014 + space)
2. If not found, return `nil`
3. Extract substring before ` — ` (trim whitespace)
4. If empty after trimming, return `nil`
5. Return extracted profile name

**Examples**:
- `"Work — Apple"` → `"Work"`
- `"Personal — GitHub — Pull Requests"` → `"Personal"`
- `"Safari"` → `nil`
- `"Work — "` → `"Work"`
- `" — Page"` → `nil` (empty profile name)

**Postconditions**:
- Returns profile name if pattern matches
- Returns `nil` if pattern doesn't match

**Errors**: None (pure function)

**Performance**: <1ms (string operation)

### matchesProfilePattern(_:)

**Purpose**: Check if window title matches Safari profile pattern (contains em dash)

**Process**:
1. Check if title contains ` — ` (space + U+2014 + space)
2. Return `true` if found, `false` otherwise

**Postconditions**:
- Returns `true` if title likely represents profile window
- Returns `false` if title is plain window (no profile)

**Errors**: None (pure function)

**Performance**: <1ms (string operation)

## Accessibility API Integration

### Required Permissions

**Permission**: `com.apple.security.automation.apple-events` (Accessibility)

**Request Flow**:
1. First Accessibility API call triggers system permission prompt
2. User must grant permissions in System Settings → Privacy & Security → Accessibility
3. Subsequent calls succeed if permissions granted

**Permission Checking**:
- Use `AXIsProcessTrusted()` before Accessibility operations
- Catch `kAXErrorAPIDisabled` from API calls
- Show user-friendly guidance dialog with deep link to System Settings

### API Usage Patterns

**Window Enumeration**:
```swift
let safariApp = AXUIElementCreateApplication(safariPID)
var windowsRef: CFTypeRef?
let error = AXUIElementCopyAttributeValue(safariApp, kAXWindowsAttribute as CFString, &windowsRef)
guard error == .success, let windows = windowsRef as? [AXUIElement] else {
    throw SafariWindowDetectorError.windowEnumerationFailed(error)
}
```

**Title Reading**:
```swift
var titleRef: CFTypeRef?
let error = AXUIElementCopyAttributeValue(window, kAXTitleAttribute as CFString, &titleRef)
guard error == .success, let title = titleRef as? String else {
    return nil // Skip window if title unavailable
}
```

**Window Activation**:
```swift
let error = AXUIElementSetAttributeValue(window, kAXMainAttribute as CFString, kCFBooleanTrue)
guard error == .success else {
    throw SafariWindowDetectorError.windowActivationFailed(error)
}
```

### Error Codes

| AXError | Meaning | Handling |
|---------|---------|----------|
| `.success` | Operation succeeded | Continue normally |
| `.apiDisabled` | Accessibility permissions not granted | Throw `.accessibilityPermissionDenied`, show guidance dialog |
| `.cannotComplete` | Operation failed (window closed, etc.) | Retry or fail gracefully |
| `.invalidUIElement` | Window element no longer valid | Throw `.invalidWindowElement` |
| `.notImplemented` | Attribute not supported (rare) | Skip window |

## State Management

**Singleton Instance**:
- `SafariWindowDetector.shared` provides global access
- No internal state (stateless service)
- All methods can be called concurrently (thread-safe)

**Caching Strategy**:
- No window caching (windows change frequently)
- Safari PID cached temporarily during enumeration
- Pattern matching results not cached (negligible performance benefit)

**Concurrency**:
- `@MainActor` isolated (Accessibility API requires main thread)
- Async methods for non-blocking operation
- Multiple concurrent calls serialized automatically

## Integration Points

**Dependencies**:
- macOS Accessibility API (`ApplicationServices` framework)
- `NSWorkspace`: Safari process enumeration and activation
- `SafariProfileManager`: Profile name lookups for matching

**Observers**:
- `SafariProfileLauncher`: Uses `findWindow(forProfile:)` to detect existing profile windows
- `BrowserLauncher`: Uses `enumerateWindows()` and `bringWindowToFront()` for routing

**Error Handling**:
- Accessibility errors surfaced to user with guidance dialog
- Routing operations fall back to default Safari on enumeration failure

## Testing Strategy

**Unit Tests**:
1. `extractProfileName("Work — Apple")` → `"Work"`
2. `extractProfileName("Personal — GitHub — Pull Requests")` → `"Personal"`
3. `extractProfileName("Safari")` → `nil`
4. `extractProfileName("Work — ")` → `"Work"`
5. `extractProfileName(" — Page")` → `nil`
6. `matchesProfilePattern("Work — Page")` → `true`
7. `matchesProfilePattern("Safari")` → `false`

**Integration Tests** (requires Accessibility permissions):
1. Launch Safari with known profile → `enumerateWindows()` → Verify window detected with correct profile name
2. Multiple Safari windows → `enumerateWindows()` → Verify all windows returned
3. Target specific profile → `findWindow(forProfile:)` → Verify correct window returned
4. Bring window to front → `bringWindowToFront(_:)` → Verify window activated

**Manual Tests**:
1. Deny Accessibility permissions → Trigger enumeration → Verify guidance dialog shown
2. Close Safari mid-enumeration → Verify graceful error handling
3. Rename Safari window (change page) → Verify profile name extraction still works

**Mock Strategy**:
- Mock `AXUIElement` responses for unit testing pattern matching
- Real Accessibility API for integration tests (requires test environment with permissions)
- Separate test Safari instance with known profile names
