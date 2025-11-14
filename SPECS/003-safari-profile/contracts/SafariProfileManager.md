# Service Contract: SafariProfileManager

**Feature**: Safari Profile Support
**Service**: SafariProfileManager
**Responsibility**: Manage Safari profile CRUD operations

## Overview

SafariProfileManager is responsible for creating, reading, updating, and deleting Safari profiles. It enforces validation rules (name constraints, order uniqueness, profile limits) and coordinates with PersistenceManager for storage.

## Interface Definition

```swift
@MainActor
final class SafariProfileManager: ObservableObject {
    // Singleton instance
    static let shared = SafariProfileManager()

    // Published state for SwiftUI reactivity
    @Published private(set) var profiles: [SafariProfile] = []

    // MARK: - CRUD Operations

    /// Creates a new Safari profile with validation
    /// - Parameters:
    ///   - name: Profile name (1-50 chars, no em dash)
    ///   - order: Keyboard shortcut order (0-9, unique among all profiles, 0 is 'Personal')
    /// - Returns: Created profile or validation error
    /// - Throws: SafariProfileError if validation fails
    func createProfile(name: String, order: Int) async throws -> SafariProfile

    /// Retrieves a profile by ID
    /// - Parameter id: Profile UUID
    /// - Returns: Safari profile if found, nil otherwise
    func getProfile(id: UUID) -> SafariProfile?

    /// Retrieves all profiles sorted by order
    /// - Returns: Array of profiles sorted by order (0-9)
    func getProfiles() -> [SafariProfile]

    /// Edits an existing profile by updating it in-place
    /// - Parameters:
    ///   - id: ID of profile to edit
    ///   - newName: Updated profile name (validated)
    ///   - newOrder: Updated keyboard shortcut order (validated)
    /// - Returns: Updated profile with new values
    /// - Throws: SafariProfileError if validation fails or profile not found
    func editProfile(id: UUID, newName: String, newOrder: Int) async throws -> SafariProfile

    /// Deletes a profile
    /// - Parameter id: Profile UUID to delete
    /// - Returns: True if deleted, false if not found
    /// - Throws: SafariProfileError if deletion fails
    func deleteProfile(id: UUID) async throws -> Bool

    // MARK: - Validation

    /// Validates profile name constraints
    /// - Parameter name: Profile name to validate
    /// - Returns: Validation result (success or failure with message)
    func validateName(_ name: String) -> ValidationResult

    /// Validates profile order constraints
    /// - Parameters:
    ///   - order: Profile order to validate (0-9)
    ///   - excludingId: Optional profile ID to exclude from duplicate check (for edits)
    /// - Returns: Validation result (success or failure with message)
    func validateOrder(_ order: Int, excludingId: UUID?) -> ValidationResult

    /// Validates profile count limit (max 10)
    /// - Returns: Validation result (success or failure with message)
    func validateProfileCount() -> ValidationResult

    // MARK: - Bulk Operations

    /// Retrieves count of rules referencing a specific profile
    /// - Parameter profileId: Safari profile UUID
    /// - Returns: Number of rules referencing this profile
    func getRuleReferenceCount(for profileId: UUID) -> Int
}
```

## Data Types

### SafariProfile

```swift
struct SafariProfile: Codable, Identifiable, Equatable {
    let id: UUID
    var name: String
    var order: Int
    var createdAt: Date
    var modifiedAt: Date
}
```

### ValidationResult

```swift
enum ValidationResult {
    case success
    case failure(String)

    var isValid: Bool {
        if case .success = self { return true }
        return false
    }

    var errorMessage: String? {
        if case .failure(let message) = self { return message }
        return nil
    }
}
```

### SafariProfileError

```swift
enum SafariProfileError: LocalizedError {
    case invalidName(String)
    case invalidOrder(String)
    case profileNotFound(UUID)
    case profileLimitReached
    case persistenceFailure(Error)

    var errorDescription: String? {
        switch self {
        case .invalidName(let message): return message
        case .invalidOrder(let message): return message
        case .profileNotFound(let id): return "Profile not found: \(id)"
        case .profileLimitReached: return "Maximum 10 Safari profiles allowed"
        case .persistenceFailure(let error): return "Failed to save profile: \(error.localizedDescription)"
        }
    }
}
```

## Method Specifications

### createProfile(name:order:)

**Purpose**: Create new active Safari profile with validation

**Preconditions**:
- `name` validated (non-empty, ≤50 chars, no em dash)
- `order` validated (0-9, unique among all profiles)
- Profile count < 10

**Process**:
1. Validate name using `validateName(_:)`
2. Validate order using `validateOrder(_:excludingId:)` with `excludingId = nil`
3. Validate profile count using `validateProfileCount()`
4. Create new `SafariProfile` with `createdAt = now`, `modifiedAt = now`
5. Append to `AppSettings.safariProfiles`
6. Call `PersistenceManager.saveSettings()`
7. Trigger CloudKit sync (debounced)
8. Update `@Published profiles` property
9. Return created profile

**Postconditions**:
- New profile added to storage
- Profile accessible via `getProfile(id:)` and `getProfiles()`
- CloudKit sync triggered

**Errors**:
- `SafariProfileError.invalidName`: Name validation failed
- `SafariProfileError.invalidOrder`: Order validation failed
- `SafariProfileError.profileLimitReached`: Already 10 profiles
- `SafariProfileError.persistenceFailure`: Save failed

**Performance**: O(n) where n = total profile count (for validation checks)

### editProfile(id:newName:newOrder:)

**Purpose**: Edit existing profile by updating it in-place

**Preconditions**:
- Profile with `id` exists
- `newName` validated (same rules as create)
- `newOrder` validated (0-9, unique among all profiles, excluding current profile)

**Process**:
1. Lookup profile using `getProfile(id:)`
2. Throw `SafariProfileError.profileNotFound` if not found
3. Validate `newName` using `validateName(_:)`
4. Validate `newOrder` using `validateOrder(_:excludingId:)` with `excludingId = id`
5. Update profile in-place: `name = newName`, `order = newOrder`, `modifiedAt = now`
6. Call `PersistenceManager.saveSettings()`
7. Trigger CloudKit sync (debounced)
8. Update `@Published profiles` property
9. Return updated profile

**Postconditions**:
- Profile updated in-place with new values
- Rules referencing this profile continue to work with updated profile
- CloudKit sync triggered

**Errors**:
- `SafariProfileError.profileNotFound`: Profile with `id` doesn't exist
- `SafariProfileError.invalidName`: Name validation failed
- `SafariProfileError.invalidOrder`: Order validation failed
- `SafariProfileError.persistenceFailure`: Save failed

**Performance**: O(n) where n = total profile count (for validation)

### deleteProfile(id:)

**Purpose**: Remove profile from storage

**Preconditions**:
- Profile with `id` exists

**Process**:
1. Lookup profile in `AppSettings.safariProfiles`
2. If not found, return `false`
3. Remove profile from array
4. Call `PersistenceManager.saveSettings()`
5. Trigger CloudKit sync (debounced)
6. Update `@Published profiles` property
7. Return `true`

**Postconditions**:
- Profile removed from storage
- Profile no longer accessible via any getter methods
- Rules referencing deleted profile remain but routing falls back to default Safari
- CloudKit sync triggered

**Errors**:
- `SafariProfileError.persistenceFailure`: Save failed

**Performance**: O(n) where n = total profile count

### Validation Methods

#### validateName(_:)

**Rules**:
1. Name cannot be empty
2. Name cannot exceed 50 characters
3. Name cannot contain em dash character (U+2014)

**Returns**: `.success` or `.failure(message)`

#### validateOrder(_:excludingId:)

**Rules**:
1. Order must be between 0 and 9 inclusive
2. Order must be unique among all profiles
3. If `excludingId` provided, exclude that profile from duplicate check (for edit operations)

**Returns**: `.success` or `.failure(message)`

#### validateProfileCount()

**Rules**:
1. Count of profiles must be < 10

**Returns**: `.success` or `.failure(message)`

## State Management

**Initialization**:
- Singleton instantiated on first access
- Loads profiles from `AppSettings.safariProfiles` in `init()`
- Subscribes to `AppSettings` changes for reactivity

**Published State**:
- `profiles: [SafariProfile]` - All profiles (active + orphaned) for UI binding
- Updated after every mutation operation
- SwiftUI views observe and react automatically

**Concurrency**:
- `@MainActor` isolated (all methods run on main thread)
- Async operations for persistence (non-blocking)
- CloudKit sync triggered asynchronously (existing infrastructure)

## Integration Points

**Dependencies**:
- `PersistenceManager`: Save/load `AppSettings`
- `CloudKitSyncManager`: Trigger sync on profile changes (via `PersistenceManager`)
- `Rule` model: Profiles referenced via `BrowserTarget.safariProfileId`

**Observers**:
- SwiftUI views bind to `@Published profiles` property
- `BrowserProfileSelector` uses `getProfiles()`
- `SafariProfileConfigView` uses all CRUD methods

## Error Handling Strategy

**User-Facing Errors**:
- Validation failures: Show inline error in UI form
- Profile not found: Show alert dialog
- Profile limit reached: Show alert with explanation of 10-profile limitation
- Persistence failure: Show alert with retry option

**Developer Errors**:
- Invalid UUID: Precondition failure (programming error)
- Concurrent modification: Not possible (`@MainActor` serializes access)

**Logging**:
- All CRUD operations logged at `.info` level
- Validation failures logged at `.debug` level
- Persistence failures logged at `.error` level

## Testing Strategy

**Unit Tests**:
1. Create profile with valid input → success
2. Create profile with invalid name → throws `SafariProfileError.invalidName`
3. Create profile with duplicate order → throws `SafariProfileError.invalidOrder`
4. Create 11th profile → throws `SafariProfileError.profileLimitReached`
5. Create profile with order 0 (Personal) → success
6. Edit profile → profile updated in-place
7. Edit with duplicate order → throws error
8. Delete profile → removed from storage
9. Delete non-existent profile → returns false

**Integration Tests**:
1. Create → Save → Restart app → Load → Profile persists
2. Edit → Save → Restart → Profile persists with updated values
3. Delete profile → Rules referencing it remain valid, routing falls back to default Safari

**Mock Dependencies**:
- `MockPersistenceManager`: Verify save calls without disk I/O
- `MockCloudKitSyncManager`: Verify sync triggers without network I/O
