# Feature Specification: Customizable Browser & Profile Ordering with Keyboard Shortcuts

**Feature Branch**: `002-browser-order-shortcuts`
**Created**: 2025-11-03
**Status**: Draft
**Input**: User description: "Problem: I am used to a particular order of browsers from my previous solution (Browser Fairy) and would like to use the same number shortcut in the browser selector. At first I thought this was no problem, I just learn the new order. However, the second scenario is when I clicked 'Refresh Browser List' in the settings, the order of the browsers changed. Solution: In the Browser Visibility 'Manage Browsers' popup, enable the user to define the order of the browsers as well as the visibility of them. That way, I can choose which browser corresponds to pressing 1 or 2 etc when I select the browser to use. Bonus: Enable me to select which profile (for Chrome etc) corresponds to which keyboard number too, so I could have 1 for Chrome 'work' profile, 2 for Chrome 'personal' profile, 3 for Safari, etc."

## User Scenarios & Testing

### User Story 1 - Custom Browser Ordering (Priority: P1)

A user wants to maintain consistent keyboard shortcuts for browser selection that match their muscle memory from previous tools. They need to be able to rearrange the order of browsers in the selection panel so that pressing "1" always opens their preferred browser, "2" opens their second choice, etc., regardless of browser refresh operations.

**Why this priority**: This is the core problem statement from the user. Without this, the feature doesn't address the primary pain point of inconsistent keyboard shortcuts when the browser list changes. This forms the foundation for all other functionality.

**Independent Test**: Can be fully tested by opening the Manage Browsers popup, dragging browsers to reorder them, saving the order, then opening the browser selection panel and verifying that keyboard numbers 1-9 correspond to the new custom order. This delivers immediate value by solving the keyboard shortcut consistency problem.

**Acceptance Scenarios**:

1. **Given** the Manage Browsers popup is open, **When** the user drags and drops browsers to reorder them, **Then** the new order is saved and persisted across app restarts
2. **Given** a custom browser order has been set, **When** the user opens the browser selection panel via a URL, **Then** keyboard shortcuts 1-9 correspond to the user's custom order, not the system-detected order
3. **Given** a custom browser order exists, **When** the user clicks "Refresh Browser List" in settings, **Then** the custom order is preserved and newly detected browsers are appended to the end
4. **Given** a browser is uninstalled, **When** the user opens the browser selection panel, **Then** the remaining browsers maintain their relative order with keyboard shortcuts adjusted accordingly
5. **Given** browser visibility settings exist (some browsers hidden), **When** displaying the selection panel, **Then** only visible browsers are shown but in the custom order (keyboard shortcuts skip hidden browsers)

---

### User Story 2 - Custom Profile Ordering (Priority: P2)

A power user with multiple browser profiles wants to assign specific keyboard shortcuts to individual profiles across different browsers. For example, they want "1" for Chrome Work profile, "2" for Chrome Personal profile, "3" for Safari, "4" for Firefox Developer profile, etc.

**Why this priority**: This is an advanced feature that provides significant value for power users but requires more complex UI and data model changes. It builds on P1 by extending ordering control to the profile level. It can be developed and tested independently as a profile-aware enhancement.

**Independent Test**: Can be tested by opening the Manage Browsers popup, expanding a browser with multiple profiles, reordering profiles independently from browsers, then opening the browser selection panel and verifying that when profile selection is enabled, pressing number keys navigates through both browsers and their profiles in the custom order.

**Acceptance Scenarios**:

1. **Given** the Manage Browsers popup is open and a browser has multiple profiles, **When** the user expands the browser and drags profiles to reorder them within that browser's section, **Then** the profile order is saved and persisted
2. **Given** profile features are enabled and custom profile ordering exists, **When** the user opens the browser selection panel, **Then** keyboard shortcuts 1-9 map to a flat list of browser+profile combinations in the custom order
3. **Given** a user has set Chrome Work as position 1 and Safari as position 2, **When** the user presses "1" in the selection panel, **Then** Chrome opens with the Work profile
4. **Given** custom profile ordering is configured, **When** the user clicks a browser card to manually select a profile, **Then** profiles appear in the custom order within that browser's profile selection sheet

---

### Edge Cases

- What happens when a browser is uninstalled and it was in the middle of the custom order? (Answer: Remove from order, shift remaining browsers up, preserve relative positions)
- What happens when "Refresh Browser List" discovers a new browser? (Answer: Append to the end of the custom order to preserve existing shortcuts)
- What happens when all visible browsers are removed from the custom order? (Answer: Use existing visibility system validation - at least one browser must be visible)
- What happens when a user tries to reorder a browser that is currently hidden via visibility settings? (Answer: Allow reordering - order and visibility are independent settings, both are respected when displaying browsers)
- What happens when profile features are disabled but custom profile ordering exists? (Answer: Fall back to browser-level ordering only, preserve profile order data for when features are re-enabled)
- What happens when a browser profile is deleted externally (e.g., user deletes it from Chrome settings)? (Answer: Remove from custom order on next profile detection, shift remaining items up)
- What happens when keyboard shortcuts 1-9 are pressed for positions beyond the number of visible browsers? (Answer: No action)
- What happens when no custom order has been set yet? (Answer: Use default alphabetical order by display name until user customizes)
- What happens when dragging a browser over hidden browsers in the Manage Browsers list? (Answer: Allow insertion between hidden browsers - order applies to all browsers regardless of visibility)

## Requirements

### Functional Requirements

- **FR-001**: System MUST persist a user-defined browser ordering that overrides the default alphabetical ordering
- **FR-002**: System MUST allow users to reorder browsers via drag-and-drop in the Manage Browsers popup
- **FR-003**: System MUST maintain custom browser order across app restarts and system reboots
- **FR-004**: System MUST preserve custom browser order when "Refresh Browser List" is triggered
- **FR-005**: System MUST append newly detected browsers to the end of the custom order when refresh occurs
- **FR-006**: System MUST respect existing browser visibility settings when displaying the browser selection panel (show only visible browsers in custom order)
- **FR-007**: System MUST assign keyboard shortcuts 1-9 sequentially to visible browsers in custom order
- **FR-008**: System MUST allow reordering of both visible and hidden browsers in the Manage Browsers popup
- **FR-009**: System MUST allow users to reorder profiles within each browser independently
- **FR-010**: System MUST persist custom profile ordering across app restarts
- **FR-011**: System MUST support a flat ordering of browser+profile combinations for keyboard shortcuts when profile features are enabled
- **FR-012**: System MUST gracefully handle browsers being uninstalled by removing them from custom order
- **FR-013**: System MUST gracefully handle profiles being deleted externally by removing them from custom order
- **FR-014**: System MUST maintain relative ordering of remaining browsers when items are removed
- **FR-015**: System MUST display browsers in consistent custom order across all UI surfaces (selection panel, rule editor browser picker)
- **FR-016**: System MUST migrate existing users to an initial custom order based on alphabetical sort on first launch after upgrade
- **FR-017**: System MUST provide a visual indicator in the Manage Browsers popup showing the current position number for each browser
- **FR-018**: System MUST allow users to reset custom order to default (alphabetical) via a button in Manage Browsers popup

### Key Entities

- **BrowserOrderEntry**: Represents a browser's position in the custom order
  - bundleID (string): Unique identifier for the browser
  - displayOrder (integer): User-defined position (0-based index)
  - profileOrder (array of strings, optional): Ordered list of profile names for this browser

- **BrowserManagementState**: Persisted configuration for browser ordering
  - orderedBrowsers (array of BrowserOrderEntry): Complete custom ordering
  - lastRefreshDate (timestamp): When browser list was last refreshed for coordination purposes
  - version (integer): Schema version for future migration support

## Success Criteria

### Measurable Outcomes

- **SC-001**: Users can reorder browsers in under 30 seconds via drag-and-drop without consulting documentation
- **SC-002**: Custom browser order persists across 100% of app restarts and system reboots without data loss
- **SC-003**: Keyboard shortcuts 1-9 consistently map to the same browsers across all sessions after custom order is set
- **SC-004**: Browser refresh operations complete without disrupting existing custom order (newly detected browsers appended to end)
- **SC-005**: Ordering changes reflect immediately in the browser selection panel (no app restart required)
- **SC-006**: Profile-level ordering (P2) supports at least 9 browser+profile combinations with keyboard shortcuts
- **SC-007**: UI updates reflect ordering changes within 100ms of user action for responsive feel
- **SC-008**: Zero keyboard shortcut conflicts occur when mixing browsers and profiles in custom order
- **SC-009**: Migration from existing browser list to custom order completes automatically on first launch with no user action required
- **SC-010**: Users report muscle memory consistency with previous tools (measured via user satisfaction survey)

## Assumptions

- Browsers are uniquely identified by bundle ID, which remains stable across app updates
- Profile names within a browser are sufficiently unique to distinguish them
- Users will not need more than 9 total browser+profile combinations for keyboard shortcuts (positions 1-9)
- Default ordering for new users will be alphabetical by display name as a reasonable starting point
- The existing BrowserDetector service can provide stable browser enumeration as a basis for custom ordering
- Browser profile detection is already implemented in BrowserProfileDetector and provides profile names that can be used for ordering
- The existing BrowserManagementView can be extended with drag-and-drop capabilities for reordering
- Profile features are controlled by the existing `settings.profileFeaturesEnabled` flag
- Browser visibility is controlled by existing `settings.browserVisibility` and is independent from ordering
- Maximum of 50 browsers total in custom order (reasonable upper bound for any user)
- The existing visibility system (BrowserVisibilityState in AppSettings) already validates that at least one browser is visible

## Dependencies

- Existing BrowserDetector service for enumerating installed browsers
- Existing BrowserProfileDetector service for detecting browser profiles
- Existing BrowserManagementView UI that will be extended with reordering capabilities
- Existing PersistenceManager for storing custom order configuration
- Existing AppSettings model with browserVisibility and profileFeaturesEnabled
- SelectionPanel, ProfileSelectionView, and RuleEditView will need updates to respect custom ordering
- Existing Browser model with bundleID and displayName properties
- Existing BrowserVisibilityState for visibility management (independent from this feature)

## Out of Scope

- Modifying the existing browser visibility system (already implemented in BrowserManagementView)
- Custom keyboard shortcuts beyond 1-9 (e.g., letter keys, function keys, custom key bindings)
- Importing/exporting browser order configurations between devices or users
- Synchronizing browser order across devices via CloudKit (users set order independently per device)
- Reordering browsers by criteria other than manual drag-and-drop (e.g., sort by most used, recent, alphabetical buttons - except reset to default)
- Grouping browsers into categories or folders
- Customizing keyboard shortcuts for individual browsers (all use sequential numbers)
- Hiding specific profiles within a browser (visibility is browser-level only via existing system)
- Multi-level hierarchical ordering (e.g., browser groups with profiles nested under them)
- Undo/redo functionality for ordering changes
- Search or filter functionality within the Manage Browsers popup
