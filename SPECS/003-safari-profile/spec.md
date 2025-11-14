# Feature Specification: Safari Profile Support

**Feature Branch**: `003-safari-profile`
**Created**: 2025-11-04
**Status**: Draft
**Input**: User description: "User Story: Safari Profile Support - As a Proxly user who uses multiple Safari profiles for different contexts (work, personal, etc.), I want Proxly to route URLs to specific Safari profiles based on my rules so that I can maintain profile separation and have URLs automatically open in the correct Safari context without manual switching."

## Clarifications

### Session 2025-11-04

- Q: When a user edits a Safari profile name or order that's already referenced in rules, should changes apply immediately, require confirmation, or be blocked? → A: Changes create a new profile entry. Old profile becomes "orphaned" in rules but rules remain valid. User must manually update rules to point to new profile.
- Q: If Accessibility permissions are revoked mid-session (while Proxly is running), should the system check permissions before each routing attempt, show a persistent warning, or only handle the error when it occurs? → A: Check permissions only when routing errors occur (window enumeration fails). Display user-facing error with guidance to re-enable permissions in System Settings. Minimal performance overhead.
- Q: Should Safari profile names prohibit the em dash character (—) and other special characters that could interfere with window title pattern matching? → A: Prohibit only the em dash character (U+2014) in profile names with validation error. All other characters allowed.
- Q: When editing a profile's order to a number already assigned to another active profile, should the edit be blocked, should the other profile become orphaned, or should orphaned profiles not count toward the duplicate order restriction? → A: Orphaned profiles don't count toward duplicate order validation. Only actively configured (non-orphaned) profiles are checked for duplicate orders.
- Q: Should orphaned profiles appear in the profile selector when users create or edit rules, or should only actively configured profiles be shown? → A: Show both active and orphaned profiles, with orphaned profiles visually distinguished (grayed out or marked with label like "(legacy)").

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Configure Safari Profiles (Priority: P1)

A user who has multiple Safari profiles (e.g., Work, Personal, Client Projects) needs to register these profiles in Proxly so the application knows which profiles exist and can target them for URL routing.

**Why this priority**: This is the foundation for all Safari profile routing functionality. Without configured profiles, no profile-specific routing can occur. This represents the minimum viable product (MVP) that enables users to prepare for profile-based routing.

**Independent Test**: Can be fully tested by opening Proxly settings, navigating to browser configuration, adding Safari profile names (e.g., "Work", "Personal"), assigning them numeric positions (1-9), saving the configuration, and verifying persistence across app restarts. Delivers immediate value by establishing profile definitions that will be used for routing.

**Acceptance Scenarios**:

1. **Given** a user with no configured Safari profiles, **When** they navigate to browser profile settings, **Then** they see an option to add Safari profiles with name and order fields
2. **Given** a user adding a Safari profile, **When** they enter a profile name "Work" and assign order "1", **Then** the profile is saved and appears in the Safari profiles list
3. **Given** a user with configured Safari profiles, **When** they restart Proxly, **Then** all configured Safari profiles persist with their names and order
4. **Given** a user attempting to assign the same order number to two profiles, **When** they save the configuration, **Then** the system prevents duplicate order assignments and shows an error message
5. **Given** a user with 9 configured Safari profiles, **When** they attempt to add a 10th profile, **Then** the system prevents adding more profiles and explains the 9-profile limitation
6. **Given** a user editing an existing Safari profile's name or order, **When** they save the changes, **Then** the system creates a new profile entry with the updated values and marks the old profile as available for deletion, while existing rules continue to reference the original profile name
7. **Given** a user entering a profile name containing the em dash character (—), **When** they attempt to save, **Then** the system displays a validation error explaining that the em dash character is not allowed in profile names
8. **Given** a user editing a profile to assign an order number already used by an orphaned profile, **When** they save, **Then** the system allows the assignment because orphaned profiles don't count toward duplicate order validation

---

### User Story 2 - Select Safari Profile in Rules (Priority: P2)

A user creating or editing a URL routing rule needs to specify Safari with a particular profile as the target browser, enabling profile-aware routing decisions.

**Why this priority**: This connects profile configuration to the routing system, enabling users to define routing rules that target specific Safari profiles. Without this, configured profiles remain unused. This builds on P1 and creates actionable routing rules.

**Independent Test**: Can be fully tested by creating a new rule, selecting Safari as the browser, choosing a specific configured profile from a dropdown, saving the rule, and verifying the profile selection persists. Delivers value by enabling users to create profile-specific routing rules that will direct URLs appropriately.

**Acceptance Scenarios**:

1. **Given** a user creating a new domain rule, **When** they select Safari as the target browser, **Then** they see a profile selector showing all actively configured Safari profiles, orphaned profiles (visually distinguished with grayed-out appearance or "(legacy)" label), plus an "Any Profile" option
2. **Given** a user with Safari profiles "Work" (order 1) and "Personal" (order 2), **When** they view the profile selector, **Then** active profiles appear in order (1: Work, 2: Personal), followed by any orphaned profiles
3. **Given** a user selecting profile "Work" for a rule, **When** they save and re-open the rule, **Then** the profile selection "Work" is retained
4. **Given** a user with no configured Safari profiles, **When** they select Safari as target browser, **Then** the profile selector shows only "Any Profile" with a prompt to configure profiles (orphaned profiles may still appear if they exist)
5. **Given** a user selecting "Any Profile", **When** they save the rule, **Then** URLs matching this rule open in Safari without profile-specific targeting
6. **Given** a user editing an existing rule that references an orphaned profile, **When** they view the profile selector, **Then** the orphaned profile appears as selected and is visually distinguished from active profiles

---

### User Story 3 - Route to Existing Safari Profile Window (Priority: P3)

When a URL matches a rule targeting a specific Safari profile, and a Safari window for that profile is already open, the system brings that window to focus, creates a new tab, and navigates to the URL—maintaining the user's profile context without manual window switching.

**Why this priority**: This is the primary use case that delivers the core value proposition. Users benefit from automatic profile-aware routing when their profile windows are already running. This is the most common scenario in daily usage.

**Independent Test**: Can be fully tested by (1) opening Safari with a specific profile window showing "Work — Some Page", (2) triggering a URL that matches a rule targeting the "Work" profile, (3) verifying the existing "Work" window comes to front, (4) confirming a new tab opens in that window with the routed URL. Delivers immediate workflow value by eliminating manual profile switching.

**Acceptance Scenarios**:

1. **Given** Safari is running with a window titled "Work — Dashboard", **When** a URL matching a rule targeting profile "Work" is processed, **Then** the "Work" window is brought to front and becomes active
2. **Given** the "Work" Safari window is now active, **When** the profile window is detected, **Then** a new tab is created in that window
3. **Given** a new tab created in the "Work" window, **When** the tab is ready, **Then** the routed URL is entered and navigation begins
4. **Given** multiple Safari windows exist including "Work — Page A" and "Personal — Page B", **When** a URL targeting "Personal" profile is processed, **Then** only the "Personal" window is activated and receives the new tab
5. **Given** Safari has a window titled "Work — " (empty page title), **When** a URL targeting "Work" profile is processed, **Then** the window is correctly identified by the "Work —" pattern and receives the URL
6. **Given** the routed URL navigation is complete, **When** the operation finishes, **Then** the entire process completes within 2 seconds from URL trigger to navigation start

---

### User Story 4 - Create New Safari Profile Window (Priority: P4)

When a URL matches a rule targeting a specific Safari profile, but no window for that profile exists, the system creates a new Safari window for that profile using the configured keyboard shortcut, then navigates to the URL—ensuring profile separation even when starting from scratch.

**Why this priority**: This handles the cold-start scenario when the target profile window isn't already running. While less common than P3 (most users keep profile windows open), it's essential for complete functionality and ensures profile routing works regardless of initial state.

**Independent Test**: Can be fully tested by (1) ensuring Safari is either not running or has no window for profile "Work" (order 1), (2) triggering a URL matching a rule targeting the "Work" profile, (3) verifying Safari launches (if needed) and a new window is created via Option+Cmd+Shift+1 keyboard shortcut, (4) confirming the routed URL opens in the new window. Delivers value for cold-start scenarios and ensures profile routing works consistently.

**Acceptance Scenarios**:

1. **Given** Safari is not running, **When** a URL targeting profile "Work" (order 1) is processed, **Then** Safari launches and the system waits for it to be ready
2. **Given** Safari is ready but no "Work" profile window exists, **When** the system needs to create a profile window for order 1, **Then** the keyboard shortcut Option+Cmd+Shift+1 is triggered
3. **Given** the keyboard shortcut was triggered, **When** Safari responds, **Then** the system waits for the new profile window to appear with a timeout of 2 seconds
4. **Given** a new "Work" profile window is created, **When** the window is detected and ready, **Then** the routed URL is entered and navigation begins
5. **Given** the keyboard shortcut fails to create a profile window within 2 seconds, **When** the timeout is reached, **Then** the system falls back to opening the URL in Safari without profile specification
6. **Given** Safari is not configured with a profile at the specified order position, **When** attempting to create a profile window, **Then** the system detects the failure and gracefully falls back to default Safari opening

---

### Edge Cases

- **What happens when a Safari window title contains the em dash character (—) multiple times?** The system uses the first occurrence of the pattern "PROFILENAME —" to identify the profile, matching only the profile name portion before the first em dash.

- **How does the system handle Safari window titles with special characters or emoji in the page title?** The profile detection pattern only examines the profile name portion before the em dash separator, so special characters in the page title do not affect profile window detection. Profile names themselves can contain any characters except the em dash (U+2014), which is prohibited during profile creation to prevent pattern matching ambiguity.

- **What happens when Safari is unresponsive or frozen during profile window creation?** If Safari does not respond to keyboard shortcuts or window creation within the 2-second timeout, the system falls back to opening the URL in default Safari without profile specification and logs the failure.

- **How does the system behave when the user has configured profile "Work" at order 1, but Safari has no profile configured at keyboard shortcut Option+Cmd+Shift+1?** The keyboard shortcut will fail to create the expected profile window. The system detects no matching window within the timeout period and falls back to opening the URL in Safari without profile specification.

- **What happens when multiple Safari windows exist for the same profile name?** The system brings the first detected window matching the profile pattern to front. If multiple windows match (e.g., "Work — Page A" and "Work — Page B"), the first match encountered during window enumeration receives the new tab.

- **How does the system handle the user deleting or renaming a configured Safari profile that is referenced in existing rules?** When a user edits a profile's name or order, the system creates a new profile entry, leaving the original profile as an "orphaned" entry. Existing rules continue to reference the orphaned profile by its original name and will function correctly if a Safari window with that profile name exists. The user must manually update rules to use the new profile name if desired. Orphaned profiles can be deleted by the user when no longer needed.

- **What happens when the user has 9 configured Safari profiles, uses all keyboard shortcuts 1-9, and later wants to reorder them?** The user can change the order assignments for existing profiles. Changing a profile from order 3 to order 5 means future window creation will use Option+Cmd+Shift+5 instead of Option+Cmd+Shift+3.

- **How does the system handle locale-specific window title patterns where the em dash (—) might be represented differently?** The initial implementation targets the standard em dash (U+2014) used in English locales. If Safari uses different separators in other locales, profile detection may fail and fall back to default Safari opening. This is documented as a known limitation.

- **What happens if Accessibility permissions are revoked while Proxly is running?** The system does not proactively check permissions before each routing attempt. When window enumeration fails due to missing permissions, the system displays a user-facing error message with guidance to re-enable Accessibility permissions in System Settings, then falls back to opening the URL in default Safari.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST allow users to manually enter Safari profile names (text input, maximum 50 characters)
- **FR-002**: System MUST allow users to assign a numeric order (1-9) to each Safari profile, mapping to keyboard shortcuts Option+Cmd+Shift+[1-9]
- **FR-003**: System MUST prevent duplicate order number assignments across actively configured (non-orphaned) Safari profiles and display an error message when attempted. Orphaned profiles do not count toward this validation.
- **FR-004**: System MUST limit actively configured (non-orphaned) Safari profiles to a maximum of 9 profiles due to keyboard shortcut limitations. Orphaned profiles do not count toward this limit.
- **FR-005**: System MUST persist configured Safari profile names and orders across application restarts
- **FR-005a**: System MUST create a new profile entry when a user edits an existing Safari profile's name or order, preserving the original profile so that existing rules continue to reference it by the original name
- **FR-005b**: System MUST allow users to delete Safari profiles, including orphaned profiles no longer actively configured but still referenced by rules
- **FR-006**: System MUST display both actively configured and orphaned Safari profiles in the browser profile selector when Safari is selected as a rule's target browser, with orphaned profiles visually distinguished (e.g., grayed out or labeled "(legacy)")
- **FR-007**: System MUST show active profiles in the profile selector ordered by their numeric order (1-9), followed by orphaned profiles
- **FR-008**: System MUST provide an "Any Profile" option in the profile selector to allow routing to Safari without profile-specific targeting
- **FR-009**: System MUST persist profile selection in saved rules and restore it when rules are edited
- **FR-010**: System MUST enumerate all Safari windows using macOS Accessibility API when processing a URL targeting a specific Safari profile
- **FR-011**: System MUST identify Safari profile windows by matching the pattern "PROFILENAME — " at the beginning of the window title (where PROFILENAME matches a configured profile)
- **FR-012**: System MUST bring the detected profile window to front using Accessibility API when a matching window is found
- **FR-013**: System MUST create a new tab in the active profile window by triggering the Cmd+T keyboard shortcut
- **FR-014**: System MUST enter the routed URL into the new tab's address bar and initiate navigation
- **FR-015**: System MUST complete the entire process (window detection → tab creation → URL navigation) within 2 seconds for existing profile windows
- **FR-016**: System MUST attempt to create a new Safari profile window using keyboard shortcut Option+Cmd+Shift+[N] (where N is the profile's configured order) when no matching profile window exists
- **FR-017**: System MUST wait up to 2 seconds for the new profile window to appear after triggering the keyboard shortcut
- **FR-018**: System MUST detect the newly created profile window by verifying the window title matches the expected profile name pattern
- **FR-019**: System MUST fall back to opening the URL in default Safari (without profile specification) when profile window creation fails or times out
- **FR-020**: System MUST fall back to opening the URL in default Safari when no matching profile window is detected and window creation fails
- **FR-021**: System MUST log profile routing attempts, successes, and fallback scenarios for debugging and monitoring
- **FR-022**: System MUST handle Safari not being installed or not responding by displaying an appropriate error message
- **FR-022a**: System MUST detect when Accessibility API calls fail due to missing or revoked permissions and display a user-facing error message with guidance to re-enable permissions in System Settings, then fall back to opening the URL in default Safari
- **FR-023**: System MUST isolate Safari-specific profile handling logic in dedicated modules/files separate from generic browser profile logic
- **FR-024**: System MUST validate Safari profile names to prevent empty names, enforce the 50-character maximum length, and prohibit the em dash character (U+2014) with a clear validation error message
- **FR-025**: System MUST integrate Safari profile ordering with the existing BrowserOrderService architecture
- **FR-026**: System MUST extend the BrowserProfile model to represent Safari profiles with name and order attributes
- **FR-027**: System MUST provide UI for Safari profile configuration that integrates with existing browser profile management interfaces

### Key Entities

- **Safari Profile**: Represents a user-configured Safari profile with a name (string, max 50 chars) and order (integer, 1-9). Used to map Proxly's profile configuration to Safari's keyboard shortcut-based profile system. Profiles can be in two states: actively configured (displayed in UI, counted toward 9-profile limit, validated for duplicate orders) or orphaned (created when a profile is edited, preserved for rules that reference the original name, not counted toward limits or duplicate order validation, can be manually deleted by user).

- **Profile Window**: Represents a detected Safari window that corresponds to a configured Safari profile. Identified by matching the window title pattern "PROFILENAME — PAGE TITLE" using the Accessibility API. Contains window reference for bringing to front and tab manipulation.

- **Profile Selector Configuration**: Represents the set of available Safari profiles displayed when configuring a rule. Includes all configured Safari profiles plus the special "Any Profile" option.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can successfully configure Safari profile names and orders, with 100% persistence across application restarts
- **SC-002**: Profile-specific URL routing succeeds in at least 95% of cases when the target profile window exists
- **SC-003**: New profile window creation completes within 2 seconds from keyboard shortcut trigger to window detection in 90% of successful cases
- **SC-004**: The complete routing flow (URL trigger to navigation start) completes within 2 seconds for existing profile windows in 95% of cases
- **SC-005**: The complete routing flow (URL trigger to navigation start) completes within 4 seconds for new profile window creation in 90% of cases
- **SC-006**: System gracefully falls back to default Safari opening in 100% of failure scenarios without crashes or hangs
- **SC-007**: Profile window detection correctly identifies the target profile window in at least 95% of cases when multiple Safari windows exist
- **SC-008**: Users can successfully create rules targeting Safari profiles and see them persist across rule edits in 100% of cases
- **SC-009**: Support tickets related to Safari profile routing failures or confusion are reduced by 80% compared to manual profile switching
- **SC-010**: Users with multiple Safari profiles report time savings of at least 30% on context switching between profiles (measured via user surveys)

## Assumptions

- **A-001**: Users understand Safari's profile concept and have already configured profiles within Safari itself before attempting to use this feature in Proxly
- **A-002**: Safari's window title pattern "PROFILENAME — PAGE TITLE" remains consistent across supported macOS versions (macOS 14.0+)
- **A-003**: The em dash character (U+2014) used in Safari window titles is consistent in English locales; other locales may require future updates
- **A-004**: Safari's keyboard shortcuts for profile window creation (Option+Cmd+Shift+[1-9]) remain stable across Safari versions
- **A-005**: The macOS Accessibility API provides sufficient permissions and capabilities to enumerate Safari windows and read window titles
- **A-006**: Users grant necessary Accessibility permissions to Proxly in System Settings for this feature to function
- **A-007**: Safari's response time to keyboard shortcuts and window creation is generally within 2 seconds on typical hardware
- **A-008**: The existing BrowserOrderService, BrowserProfile model, BrowserProfileDetector, and BrowserProfileSelector components can be extended to support Safari profiles without breaking existing functionality
- **A-009**: Profile names configured by users in Proxly will exactly match the profile names displayed in Safari window titles
- **A-010**: The current direct version of Proxly can use Accessibility API directly; future Mac App Store version will require moving this logic to ProxlyHelper

## Constraints

- **C-001**: Maximum of 9 Safari profiles supported due to keyboard shortcut limitation (Option+Cmd+Shift+[1-9])
- **C-002**: No official Safari API available for programmatic profile control; all interactions rely on reverse-engineered patterns
- **C-003**: Profile detection depends on window title pattern matching, which may be locale-dependent
- **C-004**: Accessibility API usage requires explicit user permission in macOS System Settings
- **C-005**: Future Mac App Store distribution will require migrating Accessibility API logic to ProxlyHelper due to sandbox restrictions
- **C-006**: Window detection and keyboard shortcut execution rely on Safari being the active or focusable application
- **C-007**: Implementation must remain isolated from generic browser profile logic to minimize impact on existing browser support
- **C-008**: AppleScript and Accessibility API calls introduce latency that may exceed 2-second targets on slower hardware or under system load

## Dependencies

- **D-001**: macOS Accessibility API for window enumeration, window title reading, and window activation
- **D-002**: macOS Accessibility permissions granted by user in System Settings → Privacy & Security → Accessibility
- **D-003**: Keyboard event simulation capability for triggering Option+Cmd+Shift+[1-9] and Cmd+T shortcuts
- **D-004**: AppleScript execution environment for Safari tab creation and URL navigation
- **D-005**: Existing BrowserOrderService.swift for integrating Safari profile ordering
- **D-006**: Existing BrowserProfile.swift model for extending with Safari profile representation
- **D-007**: Existing BrowserProfileDetector.swift for adapting to manual Safari profile input
- **D-008**: Existing BrowserProfileSelector.swift UI for extending with Safari profile selection
- **D-009**: Safari application must be installed on the user's system
- **D-010**: ProxlyHelper infrastructure for future Mac App Store version migration

## Out of Scope

The following items are explicitly excluded from this feature and may be considered for future iterations:

- **OS-001**: Automatic detection of Safari profile names from Safari's configuration or running processes (no API available)
- **OS-002**: Creating new Safari profiles from within Proxly
- **OS-003**: Deleting or modifying Safari profiles from within Proxly
- **OS-004**: Customizing Safari profile icons, colors, or visual appearance in Proxly
- **OS-005**: Support for Safari's Private Browsing mode as a profile target
- **OS-006**: Incognito or private mode support for specific profiles
- **OS-007**: Synchronizing Safari profile configuration across multiple Proxly-enabled devices
- **OS-008**: Detecting which Safari profile is currently active or in use
- **OS-009**: Automatic profile window creation based on Safari's internal profile state (beyond keyboard shortcuts)
- **OS-010**: Profile-specific history, bookmarks, or settings integration
- **OS-011**: Locale-specific window title pattern detection beyond standard English em dash (U+2014)
- **OS-012**: Profile window creation retry logic or advanced failure recovery beyond single fallback
