# Feature Specification: Custom URL Transformations

**Feature Branch**: `001-custom-url-transform`
**Created**: 2025-10-31
**Status**: Draft
**Input**: User description: "add an ability for user to create their own custom transformations for unsupported applications (so outside of @Proxly/Services/NativeAppPresetsService.swift) - this can also be a general feature that just allows user to create custom URL transformations for whatever they're matching in the Condition. I would assume we could use some sort of scripting language for this, like javascript"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Create Basic Template Transformation (Priority: P1)

A user wants to automatically strip tracking parameters from URLs before opening them in their browser. They create a custom transformation that removes `utm_*` and `fbclid` parameters from any URL, keeping only the clean URL. For example, `https://example.com/article?utm_source=twitter&utm_campaign=promo&id=123` becomes `https://example.com/article?id=123`.

**Why this priority**: This is the core MVP functionality - demonstrates the general-purpose nature of URL transformations. Can be used for privacy, URL cleaning, or any arbitrary transformation need. Delivers immediate value and can stand alone.

**Independent Test**: Can be fully tested by creating one template-based transformation rule, applying it to URLs with various parameters, and verifying the output matches expectations. Success means parameters are correctly manipulated.

**Acceptance Scenarios**:

1. **Given** user is on the rule creation screen, **When** they select "Custom Transformation" as the action type and define transformation logic, **Then** the transformation is saved and appears in their rules list
2. **Given** a custom transformation rule exists for removing tracking parameters, **When** user opens a URL with tracking parameters, **Then** the URL is cleaned and opened in the designated browser
3. **Given** user creates a transformation with invalid syntax, **When** they attempt to save, **Then** system shows validation error with specific feedback about what needs correction

---

### User Story 2 - Use JavaScript for Complex Logic (Priority: P2)

A user needs to redirect all Reddit mobile links (`https://m.reddit.com/...`) to old Reddit desktop format (`https://old.reddit.com/...`) but only if the path starts with `/r/` (subreddit pages). They write a JavaScript transformation that checks the path pattern and conditionally transforms the URL, leaving other Reddit pages unchanged.

**Why this priority**: Extends the basic transformation capability to handle conditional logic and complex URL manipulations that simple pattern replacement can't handle. Still independently valuable for power users who need programmatic control.

**Independent Test**: Can be tested by creating a JavaScript-based transformation, providing various sample URLs (matching and non-matching patterns), previewing the output, and verifying the conditional logic works correctly. Success means JavaScript executes correctly with proper conditional behavior.

**Acceptance Scenarios**:

1. **Given** user selects "JavaScript Transformation" mode, **When** they write a JavaScript function that takes URL components as input and returns a transformed URL, **Then** they can save and apply this transformation
2. **Given** a JavaScript transformation with conditional logic exists, **When** system processes URLs matching and not matching the condition, **Then** only matching URLs are transformed while others pass through unchanged
3. **Given** user's JavaScript contains syntax errors, **When** they test the transformation, **Then** system displays error message with line number and description
4. **Given** JavaScript execution exceeds time limit, **When** transformation runs, **Then** system falls back to opening URL unchanged and logs timeout error

---

### User Story 3 - Test and Debug Transformations (Priority: P1)

A user creates a custom transformation for redirecting Medium paywalled articles through a proxy service. Before enabling it on all Medium URLs, they use a built-in testing interface to input sample URLs and verify the transformation produces correct proxy URLs with proper encoding.

**Why this priority**: Essential for usability - without testing capability, users would need to save and trigger actual URLs to verify transformations work. This is critical for P1 because transformation errors could break workflows or produce malformed URLs.

**Independent Test**: Can be tested by creating a transformation in draft mode, entering various test URLs, and observing real-time preview of transformed outputs. Success means preview accurately shows what will happen when rule is active.

**Acceptance Scenarios**:

1. **Given** user is creating/editing a transformation rule, **When** they enter a test URL in the preview field, **Then** system immediately shows the transformed output or validation error
2. **Given** transformation contains errors, **When** user tests it, **Then** system highlights the specific error location and provides helpful correction suggestions
3. **Given** user tests multiple URLs with edge cases, **When** transformation succeeds for all, **Then** they can save with confidence knowing the rule works correctly

---

### User Story 4 - Template-Based Transformations with Variables (Priority: P2)

A user wants to redirect all Twitter/X.com links through Nitter (privacy-focused Twitter frontend). They create a template transformation that replaces the host while preserving the entire path and query parameters: `https://twitter.com/user/status/123` → `https://nitter.net/user/status/123`. The template uses variables like `{path}` and `{query}` to preserve URL structure.

**Why this priority**: Demonstrates the power of template variables for common transformation patterns. Reduces complexity compared to JavaScript for simple host/scheme replacements. Can be added after basic transformation infrastructure works.

**Independent Test**: Can be tested by creating a template-based transformation with variable substitution, applying it to multiple URLs with different paths/queries, and verifying placeholders are correctly substituted. Success means template variables expand correctly while preserving URL structure.

**Acceptance Scenarios**:

1. **Given** user creates a transformation with template syntax using `{host}`, `{path}`, `{query}` variables, **When** system processes URL `https://example.com/page/123?foo=bar`, **Then** variables are correctly substituted in the output
2. **Given** user creates host replacement template `https://nitter.net{path}{query}`, **When** URL `https://twitter.com/user/status/123?ref=home` is processed, **Then** output is `https://nitter.net/user/status/123?ref=home`
3. **Given** template references undefined variable, **When** user tests transformation, **Then** system shows warning about missing or invalid variable

---

### User Story 5 - Share and Import Transformations (Priority: P3)

A user discovers a useful transformation pattern in a community forum (e.g., YouTube → Invidious redirect) and wants to use it. They copy a JSON configuration snippet and import it, which automatically creates the rule in their Proxly instance without manual configuration.

**Why this priority**: Nice-to-have community feature that encourages knowledge sharing. Not essential for core functionality but valuable for ecosystem growth and reducing setup friction.

**Independent Test**: Can be tested by exporting an existing transformation as JSON, importing it into a fresh Proxly instance (or different machine), and verifying the rule works identically. Success means seamless transfer of transformation configurations.

**Acceptance Scenarios**:

1. **Given** user has a working transformation, **When** they select "Export" option, **Then** system generates shareable JSON containing the transformation logic
2. **Given** user receives a transformation JSON from community, **When** they import it, **Then** system validates the configuration and creates an enabled/disabled rule based on user preference
3. **Given** imported transformation conflicts with existing rule, **When** import occurs, **Then** system offers options to rename, replace, or skip

---

### Edge Cases

- What happens when JavaScript transformation throws an uncaught exception? (System should catch, log error, and fall back to opening original URL unchanged with user notification)
- How does system handle transformation that produces invalid URL syntax? (Validate output format, show error to user, offer to open original URL instead)
- What happens when user creates circular transformation (e.g., transforms URL back to itself)? (Detect loops, prevent infinite recursion, show error during validation)
- How does system handle transformations attempting to access network or file system? (Sandbox JavaScript environment to prevent external access, only allow URL string manipulation)
- What happens when transformed URL has custom scheme not registered on system? (Open as-is and let macOS handle - may prompt user to find app or show error)
- How does system handle very slow JavaScript execution? (Implement timeout of 1 second, kill execution, fall back to original URL unchanged)
- What happens when transformation is applied to non-HTTP(S) URL? (Allow transformations on any URL scheme - user decides what to transform)
- What happens when template variable references non-existent URL component? (Substitute with empty string or show validation warning depending on strictness mode)
- How does system handle percent-encoding in transformations? (Provide both raw and encoded versions of variables; preserve encoding by default)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide UI for users to create custom URL transformation rules associated with any domain/path pattern
- **FR-002**: System MUST support template-based transformations using placeholder syntax for URL components (scheme, host, port, path, query parameters, fragment)
- **FR-003**: System MUST support JavaScript-based transformations for complex URL manipulation logic
- **FR-004**: System MUST execute JavaScript transformations in a sandboxed environment with no access to file system, network, or system APIs
- **FR-005**: System MUST enforce execution timeout of 1 second for JavaScript transformations to prevent hanging
- **FR-006**: System MUST provide real-time preview/testing interface where users can input sample URLs and see transformation output before saving
- **FR-007**: System MUST validate transformation output to ensure it produces syntactically valid URLs
- **FR-008**: System MUST detect and prevent circular transformations that would cause infinite loops
- **FR-009**: System MUST catch and handle JavaScript execution errors gracefully, falling back to opening original URL unchanged
- **FR-010**: System MUST allow users to enable/disable custom transformations without deleting them
- **FR-011**: System MUST persist custom transformations with CloudKit sync support across devices
- **FR-012**: System MUST validate JavaScript syntax before saving transformation rule
- **FR-013**: System MUST provide helpful error messages indicating specific validation failures in transformations
- **FR-014**: Users MUST be able to export custom transformations as JSON for sharing
- **FR-015**: Users MUST be able to import transformation configurations from JSON files or clipboard
- **FR-016**: System MUST preserve existing rule evaluation behavior while adding custom transformation capability
- **FR-017**: Custom transformations MUST integrate with existing rule priority and pattern matching system
- **FR-018**: System MUST provide template variable reference and JavaScript API documentation within the UI
- **FR-019**: System MUST log transformation failures with sufficient detail for debugging
- **FR-020**: System MUST allow transformations on any URL scheme, not just HTTP(S)

### Key Entities

- **CustomTransformation**: Represents a user-defined URL transformation rule
  - Unique identifier
  - Associated domain/path patterns (reuses existing Rule pattern matching)
  - Transformation type (template or JavaScript)
  - Transformation logic (template string or JavaScript code)
  - Target browser or app (optional - can transform and still route normally)
  - Enabled/disabled state
  - CloudKit sync metadata (recordID, lastModified, deviceID)
  - Creation and last modified timestamps

- **TransformationResult**: Output of applying a transformation
  - Original URL
  - Transformed URL
  - Success/failure status
  - Error message (if failed)
  - Execution time (for performance monitoring)
  - Whether fallback was used

- **TemplateContext**: Variables available in template transformations
  - scheme: URL scheme (http, https, etc.)
  - host: hostname/domain
  - port: port number (if present)
  - path: full path including leading slash
  - query: query string including leading ?
  - fragment: fragment/hash including leading #
  - Individual query parameter access via {query:paramName}
  - Path component access via {pathComponent:N}

- **JavaScriptContext**: Sandboxed JavaScript execution environment
  - Input object with URL properties (scheme, host, path, query, etc.)
  - Return value must be a string (transformed URL)
  - Execution timeout limit (1 second)
  - Restricted API surface (no I/O, no network, only string manipulation)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can create a basic template transformation and successfully apply it to a URL in under 3 minutes
- **SC-002**: JavaScript transformations execute within 100ms for 95% of cases (excluding intentional timeout scenarios)
- **SC-003**: Custom transformations have system failure rate below 1% (excluding user logic errors)
- **SC-004**: Users can test and preview transformations with immediate feedback (under 100ms response time)
- **SC-005**: System successfully detects and prevents 100% of circular transformation loops before execution
- **SC-006**: 80% of users successfully create their first custom transformation without consulting external documentation
- **SC-007**: Custom transformations sync across devices via CloudKit with 99% success rate
- **SC-008**: JavaScript sandbox successfully blocks 100% of file system, network, and privileged API access attempts
- **SC-009**: User satisfaction score of 4+ out of 5 for custom transformation feature usability
- **SC-010**: Transformation validation catches 95% of common user errors (invalid syntax, malformed templates) before saving

## Assumptions

- Users creating custom transformations understand basic URL structure concepts (scheme, host, path, query)
- JavaScript mode is an advanced feature expected to be used by ~10-20% of users
- Template mode will satisfy ~80% of custom transformation use cases
- Users will primarily share transformations through community channels (forums, Discord, documentation sites)
- Average user will create 1-5 custom transformations
- Default timeout of 1 second is sufficient for JavaScript transformations - assumes no complex algorithmic processing needed
- Sandboxing JavaScript via JavaScriptCore framework provides adequate security without need for separate process isolation
- Users understand that transformation errors will fall back to opening the original URL unchanged
- Custom transformations integrate with existing rule system without needing separate priority/evaluation scheme
- Most transformations will be applied to HTTP(S) URLs, but system should support any scheme

## Dependencies

- Existing domain pattern matching system (Rule.domains, URLProcessingEngine)
- CloudKit sync infrastructure (CloudKitSyncManager)
- Rule priority and evaluation engine (URLProcessor)
- URL parsing capabilities (URLComponents, URL class)
- JavaScript execution environment (JavaScriptCore framework)

## Scope Boundaries

### In Scope

- Template-based transformations with placeholder variables for URL components
- JavaScript-based transformations with sandboxed execution
- Real-time testing and preview interface within rule editor
- Validation of transformation syntax and output
- Import/export of transformation configurations as JSON
- CloudKit sync for custom transformations
- Error handling and fallback to original URL behavior
- Integration with existing rule evaluation and pattern matching system
- Transformations on any URL scheme (HTTP, HTTPS, custom schemes, etc.)
- Documentation/examples of template syntax and JavaScript API within UI

### Out of Scope

- Built-in marketplace or repository for sharing transformations
- Visual/drag-and-drop transformation builder (text-based editing only)
- Support for programming languages other than JavaScript
- Network-based transformations (e.g., API calls to resolve shortened URLs)
- Machine learning or AI-assisted transformation suggestions
- Mobile app support (macOS only)
- Transformation version history or rollback (beyond standard CloudKit sync)
- Collaborative editing of transformations
- A/B testing of different transformation variants
- Automatic detection of optimal transformation for a given URL
- Native app availability checking (transform URL regardless of whether target app is installed)
