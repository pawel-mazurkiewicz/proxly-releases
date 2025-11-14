# Native App URL Transformation – Rule-Level Architecture

**Version:** 2.0 (rewrite)  
**Date:** 2025-10-08  
**Status:** Draft  
**Author:** Architecture Working Group

---

## Executive Summary

Proxly will let users send qualifying web links directly to installed native applications (e.g. `https://www.figma.com/file/abc` → `figma://file/abc`). Rather than introducing a parallel rule system, the transformation capability becomes an option inside existing routing rules. A rule can opt into transforming the incoming URL before Proxly launches the configured opener. This keeps the mental model intact (one rule decides everything about how the URL opens), guarantees that rule persistence continues to own configuration, and allows us to evolve additional transformation behaviours without fragmenting the architecture.

Key architectural decisions:
- Extend `Rule` with a self-contained `NativeAppTransformation` struct describing the transformation intent and patterns.
- Introduce an `actor NativeAppTransformationEngine` that evaluates patterns in a thread-safe way and produces transformed URLs plus audit data.
- Invoke the engine from the rule execution path (after a rule matches but before `BrowserLauncher` runs) so that licence/consent checks and fallbacks remain unchanged.
- Surface configuration inside the existing rule editor UI.
- Address previously identified risks: deterministic pattern identifiers, safe template handling, explicit AppEnumeration APIs, and a single persistence source of truth (the rule model).

---

## 1. Feature Overview

### 1.1 Problem
Users prefer native desktop apps for certain services. Proxly currently only routes to browsers; even if a matching rule exists, it cannot translate the URL into a scheme the native app understands.

### 1.2 Goal
Allow a routing rule to optionally transform the matched URL into a native app URL and launch that app. Users control:
- Whether transformation happens at all for the rule.
- Which transformation pattern applies when multiple matches are possible.
- Fallback behaviour if the target native app is unavailable or the transformation fails.

### 1.3 Non-Goals (Phase 1)
- User-authored arbitrary transformation scripts.
- Community pattern sharing or remote updates.
- Analytics or telemetry for transformations.
- Browser extensions for pre-flight testing.

---

## 2. Design Principles

1. **Rule-Centric** – All configuration, persistence, and evaluation live within `Rule`.
2. **Predictable Identity** – Patterns have deterministic identifiers to support edits and diffs.
3. **Safe Transformations** – Use URL component APIs and explicit validation to avoid malformed or malicious URLs.
4. **Thread Safety** – Actor isolation protects caches, rate limits, and Combine publications.
5. **Progressive Enhancement** – Tiered pattern complexity enables incremental delivery.

---

## 3. Data Model Changes

### 3.1 Rule Extension

File: `Proxly/Models/Rule.swift`

```swift
struct Rule: Identifiable, Codable {
    // existing properties…

    var nativeAppTransformation: NativeAppTransformation?
}
```

`nativeAppTransformation == nil` means “open as-is”.

### 3.2 NativeAppTransformation Struct

```swift
struct NativeAppTransformation: Codable, Equatable {
    var isEnabled: Bool
    var targetBundleID: String       // e.g. "com.figma.Desktop"
    var nativeScheme: String         // e.g. "figma"
    var fallback: FallbackBehaviour
    var patterns: [TransformationPattern]
    var lastUpdated: Date

    enum FallbackBehaviour: String, Codable {
        case useRuleBrowser          // Use the rule’s browser target
        case useDefaultBrowser       // Fall back to system default
        case promptUser              // Show selection panel
    }
}
```

### 3.3 TransformationPattern

```swift
struct TransformationPattern: Codable, Equatable, Identifiable {
    let id: TransformationPatternID
    let matchers: [PatternMatcher]         // host/path conditions
    let transformation: TransformationLogic
    let notes: String?
    let examples: [TransformationExample]

    struct TransformationExample: Codable, Equatable {
        let input: URL
        let output: URL
        let description: String
    }
}
```

Deterministic identifier:

```swift
struct TransformationPatternID: RawRepresentable, Codable, Hashable {
    let rawValue: String

    init(rawValue: String) { self.rawValue = rawValue }

    static func derived(appBundleID: String, slug: String) -> Self {
        let normalised = slug.lowercased().replacingOccurrences(of: "\\s+", with: "-", options: .regularExpression)
        return TransformationPatternID(rawValue: "\(appBundleID)#\(normalised)")
    }
}
```

### 3.4 Pattern Matchers

```swift
enum PatternMatcher: Codable, Equatable {
    case host(HostPattern)
    case path(PathPattern)
    case query(key: String, value: String?)
    case customRegex(NSRegularExpressionCodable)
}
```

`HostPattern` reuses existing wildcard/regex helpers in `URLProcessingEngine`.

### 3.5 TransformationLogic

```swift
enum TransformationLogic: Codable, Equatable {
    case schemeReplacement(preserveQuery: Bool)
    case schemeAndHost(rewriteHost: Bool, preserveQuery: Bool)
    case parameterMap(ParameterMap)
    case template(Template)
}
```

`ParameterMap` and `Template` definitions include percent-encoding instructions (see Section 5).

---

## 4. NativeAppTransformationEngine

### 4.1 Overview
- File: `Proxly/Services/NativeAppTransformationEngine.swift`
- Type: `actor` to model concurrent access from routing and UI.

Responsibilities:
1. Evaluate a URL against a rule’s patterns.
2. Produce a transformed URL or a failure reason.
3. Cache pattern matches per `(bundleID, matcherSignature)` to avoid recomputation.
4. Validate that the target scheme is registered and safe.

### 4.2 Public API

```swift
actor NativeAppTransformationEngine {
    static let shared = NativeAppTransformationEngine(
        appEnumeration: AppEnumerationServicing = AppEnumerationService.shared,
        clock: Clock = .shared
    )

    func transform(
        url: URL,
        with config: NativeAppTransformation,
        ruleContext: RuleContext
    ) async -> TransformationOutcome

    func canTransform(_ config: NativeAppTransformation) async -> Bool
    func refreshCaches()
}
```

`RuleContext` bundles `ruleID`, `ruleName`, analytics metadata, and original URL for logging.

### 4.3 Outcome Enum

```swift
enum TransformationOutcome {
    case transformed(url: URL, patternID: TransformationPatternID)
    case skipped(reason: SkipReason)

    enum SkipReason {
        case disabled
        case appMissing(bundleID: String)
        case noPatternMatched
        case patternFailed(patternID: TransformationPatternID, error: TransformationError)
    }
}
```

### 4.4 Rate Limiting
- Keep a `(ruleID, minuteBucket)` counter capped at 300/minute.
- Enforced inside the actor for consistency; if exceeded we surface `.patternFailed(..., .rateLimited)`.
- Rate limits are defensive; we log and continue with original URL.

---

## 5. Transformation Execution

### 5.1 Matching
- Combine matchers using logical AND.
- Cache key = `bundleID + host + pathPrefix + patternID`.
- Avoid the previous host-only cache collision issue.

### 5.2 Logic Implementations

#### Scheme Replacement
Uses `URLComponents` to set the scheme, optionally clearing queries.

#### Scheme & Host
Copies path, optionally rewrites host from a template (e.g. Slack workspace).

#### Parameter Map
`ParameterMap` supplies:

```swift
struct ParameterMap: Codable, Equatable {
    var preserveUnmapped: Bool
    var mappings: [String: String]    // source -> destination
    var dropKeys: [String]
}
```

The engine builds `URLComponents`, iterates query items, applies mappings, percent-encodes via `URLQueryItem`.

#### Template
`Template` stores a format string with named placeholders and a schema for escaping:

```swift
struct Template: Codable, Equatable {
    var format: String                // e.g. "{scheme}://open?url={url:percentEncoded}"
    var allowedVariables: [TemplateVariable]
}

enum TemplateVariable: String, Codable {
    case scheme, host, hostPrefix, path, query, fragment, fullURL
    case pathComponent(Int)
}
```

`hostPrefix` resolves to the leading label(s) before a known suffix (e.g. `workspace` for `workspace.slack.com`) when the matcher specifies a wildcard host.

The engine validates that the format only references entries in `allowedVariables` and automatically encodes `{fullURL:percentEncoded}` (and similar `:percentEncoded` suffix) variants. Templates produce `URLComponents` rather than plain string concatenation.

### 5.3 Validation
- Confirm result scheme equals `nativeScheme`.
- Ask `AppEnumerationServicing` for `isSchemeRegistered(_:)` and `appSupportsScheme(bundleID:scheme:)`.
- Reject forbidden prefixes (`javascript:`, `data:`, `file:`, `about:`).
- Enforce max length 4 KB.
- Verify components produce a valid, absolute `URL`.

### 5.4 Error Types

```swift
enum TransformationError: LocalizedError {
    case invalidInput
    case noHost
    case parameterMismatch
    case templateVariableForbidden(String)
    case templateOutputInvalid
    case rateLimited
    case schemeNotRegistered(String)
}
```

All surfaced through `SkipReason.patternFailed`.

---

## 6. Integration with Rule Processing

### 6.1 Evaluation Flow

File: `Proxly/Services/URLProcessor.swift`

1. Existing licence, consent, and rule enablement checks run first (no change).
2. `URLProcessingEngine` returns the rule as today.
3. Before launching the rule’s chosen browser target:
    - If `rule.nativeAppTransformation?.isEnabled == true`, call the engine.
    - Await outcome.
4. `TransformationOutcome` handling:
    - `.transformed(url: nativeURL, patternID:)`
        - Use `BrowserLauncher` path that opens specific bundle determined by `nativeAppTransformation.targetBundleID`.
        - Log via `AppLogger`: include rule ID/name, pattern ID, input/output.
        - Skip the browser target entirely.
    - `.skipped(reason:)`
        - Respect configured fallback:
            - `.useRuleBrowser` (default): continue with original rule path (browser/profiles etc.).
            - `.useDefaultBrowser`: call `openURLWithDefaultBrowser`.
            - `.promptUser`: show selection panel.
        - Include reason in logs/notifications.

### 6.2 Interaction with Opener Constraints
Existing opener constraints stay in effect; a rule can still use opener matching, and transformation only kicks in after the rule is chosen.

### 6.3 Telemetry Hooks
Add log levels but postpone analytics events until Phase 2.

---

## 7. AppEnumerationService Additions

Interface protocol (`AppEnumerationServicing`) gains:

```swift
protocol AppEnumerationServicing {
    func isSchemeRegistered(_ scheme: String) -> Bool
    func appSupportsScheme(bundleID: String, scheme: String) -> Bool
    func applicationInfo(for bundleID: String) -> AppInfo?
}
```

Existing service already scans apps; augment persistence with scheme lookup tables keyed by bundle identifier.

---

## 8. User Experience

### 8.1 Rule Editor

File: `Proxly/Views/Rules/RuleDetailView.swift`

Add a “Native App” section:
- Toggle: “Open in native app when available”.
- Picker for detected native apps (populated from `AppEnumerationService`).
- Summary of selected patterns with edit button.
- Inline tester (open a sheet to paste URL and preview transformation).

SwiftUI bindings call dedicated async methods on the engine/service rather than mutating `@Published private(set)` properties.

### 8.2 Pattern Library

Provide presets stored in `docs/patterns/native-app-presets.json`. The rule editor can import a preset into the rule’s config (copy, not reference) to keep persistence single-sourced.

### 8.3 Error Surfacing
- If transformation skipped due to missing app, show notification with “Open in browser instead” CTA (honours fallback).
- Provide inline validation errors when saving a rule (e.g. invalid template variable).

---

## 9. Persistence & Migration

- Rules already persist to `Rule` storage; no new top-level keys.
- Migration Step: when loading rules, default `nativeAppTransformation` to `nil`. For built-in presets shipped with the app, we append them to the initial rule templates during onboarding but never inject silently into existing user rules.
- Deterministic IDs guarantee diff-friendly syncing via iCloud (future).

---

## 10. Security Considerations

1. **Scheme Whitelist** – Only schemes registered by installed apps are permitted.
2. **Forbidden Schemes** – Reject known dangerous schemes and path traversal patterns.
3. **Template Sandboxing** – Templates only expose whitelisted tokens; no arbitrary evaluation.
4. **Size Limits** – Enforce length < 4 KB to avoid DoS.
5. **Audit Trail** – Log every transformation (rule, pattern, host) for forensics; redact query parameters flagged as sensitive by the pattern config.
6. **User Consent** – No network calls; transformations run locally.

---

## 11. Testing Strategy

### 11.1 Unit Tests (new)
File: `ProxlyTests/NativeAppTransformationEngineTests.swift`
- Pattern matching per tier.
- Rate limit behaviour.
- Validation edge cases (missing host, forbidden scheme, template misuse).
- Cache correctness: identical host/path with different pattern IDs should not collide.

### 11.2 Integration Tests
File: `ProxlyTests/RuleExecutionTransformationTests.swift`
- Feed URLs through `URLProcessingEngine` + `URLProcessor` to ensure rules launch native apps when configured.
- Test fallback modes.

### 11.3 UI Tests (optional)
XCTest UI harness to confirm rule editor toggle and tester behave.

---

## 12. Phased Delivery

### Phase 1 – MVP (2–3 weeks)
- Engine implementation with scheme replacement & scheme+host tiers.
- Presets: Figma, Notion, Spotify, Linear, Discord, Trello.
- Rule UI toggle, picker, basic tester.
- Unit + integration tests.

### Phase 2 – Enhanced (4–6 weeks)
- Parameter mapping & template tiers.
- Additional presets: Zoom, Slack, VS Code, Obsidian.
- Tester improvements (show applied pattern ID, copy button).
- Notification UX for missing apps.

### Phase 3 – Advanced (future)
- User-authored presets library.
- Sync / sharing of transformation configs.
- Usage analytics (opt-in).

---

## 13. Open Questions

1. Should rule evaluation short-circuit if transformation succeeds, or should we still record a “rule executed” analytics event downstream? (Likely yes; needs telemetry design.)
2. Do we allow multiple native apps per rule (ordered list) or exactly one? This spec assumes one; revisit if user demand emerges.
3. How do we surface transformation success/failure in the UI history view? (Potential future enhancement.)

---

## 14. Appendix: Example Patterns

### Figma (preset)
```json
{
  "id": "com.figma.Desktop#file",
  "matchers": [
    { "host": { "wildcard": "*.figma.com" } },
    { "path": { "regex": "^/(file|proto)/" } }
  ],
  "transformation": { "schemeAndHost": { "rewriteHost": false, "preserveQuery": true } },
  "examples": [
    {
      "input": "https://www.figma.com/file/abc123/Design",
      "output": "figma://www.figma.com/file/abc123/Design",
      "description": "Design file"
    }
  ]
}
```

### Zoom
```json
{
  "id": "us.zoom.xos#meeting",
  "matchers": [
    { "host": { "exact": "zoom.us" } },
    { "path": { "regex": "^/j/" } }
  ],
  "transformation": {
    "parameterMap": {
      "preserveUnmapped": false,
      "mappings": { "pwd": "pwd", "confno": "j" },
      "dropKeys": []
    }
  },
  "examples": [
    {
      "input": "https://zoom.us/j/123456789?pwd=abc",
      "output": "zoommtg://zoom.us/join?confno=123456789&pwd=abc",
      "description": "Join meeting"
    }
  ]
}
```

### Notion
```json
{
  "id": "notion.id#workspace",
  "matchers": [
    { "host": { "wildcard": "*.notion.so" } }
  ],
  "transformation": {
    "schemeAndHost": { "rewriteHost": false, "preserveQuery": true }
  },
  "examples": [
    {
      "input": "https://acme.notion.so/page-abc123",
      "output": "notion://acme.notion.so/page-abc123",
      "description": "Workspace page"
    }
  ]
}
```

### Spotify
```json
{
  "id": "com.spotify.client#media",
  "matchers": [
    { "host": { "exact": "open.spotify.com" } },
    { "path": { "regex": "^/(track|album|playlist|episode)/" } }
  ],
  "transformation": {
    "schemeReplacement": { "preserveQuery": true }
  },
  "examples": [
    {
      "input": "https://open.spotify.com/track/123?si=abc",
      "output": "spotify://track/123?si=abc",
      "description": "Track"
    }
  ]
}
```

### Linear
```json
{
  "id": "com.linearapp.desktop#issue",
  "matchers": [
    { "host": { "exact": "linear.app" } },
    { "path": { "regex": "^/[^/]+/issue/[A-Z]+-[0-9]+$" } }
  ],
  "transformation": {
    "schemeReplacement": { "preserveQuery": false }
  },
  "examples": [
    {
      "input": "https://linear.app/team/issue/ABC-123",
      "output": "linear://linear.app/team/issue/ABC-123",
      "description": "Issue deep link"
    }
  ]
}
```

### Discord
```json
{
  "id": "com.hnc.Discord#channels",
  "matchers": [
    { "host": { "exact": "discord.com" } },
    { "path": { "regex": "^/channels/[0-9]+/[0-9]+$" } }
  ],
  "transformation": {
    "template": {
      "format": "discord://-/channels/{pathComponent(1)}/{pathComponent(2)}",
      "allowedVariables": ["pathComponent(1)", "pathComponent(2)"]
    }
  },
  "examples": [
    {
      "input": "https://discord.com/channels/123/456",
      "output": "discord://-/channels/123/456",
      "description": "Server channel"
    }
  ]
}
```

### Trello
```json
{
  "id": "com.atlassian.trello#card",
  "matchers": [
    { "host": { "exact": "trello.com" } }
  ],
  "transformation": {
    "schemeAndHost": {
      "rewriteHost": false,
      "preserveQuery": true
    }
  },
  "examples": [
    {
      "input": "https://trello.com/c/abc123/card-name",
      "output": "trello://trello.com/c/abc123/card-name",
      "description": "Card"
    }
  ]
}
```

### Slack
```json
{
  "id": "com.tinyspeck.slackmacgap#channel",
  "matchers": [
    { "host": { "wildcard": "*.slack.com" } },
    { "path": { "regex": "^/archives/[A-Z0-9]+$" } }
  ],
  "transformation": {
    "template": {
      "format": "slack://channel?team={hostPrefix}&id={pathComponent(1)}",
      "allowedVariables": ["hostPrefix", "pathComponent(1)"]
    }
  },
  "examples": [
    {
      "input": "https://workspace.slack.com/archives/C123456",
      "output": "slack://channel?team=workspace&id=C123456",
      "description": "Channel"
    }
  ]
}
```

> `hostPrefix` resolves to the host without the `.slack.com` suffix.

### VS Code
```json
{
  "id": "com.microsoft.VSCode#github",
  "matchers": [
    { "host": { "exact": "github.com" } },
    { "path": { "regex": "^/[^/]+/[^/]+/(blob|tree)/.*" } }
  ],
  "transformation": {
    "template": {
      "format": "vscode://vscode.github/open?url={fullURL:percentEncoded}",
      "allowedVariables": ["fullURL"]
    }
  },
  "examples": [
    {
      "input": "https://github.com/user/repo/blob/main/file.swift",
      "output": "vscode://vscode.github/open?url=https%3A%2F%2Fgithub.com%2Fuser%2Frepo%2Fblob%2Fmain%2Ffile.swift",
      "description": "GitHub file"
    }
  ]
}
```

### Obsidian
```json
{
  "id": "md.obsidian#vault",
  "matchers": [
    { "host": { "exact": "obsidian.md" } },
    { "path": { "regex": "^/vault/.+" } }
  ],
  "transformation": {
    "template": {
      "format": "obsidian://{path}",
      "allowedVariables": ["path"]
    }
  },
  "examples": [
    {
      "input": "https://obsidian.md/vault/MyVault/Note",
      "output": "obsidian://vault/MyVault/Note",
      "description": "Vault note"
    }
  ]
}
```

---  
**End of Specification**
- AppEnumerationService additions now include `applicationInfo(for:)` to surface bundle metadata for the rule editor.
