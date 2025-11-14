# Research: Browser Ordering Technical Decisions

**Date**: 2025-11-03
**Feature**: Customizable Browser & Profile Ordering with Keyboard Shortcuts
**Status**: Complete

## Overview

This document captures technical research and decisions made during Phase 0 of the planning process. All four research questions have been resolved.

---

## Decision 1: CloudKit Sync Strategy

**Question**: Should browser order sync across devices via CloudKit or remain device-local?

**Recommendation**: **Keep browser order device-local only** (separate from synced AppSettings)

### Rationale

1. **Different hardware capabilities**: A MacBook Pro and Mac Mini likely have different browsers installed. Custom browser order on one device (e.g., Chrome → Firefox → Safari) won't make sense on another device that might only have Safari and Brave. Syncing order would lead to:
   - References to browsers that don't exist on the device
   - Confusing UX when order suddenly changes due to cloud sync
   - Need for complex merge logic to handle missing browsers

2. **Existing pattern analysis**: The codebase currently syncs **functional preferences** but keeps **device-specific data** local:
   - **Synced via CloudKit** (in AppSettings): `rulesEnabled`, `defaultBehavior`, `notifyOnFallback`, `enableLogging`, `showTooltips`, `browserVisibility`
   - **Device-specific (not synced)**: Installed apps cache, URL scheme info, last app scan date

3. **Browser visibility precedent**: Interestingly, `browserVisibility` IS synced as part of AppSettings. However, this is fundamentally different:
   - Visibility is boolean: "should Chrome be shown?" works across devices
   - Order is positional: "Chrome in position 2" breaks if Chrome isn't installed
   - Visibility degrades gracefully: hiding a non-existent browser is harmless
   - Order requires reconciliation: what happens to position 2 if that browser doesn't exist?

4. **User value consideration**: Users likely want:
   - **Same visibility rules** across devices (hide Chromium browsers everywhere)
   - **Same routing rules** across devices (github.com always opens in Firefox)
   - **Different browser order** per device (work Mac: Chrome first; home Mac: Safari first)

### Implementation Approach

Store browser order in a separate UserDefaults key that is NOT part of AppSettings:

```swift
// In PersistenceManager
private enum Keys {
    // ... existing keys
    static let browserOrder = "browserOrder"  // Device-local, not synced
}

func saveBrowserOrder(_ order: [String]) {  // Array of bundleIDs
    do {
        let data = try JSONEncoder().encode(order)
        userDefaults.set(data, forKey: Keys.browserOrder)
    } catch {
        AppLogger.log("Error saving browser order: \(error)", level: .error)
    }
}

func loadBrowserOrder() -> [String] {
    guard let data = userDefaults.data(forKey: Keys.browserOrder) else { return [] }
    do {
        return try JSONDecoder().decode([String].self, from: data)
    } catch {
        AppLogger.log("Error loading browser order: \(error)", level: .error)
        return []
    }
}
```

This approach:
- Uses `UserDefaults.standard` (same as current implementation)
- Does NOT trigger CloudKit sync
- Automatically isolated per-device
- No CloudKitSyncManager changes needed

### Trade-offs

**Pros**:
- No sync conflicts from device-specific browser installations
- Simpler implementation (no merge logic needed)
- Follows existing architectural pattern (device-specific data stays local)
- Better UX: users can customize per-device without unexpected changes
- No risk of broken references to non-existent browsers
- Consistent with how installed apps cache is handled

**Cons**:
- Users must manually set browser order on each device
- No benefit for users who want identical setups across all devices
- Slight inconsistency: `browserVisibility` syncs but `browserOrder` doesn't (though this makes logical sense)

**Alternative considered**: Sync with reconciliation logic (rejected due to complexity and questionable user value)

---

## Decision 2: SwiftUI Drag-and-Drop Implementation Pattern

**Question**: What's the best approach for drag-and-drop in a ScrollView with LazyVStack?

**Recommendation**: **Use SwiftUI's native `.draggable()` + `.dropDestination()` API pattern** (macOS 14+)

This is the modern, declarative approach that's already proven in `RuleListView.swift` (lines 289-293) and works well with `List` + `ForEach` structures.

### Code Example

```swift
// Based on existing RuleListView implementation
struct BrowserManagementView: View {
    @Binding var settings: AppSettings
    @State private var availableBrowsers: [Browser] = []

    var body: some View {
        ScrollView {
            LazyVStack(spacing: 12) {
                ForEach(availableBrowsers) { browser in
                    browserRow(for: browser)
                        .draggable(browser)
                        .dropDestination(for: Browser.self) { droppedBrowsers, location in
                            guard let droppedBrowser = droppedBrowsers.first else { return false }
                            moveBrowser(dragged: droppedBrowser, droppedOn: browser)
                            return true
                        }
                }
            }
            .padding(.vertical, 8)
        }
    }

    private func moveBrowser(dragged: Browser, droppedOn: Browser) {
        guard dragged.bundleID != droppedOn.bundleID else { return }

        guard let fromIndex = availableBrowsers.firstIndex(where: { $0.bundleID == dragged.bundleID }),
              let toIndex = availableBrowsers.firstIndex(where: { $0.bundleID == droppedOn.bundleID })
        else { return }

        // Update local state
        availableBrowsers.move(fromOffsets: IndexSet(integer: fromIndex), toOffset: toIndex)

        // Persist to BrowserOrderService
        BrowserOrderService.shared.updateOrder(availableBrowsers.map { $0.bundleID })
    }
}
```

**Prerequisites**: Browser model must conform to `Transferable` protocol:
```swift
// In Browser.swift
extension Browser: Transferable {
    static var transferRepresentation: some TransferRepresentation {
        CodableRepresentation(contentType: .json)
    }
}
```

### Visual Feedback Strategy

- **Drag preview**: Native system preview showing browser icon + name (provided automatically by `.draggable()`)
- **Drop indicator**: SwiftUI provides implicit visual feedback when hovering over drop zones
- **Animation**: Use `.animation(.default, value: availableBrowsers)` on the `LazyVStack` to animate reordering
- **Custom preview** (optional): Pass custom view to `.draggable(_:preview:)` for branded appearance

### Accessibility Considerations

**Keyboard reordering**: Implement using `.accessibilityAdjustableAction` (pattern already used in BrowserManagementView lines 136-156):
```swift
.accessibilityAdjustableAction { direction in
    let currentIndex = availableBrowsers.firstIndex(where: { $0.bundleID == browser.bundleID }) ?? 0
    let newIndex: Int

    switch direction {
    case .increment:
        newIndex = min(currentIndex + 1, availableBrowsers.count - 1)
    case .decrement:
        newIndex = max(currentIndex - 1, 0)
    @unknown default:
        return
    }

    if currentIndex != newIndex {
        availableBrowsers.move(fromOffsets: IndexSet(integer: currentIndex), toOffset: newIndex)
        BrowserOrderService.shared.updateOrder(availableBrowsers.map { $0.bundleID })
    }
}
```

**VoiceOver announcements**: Add dynamic position labels:
```swift
.accessibilityLabel(Text("Position \(index + 1) of \(availableBrowsers.count): \(browser.displayName)"))
.accessibilityHint(Text("Swipe up or down to reorder"))
```

### macOS 14+ Compatibility

- ✅ `.draggable()` available in macOS 13.0+ (Proxly targets macOS 14.0+)
- ✅ `.dropDestination(for:action:)` available in macOS 13.0+
- ✅ `Transferable` protocol available in macOS 13.0+
- Pattern is already proven in `RuleListView.swift`

### Implementation Notes

1. Rule already conforms to `Transferable` (line 12 in Rule.swift), so Browser needs same treatment
2. Suppress reload notifications during drag operations to prevent UI flicker (see `suppressNextReload` pattern in RuleListView line 382)
3. Use `@State` for browser list to enable SwiftUI automatic diffing and animation

**Alternative Considered**: Using `List` with `.onMove()` modifier + edit mode toggle. Rejected because:
- Requires explicit "Edit" mode toggle (extra user friction)
- RuleListView demonstrates always-on drag works well
- Proxly design philosophy favors direct manipulation over mode switching

---

## Decision 3: Profile Ordering Data Model

**Question**: Should profile order be nested under browser entries or flattened?

**Recommendation**: **Flat Structure (Option B)**

### Data Structure

```swift
/// Represents a single item in the custom browser/profile ordering.
/// Each entry maps to one keyboard shortcut position (1-9).
struct BrowserOrderEntry: Codable, Identifiable, Hashable {
    /// Stable identifier for SwiftUI list operations
    let id: UUID

    /// Bundle ID of the browser (e.g., "com.google.Chrome")
    let bundleID: String

    /// Optional profile name. If nil, this entry represents the browser itself
    /// (default profile behavior). If set, represents a specific browser profile.
    let profileName: String?

    /// Zero-based position in the custom order. Lower values appear first.
    /// This directly maps to keyboard shortcuts: displayOrder 0 → key "1", etc.
    let displayOrder: Int

    init(bundleID: String, profileName: String? = nil, displayOrder: Int) {
        self.id = UUID()
        self.bundleID = bundleID
        self.profileName = profileName
        self.displayOrder = displayOrder
    }

    /// Converts this entry to a BrowserTarget for URL processing
    func toBrowserTarget() -> BrowserTarget {
        return BrowserTarget(bundleID: bundleID, profileName: profileName)
    }
}

/// Persisted state for browser ordering feature
struct BrowserOrderState: Codable, Equatable {
    /// Flat array of all custom-ordered items (browsers and profiles).
    /// Empty array means no custom order (use default alphabetical).
    var orderedEntries: [BrowserOrderEntry]

    /// Timestamp of last manual reordering action (for debugging/logging)
    var lastModified: Date

    /// Schema version for future migrations
    let version: Int = 1

    init(orderedEntries: [BrowserOrderEntry] = [], lastModified: Date = Date()) {
        self.orderedEntries = orderedEntries
        self.lastModified = lastModified
    }
}
```

### Rationale

**P1 Implementation (Browser-Only)**:
Both nested and flat are equally simple for browser-only ordering. Each browser is one entry with `profileName: nil`.

**P2 Extension (Profile-Level Ordering) - CRITICAL**:

The spec requires **interleaving** browsers and profiles:
> "Users should be able to interleave browsers and profiles (e.g., Chrome Work → Safari → Chrome Personal → Firefox)"

**Nested structure CANNOT satisfy this requirement**:
- With nested `profileOrder: [String]` inside browser entries, all profiles are locked inside their browser
- If Chrome has `displayOrder: 0` and Safari has `displayOrder: 1`, ALL Chrome profiles appear before Safari
- **Cannot** place Safari between Chrome Work and Chrome Personal

**Flat structure naturally supports interleaving**:
```swift
// Example: Chrome Work → Safari → Chrome Personal → Firefox
[
    BrowserOrderEntry(bundleID: "com.google.Chrome", profileName: "Work", displayOrder: 0),
    BrowserOrderEntry(bundleID: "com.apple.Safari", profileName: nil, displayOrder: 1),
    BrowserOrderEntry(bundleID: "com.google.Chrome", profileName: "Personal", displayOrder: 2),
    BrowserOrderEntry(bundleID: "org.mozilla.firefox", profileName: nil, displayOrder: 3)
]
```

**UI Ergonomics**:
- Flat structure: Single-level drag-and-drop, any item can be placed anywhere
- Nested structure: Requires complex two-level dragging, profiles cannot escape their browser

**Keyboard Shortcut Mapping**:
Spec states: "keyboard shortcuts 1-9 map to a flat list of browser+profile combinations"
- Flat structure: `displayOrder` directly maps to keyboard shortcut (natural representation)
- Nested structure: Requires runtime flattening at every display point (duplication, bug risk)

### Migration Path (P1 → P2)

**Phase 1** (Browser-Only):
```swift
BrowserOrderState(
    orderedEntries: [
        BrowserOrderEntry(bundleID: "com.apple.Safari", profileName: nil, displayOrder: 0),
        BrowserOrderEntry(bundleID: "com.google.Chrome", profileName: nil, displayOrder: 1)
    ]
)
```

**Phase 2** (Profiles Enabled):
```swift
// User expands Chrome and inserts profiles:
BrowserOrderState(
    orderedEntries: [
        BrowserOrderEntry(bundleID: "com.apple.Safari", profileName: nil, displayOrder: 0),
        BrowserOrderEntry(bundleID: "com.google.Chrome", profileName: "Work", displayOrder: 1),
        BrowserOrderEntry(bundleID: "com.google.Chrome", profileName: "Personal", displayOrder: 2)
    ]
)

// User drags Safari between Chrome profiles (interleaving):
BrowserOrderState(
    orderedEntries: [
        BrowserOrderEntry(bundleID: "com.google.Chrome", profileName: "Work", displayOrder: 0),
        BrowserOrderEntry(bundleID: "com.apple.Safari", profileName: nil, displayOrder: 1),
        BrowserOrderEntry(bundleID: "com.google.Chrome", profileName: "Personal", displayOrder: 2)
    ]
)
```

### Trade-offs

**Pros**:
- ✅ **Spec compliance**: Directly supports "interleave browsers and profiles" requirement
- ✅ **Keyboard shortcuts**: Natural 1:1 mapping between `displayOrder` and shortcut keys
- ✅ **UI simplicity**: Single-level drag-and-drop, familiar UX pattern
- ✅ **Consistency**: No flattening logic duplication across UI surfaces
- ✅ **Flexibility**: Full control over ordering (any profile can be placed anywhere)
- ✅ **Testability**: Simpler test cases (no nested logic to test)

**Cons**:
- ❌ **Migration complexity**: P1 → P2 requires renumbering entries (but runs once)
- ❌ **Duplicate bundleIDs**: Same browser appears multiple times (but this is intentional for profiles)

**Why Nested Was Rejected**:
- ❌ **Cannot interleave**: Spec explicitly requires interleaving, nested structure cannot deliver
- ❌ **Runtime overhead**: Requires flattening logic at every display point
- ❌ **Complexity**: Two-level drag-and-drop harder to implement and use

---

## Decision 4: Migration Strategy for Existing Users

**Question**: How do we initialize custom order for users upgrading from a version without this feature?

**Recommendation**: **Approach A: Alphabetical by display name**

### Current BrowserDetector Behavior

**Order**: Browsers are returned **alphabetically sorted by display name**
- `AppEnumerationService.getBrowserApps()` sorts by: `$0.displayName.lowercased() < $1.displayName.lowercased()` (line 269)
- `BrowserDetector.detectBrowsers()` preserves this alphabetical order
- **Deterministic**: Yes, completely consistent across app launches and devices

**Example order**: Brave Browser, Google Chrome, Microsoft Edge, Safari

### Implementation

Add to `MigrationManager.swift` as version 6:

```swift
private func migrateToVersion6() {
    AppLogger.log("Migrating to version 6: Initialize custom browser order", level: .info)

    var settings = PersistenceManager.shared.loadAppSettings()

    // Only initialize if not already set
    guard settings.customBrowserOrder.isEmpty else {
        AppLogger.log("Custom browser order already initialized, skipping", level: .info)
        return
    }

    // Get browsers in their current alphabetical order
    let browsers = BrowserDetector.shared.detectBrowsers()

    // Initialize order with current alphabetical order (preserves existing behavior)
    settings.customBrowserOrder = browsers.map { $0.bundleID }

    // Save without triggering sync (browser order is device-local)
    PersistenceManager.shared.saveAppSettings(settings, triggerSync: false)

    AppLogger.log("Migration completed: Initialized order for \(browsers.count) browsers", level: .info)
}
```

**When to run**: During app launch, as part of `MigrationManager.shared.runMigrations()` in `ProxlyApp.swift`

### User Experience

**Will users notice any change?**
**No**, because:
- Current order is already alphabetical
- Migration preserves alphabetical order
- No UI changes until they explicitly use reordering feature
- New users also start with alphabetical order

**Should we show a one-time hint about the new feature?**
**No**, because:
- Feature is discoverable through standard macOS drag-and-drop patterns
- Tooltip system will explain when users hover
- "What's New" in app updates can mention it
- No need to interrupt users with promotional messaging

### Why Alphabetical is Best

1. **Zero behavior change**: Users already see browsers alphabetically
2. **Zero migration code complexity**: Current detection already returns alphabetical order
3. **Predictable**: Consistent across devices
4. **No surprises**: Users upgrading won't notice any change initially
5. **Existing precedent**: App lists are alphabetically sorted in multiple places (AppEnumerationService, RuleEditView)

**Alternatives Rejected**:
- **Usage-based ordering**: No usage data tracked, privacy implications, complex
- **Random detection order**: Already alphabetical, so moot point

---

## Summary

All four research questions have been resolved:

1. **CloudKit Sync**: Device-local only (no sync)
2. **Drag-and-Drop**: Native SwiftUI `.draggable()` + `.dropDestination()` pattern
3. **Data Model**: Flat structure with `BrowserOrderEntry { bundleID, profileName?, displayOrder }`
4. **Migration**: Alphabetical initialization preserving existing behavior

These decisions form the foundation for Phase 1 design artifacts.
