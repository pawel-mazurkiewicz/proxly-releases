# Data Model: Customizable Browser & Profile Ordering

**Feature**: 002-browser-order-shortcuts
**Created**: 2025-11-03
**Based on**: [research.md](./research.md) Decision 3 (Flat Structure)

## Overview

This document defines the data structures for storing user-defined browser and profile ordering. The design uses a **flat structure** where each browser or profile is represented as a separate entry with an explicit display order. This enables the key requirement of **interleaving browsers and profiles** (e.g., Chrome Work → Safari → Chrome Personal → Firefox).

## Core Entities

### BrowserOrderEntry

Represents a single item in the custom browser/profile ordering. Each entry maps to one keyboard shortcut position (1-9).

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
```

**Properties**:
- `id`: UUID for SwiftUI `ForEach` operations and drag-and-drop tracking
- `bundleID`: Unique browser identifier (e.g., "com.google.Chrome", "com.apple.Safari")
- `profileName`: Optional profile name (nil = browser default profile, string = specific profile)
- `displayOrder`: Zero-based index determining visual position and keyboard shortcut mapping

**Validation Rules**:
- `bundleID` must be non-empty
- `displayOrder` must be >= 0
- `displayOrder` values should be unique within a BrowserOrderState (enforced by service layer)
- `profileName` is optional (nil is valid for browser-only entries)

**Equality**: Uses `Hashable` conformance for efficient duplicate detection in sets/dictionaries

---

### BrowserOrderState

Persisted state for the browser ordering feature. This structure is stored in UserDefaults as device-local data (not synced via CloudKit).

```swift
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

**Properties**:
- `orderedEntries`: Array of BrowserOrderEntry items sorted by `displayOrder`
- `lastModified`: Timestamp updated whenever user manually reorders items
- `version`: Schema version (starts at 1) for future migration support

**Storage**:
- Stored in `UserDefaults.standard` under key `"browserOrder"`
- JSON-encoded using `JSONEncoder`/`JSONDecoder`
- **Device-local only** (does not sync via CloudKit)

**Default Behavior**:
- Empty `orderedEntries` array = use default alphabetical order by display name
- First launch initializes with alphabetical order from `BrowserDetector`

---

## State Transitions

### Browser Uninstalled

**Trigger**: System browser is uninstalled, detected during next browser refresh

**Action**:
1. Remove all `BrowserOrderEntry` items with matching `bundleID`
2. Renumber remaining entries to close gaps in `displayOrder`
3. Update `lastModified` timestamp

**Example**:
```swift
// Before: Chrome, Safari, Firefox (orders 0, 1, 2)
orderedEntries = [
    BrowserOrderEntry(bundleID: "com.google.Chrome", displayOrder: 0),
    BrowserOrderEntry(bundleID: "com.apple.Safari", displayOrder: 1),
    BrowserOrderEntry(bundleID: "org.mozilla.firefox", displayOrder: 2)
]

// Safari uninstalled
// After: Chrome, Firefox (orders 0, 1)
orderedEntries = [
    BrowserOrderEntry(bundleID: "com.google.Chrome", displayOrder: 0),
    BrowserOrderEntry(bundleID: "org.mozilla.firefox", displayOrder: 1)  // renumbered from 2 → 1
]
```

---

### Browser Newly Detected

**Trigger**: "Refresh Browser List" discovers a new browser not in custom order

**Action**:
1. Append new browser to end of `orderedEntries` with `displayOrder = max(currentOrders) + 1`
2. Set `profileName = nil` (browser-only entry)
3. Update `lastModified` timestamp

**Example**:
```swift
// Before: Chrome (0), Safari (1), Firefox (2)
// Brave Browser installed and detected

// After: Chrome (0), Safari (1), Firefox (2), Brave (3)
orderedEntries.append(
    BrowserOrderEntry(bundleID: "com.brave.Browser", profileName: nil, displayOrder: 3)
)
```

---

### Profile Deleted Externally

**Trigger**: User deletes a browser profile from browser settings (e.g., removes "Work" profile in Chrome)

**Action**:
1. On next profile scan, detect missing profile
2. Remove `BrowserOrderEntry` with matching `bundleID` and `profileName`
3. Renumber remaining entries
4. Update `lastModified` timestamp

**Example**:
```swift
// Before: Chrome Work (0), Safari (1), Chrome Personal (2)
orderedEntries = [
    BrowserOrderEntry(bundleID: "com.google.Chrome", profileName: "Work", displayOrder: 0),
    BrowserOrderEntry(bundleID: "com.apple.Safari", profileName: nil, displayOrder: 1),
    BrowserOrderEntry(bundleID: "com.google.Chrome", profileName: "Personal", displayOrder: 2)
]

// "Work" profile deleted in Chrome settings
// After: Safari (0), Chrome Personal (1)
orderedEntries = [
    BrowserOrderEntry(bundleID: "com.apple.Safari", profileName: nil, displayOrder: 0),  // renumbered from 1 → 0
    BrowserOrderEntry(bundleID: "com.google.Chrome", profileName: "Personal", displayOrder: 1)  // renumbered from 2 → 1
]
```

---

### User Drags to Reorder

**Trigger**: User drags a browser/profile in the Manage Browsers popup

**Action**:
1. Calculate new `displayOrder` values for all affected entries
2. Update `orderedEntries` array with new order
3. Update `lastModified` timestamp
4. Persist to UserDefaults immediately

**Example**:
```swift
// Before: Safari (0), Chrome (1), Firefox (2)
// User drags Firefox to position 0 (above Safari)

// After: Firefox (0), Safari (1), Chrome (2)
orderedEntries = [
    BrowserOrderEntry(bundleID: "org.mozilla.firefox", displayOrder: 0),  // moved from 2 → 0
    BrowserOrderEntry(bundleID: "com.apple.Safari", displayOrder: 1),     // shifted from 0 → 1
    BrowserOrderEntry(bundleID: "com.google.Chrome", displayOrder: 2)     // shifted from 1 → 2
]
```

---

### Reset to Default

**Trigger**: User clicks "Reset Order" button in Manage Browsers popup

**Action**:
1. Clear `orderedEntries` array
2. Reinitialize with alphabetical order from `BrowserDetector.detectBrowsers()`
3. Update `lastModified` timestamp
4. Persist to UserDefaults

**Example**:
```swift
// Before: User custom order (Firefox, Safari, Chrome)
// After reset: Alphabetical order (Brave, Chrome, Firefox, Safari)

let browsers = BrowserDetector.shared.detectBrowsers()  // Already alphabetically sorted
orderedEntries = browsers.enumerated().map { index, browser in
    BrowserOrderEntry(bundleID: browser.bundleID, profileName: nil, displayOrder: index)
}
```

---

## Validation Rules

### BrowserOrderEntry Validation

1. **Non-empty bundleID**: `bundleID.isEmpty == false`
2. **Non-negative displayOrder**: `displayOrder >= 0`
3. **Valid profile name**: If `profileName != nil`, must be non-empty string

### BrowserOrderState Validation

1. **Unique displayOrder**: No two entries should have the same `displayOrder`
   - If duplicates found, sort by `displayOrder` and renumber sequentially
2. **Sequential displayOrder**: Should start at 0 and increment by 1
   - Gaps are allowed temporarily during drag operations, but cleaned up on save
3. **No orphaned profiles**: Every `profileName` must correspond to an actual browser profile
   - Validated against `BrowserProfileDetector` results
4. **No invalid browsers**: Every `bundleID` must correspond to an installed browser
   - Validated against `BrowserDetector` results

### Service Layer Enforcement

The `BrowserOrderService` is responsible for maintaining these invariants:
- On load: Clean up entries referencing uninstalled browsers
- On save: Renumber entries to ensure sequential `displayOrder`
- On browser refresh: Append new browsers, remove uninstalled ones
- On profile scan: Remove entries for deleted profiles

---

## Integration with Existing Systems

### Browser Visibility Integration

**Independent but Coordinated**:
- `BrowserOrderState` controls **position** (what order items appear)
- `BrowserVisibilityState` controls **visibility** (which browsers are shown)
- Both are respected when displaying browser lists

**Example**:
```swift
// Custom order: Chrome (0), Safari (1), Firefox (2), Brave (3)
// Visibility: Chrome hidden, Safari visible, Firefox visible, Brave visible

// Selection panel displays: Safari (key "1"), Firefox (key "2"), Brave (key "3")
// Chrome is in order but not shown (visibility = false)
```

**UI Behavior**:
- Manage Browsers popup shows **all browsers** in custom order (including hidden ones)
- Selection panel shows **only visible browsers** in custom order
- Keyboard shortcuts skip hidden browsers (sequential numbering of visible items only)

---

### Profile Features Integration

**Conditional Behavior**:
- If `settings.profileFeaturesEnabled == true`: Support profile-level entries
- If `settings.profileFeaturesEnabled == false`: Ignore `profileName`, treat as browser-only

**Example**:
```swift
// Profile features enabled
orderedEntries = [
    BrowserOrderEntry(bundleID: "com.google.Chrome", profileName: "Work", displayOrder: 0),
    BrowserOrderEntry(bundleID: "com.apple.Safari", profileName: nil, displayOrder: 1),
    BrowserOrderEntry(bundleID: "com.google.Chrome", profileName: "Personal", displayOrder: 2)
]
// Selection panel shows: Chrome Work (1), Safari (2), Chrome Personal (3)

// Profile features disabled
// Selection panel shows: Chrome (1), Safari (2)
// Profile-specific entries are collapsed to browser level
```

---

## Keyboard Shortcut Mapping

**Algorithm**:
1. Filter `orderedEntries` by browser visibility (remove hidden browsers)
2. If profile features enabled: Keep all visible entries
3. If profile features disabled: Group by `bundleID`, take first entry per browser
4. Sort by `displayOrder`
5. Assign keyboard shortcuts 1-9 to first 9 items
6. Ignore items beyond position 9

**Example**:
```swift
// orderedEntries (all visible):
// [Chrome Work (0), Safari (1), Chrome Personal (2), Firefox (3), Brave (4)]

// Profile features enabled:
// Key "1" → Chrome Work
// Key "2" → Safari
// Key "3" → Chrome Personal
// Key "4" → Firefox
// Key "5" → Brave

// Profile features disabled:
// Key "1" → Chrome (Work profile)
// Key "2" → Safari
// Key "3" → Firefox
// Key "4" → Brave
```

---

## Persistence Format

### UserDefaults Key
```swift
private enum Keys {
    static let browserOrder = "browserOrder"
}
```

### JSON Structure
```json
{
  "orderedEntries": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "bundleID": "com.google.Chrome",
      "profileName": "Work",
      "displayOrder": 0
    },
    {
      "id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
      "bundleID": "com.apple.Safari",
      "profileName": null,
      "displayOrder": 1
    }
  ],
  "lastModified": "2025-11-03T10:30:00Z",
  "version": 1
}
```

---

## Migration Path

### Version 1 → Future Versions

**Current Version**: 1 (initial implementation)

**Future Migration Support**:
- `version` field enables forward compatibility
- Migration logic in `PersistenceManager.loadBrowserOrder()`
- Pattern: Check `version`, apply transformations, update `version` field

**Example Future Migration** (hypothetical v2):
```swift
func loadBrowserOrder() -> BrowserOrderState {
    guard let data = userDefaults.data(forKey: Keys.browserOrder) else {
        return BrowserOrderState()  // Default empty state
    }

    do {
        var state = try JSONDecoder().decode(BrowserOrderState.self, from: data)

        // Hypothetical v1 → v2 migration
        if state.version == 1 {
            // Add new field, transform structure, etc.
            state = migrateV1toV2(state)
            saveBrowserOrder(state)  // Persist upgraded version
        }

        return state
    } catch {
        AppLogger.log("Error loading browser order: \(error)", level: .error)
        return BrowserOrderState()  // Graceful degradation
    }
}
```

---

## Performance Characteristics

**Storage Size**:
- Each entry: ~100-150 bytes (UUID + bundleID + optional profileName + displayOrder)
- Typical user: 5-10 browsers = ~1KB
- Maximum (50 browsers): ~7.5KB

**Read/Write Performance**:
- Load from UserDefaults: <1ms
- JSON decoding: <1ms
- Save to UserDefaults: <2ms
- Total roundtrip: <5ms

**Memory**:
- In-memory representation: ~1KB for typical user
- No caching needed (load on demand from UserDefaults)

---

## Testing Considerations

### Unit Test Coverage

**BrowserOrderEntry**:
- UUID uniqueness across multiple initializations
- `toBrowserTarget()` conversion correctness
- Codable encoding/decoding roundtrip

**BrowserOrderState**:
- Empty state initialization
- Codable persistence roundtrip
- `lastModified` timestamp updates
- Version field preservation

**Validation**:
- Invalid bundleID (empty string) rejection
- Negative displayOrder rejection
- Duplicate displayOrder detection and cleanup
- Orphaned profile removal

### Integration Test Coverage

**State Transitions**:
- Browser uninstall removes entries and renumbers
- Browser install appends to end
- Profile deletion removes specific entries
- Drag-and-drop updates order correctly
- Reset clears and reinitializes

**Browser Visibility Integration**:
- Hidden browsers excluded from keyboard shortcuts
- Hidden browsers retained in custom order
- Visibility changes don't affect order

**Profile Features Integration**:
- Profile-level entries work when features enabled
- Profile-level entries ignored when features disabled
- Toggling features preserves profile order data

---

## Open Questions

None. All design decisions finalized in [research.md](./research.md).
