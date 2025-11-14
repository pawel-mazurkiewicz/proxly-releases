# Data Model: Safari Profile Support

**Feature**: Safari Profile Support
**Branch**: `003-safari-profile`
**Date**: 2025-11-04

## Overview

This document defines the data structures for Safari profile management in Proxly. The model extends the existing BrowserProfile system while adding Safari-specific entities for profile configuration, window detection, and orphaned profile lifecycle management.

## Core Entities

### 1. SafariProfile

**Purpose**: Represents a user-configured Safari profile with manual name/order configuration.

**Fields**:

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `id` | `UUID` | Required, unique | Primary identifier for the profile |
| `name` | `String` | Required, 1-50 chars, no em dash (U+2014) | User-defined profile name matching Safari window title |
| `order` | `Int` | Required, 0-9, unique among all profiles | Maps to keyboard shortcut Option+Cmd+Shift+[order], where 0 is always 'Personal' |
| `createdAt` | `Date` | Required | Timestamp when profile was created |
| `modifiedAt` | `Date` | Required | Timestamp when profile was last modified |

**Validations**:
- `name` must not be empty
- `name` must not exceed 50 characters
- `name` must not contain em dash character (U+2014)
- `order` must be between 0 and 9 inclusive
- `order` must be unique among all profiles
- Maximum 10 profiles allowed

**Relationships**:
- Referenced by `Rule.browserTarget` (via `BrowserTarget` struct)
- Stored in `AppSettings.safariProfiles: [SafariProfile]`
- Synced via CloudKit (uses existing sync infrastructure)

**Lifecycle Transitions**:
- **Create**: New profile with `createdAt = now`, `modifiedAt = now`
- **Edit**: Update existing profile in-place with new values, `modifiedAt = now`
- **Delete**: Remove profile from array

**JSON Example**:
```json
{
  "id": "A1B2C3D4-E5F6-7890-ABCD-EF1234567890",
  "name": "Work",
  "order": 1,
  "createdAt": "2025-11-04T10:30:00Z",
  "modifiedAt": "2025-11-04T10:30:00Z"
}
```

### 2. BrowserProfile (Extended)

**Purpose**: Existing model extended to support Safari profiles alongside existing browser profile types.

**New Safari-Specific Fields**:

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `safariProfileId` | `UUID?` | Optional, references `SafariProfile.id` | Links to configured Safari profile when browser is Safari |

**Updated Fields** (no changes to existing fields):
- `browserIdentifier`: Existing field, value is `"com.apple.Safari"` for Safari profiles
- `profileName`: Existing field, matches `SafariProfile.name` for display purposes
- `profilePath`: Existing field, not applicable to Safari (Safari profiles not file-based)

**Validation Rules**:
- If `browserIdentifier == "com.apple.Safari"` and `safariProfileId != nil`, must reference valid `SafariProfile`

**JSON Example**:
```json
{
  "browserIdentifier": "com.apple.Safari",
  "profileName": "Work",
  "profilePath": null,
  "safariProfileId": "A1B2C3D4-E5F6-7890-ABCD-EF1234567890"
}
```

### 3. SafariWindowInfo (Runtime Only)

**Purpose**: Represents detected Safari window information during routing operations. Not persisted.

**Fields**:

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `windowElement` | `AXUIElement` | Required | Accessibility API window reference |
| `title` | `String` | Required | Complete window title from Accessibility API |
| `profileName` | `String?` | Optional | Extracted profile name if title matches pattern "PROFILENAME — PAGE TITLE" |
| `pageTitle` | `String?` | Optional | Extracted page title (text after em dash) |

**Pattern Matching**:
- Split window title on first occurrence of em dash (U+2014) with space: " — "
- `profileName` = substring before first " — " (trimmed)
- `pageTitle` = substring after first " — " (trimmed)
- If no em dash found or pattern doesn't match, `profileName = nil`

**Example Matching**:
- `"Work — Apple"` → `profileName: "Work"`, `pageTitle: "Apple"`
- `"Personal — GitHub — Pull Requests"` → `profileName: "Personal"`, `pageTitle: "GitHub — Pull Requests"`
- `"Safari"` → `profileName: nil`, `pageTitle: nil`
- `"Work — "` → `profileName: "Work"`, `pageTitle: ""`

### 4. AppSettings (Extended)

**Purpose**: Application-wide configuration, extended to include Safari profile storage.

**New Safari-Specific Fields**:

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `safariProfiles` | `[SafariProfile]` | Required, default `[]` | Array of all Safari profiles |

**Storage**:
- Persisted in `UserDefaults.standard` with key `"appSettings"`
- JSON-encoded with `JSONEncoder`
- Synced via CloudKit (existing sync infrastructure)

**Migration**:
- Existing `AppSettings` without `safariProfiles` field will decode with empty array (default value)
- No schema migration required (backward compatible)

## Data Flow

### Profile Configuration Flow

1. **User creates Safari profile**:
   - UI validates input (name, order, em dash check)
   - `SafariProfileManager.create(name:order:)` called
   - New `SafariProfile` created with `isOrphaned = false`
   - Appended to `AppSettings.safariProfiles`
   - `PersistenceManager.saveSettings()` called
   - CloudKit sync triggered (debounced)

2. **User edits Safari profile**:
   - UI pre-fills existing profile data
   - User changes name or order
   - UI validates new input
   - `SafariProfileManager.edit(id:newName:newOrder:)` called
   - Profile updated in-place with new values and `modifiedAt = now`
   - `PersistenceManager.saveSettings()` called
   - CloudKit sync triggered

3. **User deletes Safari profile**:
   - User selects profile for deletion
   - Confirmation dialog shown if profile referenced by rules
   - `SafariProfileManager.delete(id:)` called
   - Profile removed from `AppSettings.safariProfiles`
   - Rules referencing deleted profile remain valid but fallback to default Safari
   - `PersistenceManager.saveSettings()` called
   - CloudKit sync triggered

### URL Routing Flow

1. **URL received by Proxly**:
   - `URLProcessor.processURL(_:opener:)` called
   - `URLProcessingEngine.findBestRule()` evaluates rules
   - Rule matches with `browserTarget.browserIdentifier == "com.apple.Safari"`
   - Rule has `browserTarget.safariProfileId` set

2. **Safari profile lookup**:
   - `SafariProfileManager.getProfile(id:)` called
   - Returns `SafariProfile` if found

3. **Window detection**:
   - `SafariWindowDetector.enumerateWindows()` called
   - Returns `[SafariWindowInfo]` with extracted profile names
   - Match `SafariProfile.name` against `SafariWindowInfo.profileName`
   - If match found, return `SafariWindowInfo.windowElement`

4. **Profile window creation** (if no existing window):
   - `SafariProfileLauncher.createProfileWindow(order:)` called
   - Simulates keyboard shortcut Option+Cmd+Shift+[order]
   - Waits up to 2 seconds for window creation
   - Re-enumerates windows to detect new profile window
   - If timeout or creation fails, falls back to default Safari

## Validation Rules

### Profile Name Validation

```swift
func validateProfileName(_ name: String) -> ValidationResult {
    if name.isEmpty {
        return .failure("Profile name cannot be empty")
    }
    if name.count > 50 {
        return .failure("Profile name cannot exceed 50 characters")
    }
    if name.unicodeScalars.contains(where: { $0 == "\u{2014}" }) {
        return .failure("Profile name cannot contain the em dash character (—)")
    }
    return .success
}
```

### Profile Order Validation

```swift
func validateProfileOrder(_ order: Int, excludingProfileId: UUID? = nil) -> ValidationResult {
    if order < 0 || order > 9 {
        return .failure("Profile order must be between 0 and 9")
    }

    let profiles = AppSettings.shared.safariProfiles.filter {
        $0.id != excludingProfileId
    }

    if profiles.contains(where: { $0.order == order }) {
        return .failure("Order \(order) is already assigned to another profile")
    }

    return .success
}
```

### Profile Count Validation

```swift
func validateProfileCount() -> ValidationResult {
    let count = AppSettings.shared.safariProfiles.count

    if count >= 10 {
        return .failure("Maximum 10 Safari profiles allowed (due to keyboard shortcut limitation)")
    }

    return .success
}
```

## Migration Strategy

### Phase 1: Initial Implementation (Current Feature)

**Schema Changes**:
- Add `SafariProfile` struct (new)
- Add `safariProfiles: [SafariProfile]` to `AppSettings` (backward compatible, default `[]`)
- Add `safariProfileId: UUID?` to `BrowserProfile` (backward compatible, default `nil`)

**Data Migration**:
- Existing `AppSettings` decode with empty `safariProfiles` array
- Existing `BrowserProfile` instances decode with `safariProfileId = nil`
- No migration code required (uses default values)


## Performance Considerations

### Data Access Patterns

**Frequent Operations** (optimize for speed):
- Profile lookup by ID: O(n) linear scan (acceptable for ≤10 profiles)
- Profile enumeration: O(n) iteration
- Order conflict check: O(n) linear scan

**Infrequent Operations** (acceptable overhead):
- Profile creation: Append to array + save + sync
- Profile editing: Update in-place + save + sync
- Profile deletion: Remove from array + save + sync

### Memory Footprint

**Per SafariProfile** (estimated):
- UUID (16 bytes) + String (avg 20 bytes) + Int (8 bytes) + 2 × Date (16 bytes) = ~76 bytes
- Expected max: 10 profiles = ~760 bytes
- Negligible memory impact

### CloudKit Sync

**Sync Behavior**:
- Safari profiles embedded in `AppSettings` object
- Single CloudKit record per device (`AppSettings`)
- Last-write-wins conflict resolution based on `lastModified` timestamp
- Full profile array synced on each change (no delta sync)

**Sync Frequency**:
- Debounced/throttled via existing `PersistenceManager` patterns
- Typical: 1-2 seconds after last change
- Conflict resolution automatic (last-write-wins)

## Security & Privacy

**Data Protection**:
- Safari profile names stored locally in `UserDefaults.standard` (not encrypted at rest)
- CloudKit sync uses Apple's end-to-end encryption (CKRecord in private database)
- No sensitive URL content stored in profiles (only profile names/orders)

**Access Control**:
- Profile configuration requires no special permissions
- Profile routing requires macOS Accessibility permissions (for window detection)
- Permission errors caught and user shown guidance to System Settings

**Privacy Considerations**:
- Profile names may reveal user's work/personal context (acceptable, user-defined)
- Window title pattern matching reads Safari window titles (necessary for feature, no storage)
- No URL logging beyond existing Proxly debug logging patterns

## Testing Strategy

### Unit Tests

1. **ProfileNameValidationTests**:
   - Empty name rejected
   - Name > 50 chars rejected
   - Name with em dash rejected
   - Valid names accepted (letters, numbers, spaces, emoji, hyphens, en dash)

2. **ProfileOrderValidationTests**:
   - Order < 0 rejected
   - Order > 9 rejected
   - Order 0 accepted (reserved for 'Personal' profile)
   - Duplicate order rejected
   - Order unique after excluding specific profile ID (for edits)

3. **ProfileLifecycleTests**:
   - Create profile
   - Edit profile updates in-place
   - Delete removes profile
   - Profile count limit enforced (max 10)

4. **WindowPatternMatchingTests**:
   - "Work — Apple" extracts "Work" and "Apple"
   - "Personal — GitHub — Pull Requests" extracts correctly
   - "Safari" returns nil profile name
   - "Work — " (empty page title) handled correctly

### Integration Tests

1. **ProfileConfigurationIntegrationTest**:
   - Create profile → Persist → Restart app → Load profile
   - Edit profile → Updates in-place → Persists changes
   - Delete profile → Removed from storage

2. **RuleIntegrationTest**:
   - Create rule with Safari profile → Save → Load → Profile reference preserved
   - Edit profile → Rule reference still valid → Routing works with updated profile
   - Delete profile → Rule reference becomes invalid → Routing falls back to default Safari

### Manual Testing

1. **UI Testing**:
   - Profile configuration form validation (real-time feedback)
   - Profile selector in rules shows all profiles
   - Dark mode and Light mode appearance

2. **Routing Testing**:
   - Configure Safari profile → Create rule → Open URL → Verify routing to correct window
   - Edit profile → Open URL → Verify routing to updated profile window
   - Delete profile → Open URL → Verify fallback to default Safari
