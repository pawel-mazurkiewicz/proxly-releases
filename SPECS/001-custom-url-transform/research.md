# Technical Research: Custom URL Transformations

**Feature**: 001-custom-url-transform
**Date**: 2025-10-31
**Status**: Complete

## Overview

This document consolidates research findings for implementing custom URL transformation capabilities in Proxly. The feature requires template parsing, JavaScript sandboxing, URL component extraction, and CloudKit sync integration.

---

## 1. JavaScriptCore Sandboxing

### Decision
Use JavaScriptCore's built-in isolation without additional explicit sandboxing. Do NOT expose any native Swift functions or objects to the JSContext beyond what's strictly necessary (URL component inputs only).

### Rationale
- JavaScriptCore in iOS/macOS runs in-process but with memory isolation
- By default, JavaScript has no access to file system, network, or system APIs
- The sandbox is implicit - JavaScript can only access what you explicitly expose via `setObject(_:forKeyedSubscript:)`
- This approach is sufficient for our use case since we're only passing URL strings and components

### Implementation Notes
```swift
let context = JSContext()!

// Create isolated context - DO NOT add any native objects
context.exceptionHandler = { context, exception in
    // Log exception for debugging
    print("JS Error: \(exception?.toString() ?? "unknown")")
}

// Only expose input data (no native APIs)
context.setObject([
    "url": urlString,
    "scheme": components.scheme ?? "",
    "host": components.host ?? "",
    "path": components.path,
    "query": components.query ?? "",
    "fragment": components.fragment ?? ""
], forKeyedSubscript: "input" as NSString)

// Execute user code
let result = context.evaluateScript(userJavaScript)
```

**Security boundaries:**
- No `fetch`, `XMLHttpRequest`, or network APIs (not available in JSContext)
- No `require`, `import`, or module loading (not available)
- No file system access (not available)
- No setTimeout/setInterval by default (we won't add these polyfills)
- User code can ONLY manipulate strings and perform calculations

### Alternatives Considered
- **WKWebView with JavaScript**: Too heavyweight, requires WebKit which has more attack surface
- **Separate sandboxed process**: Unnecessary complexity, JSContext isolation is sufficient
- **Custom parser/DSL**: Less flexible, wouldn't support complex logic users might need

---

## 2. JavaScript Execution Timeout

### Decision
Use `DispatchWorkItem` with cancellation to implement 1-second timeout. Kill execution by invalidating the JSContext if timeout occurs.

### Rationale
- JavaScriptCore doesn't have built-in timeout mechanism
- DispatchWorkItem allows clean cancellation without thread termination
- Invalidating JSContext immediately stops execution (context becomes unusable but that's acceptable)
- 1 second is generous for string manipulation but prevents infinite loops

### Implementation Notes
```swift
func executeJavaScriptWithTimeout(
    _ script: String,
    input: [String: Any],
    timeout: TimeInterval = 1.0
) -> Result<String, TransformationError> {
    let context = JSContext()!
    var result: Result<String, TransformationError>?

    let workItem = DispatchWorkItem {
        context.setObject(input, forKeyedSubscript: "input" as NSString)

        context.exceptionHandler = { _, exception in
            result = .failure(.javascriptError(exception?.toString() ?? "Unknown error"))
        }

        if let jsResult = context.evaluateScript(script),
           let urlString = jsResult.toString() {
            result = .success(urlString)
        } else {
            result = .failure(.javascriptError("Script did not return a string"))
        }
    }

    DispatchQueue.global(qos: .userInitiated).async(execute: workItem)

    // Wait for completion or timeout
    let timeoutResult = workItem.wait(timeout: .now() + timeout)

    if timeoutResult == .timedOut {
        workItem.cancel()
        return .failure(.timeout)
    }

    return result ?? .failure(.javascriptError("No result"))
}
```

### Alternatives Considered
- **pthread_cancel**: Too low-level, dangerous, can corrupt memory
- **NSOperation cancellation**: Doesn't actually stop JavaScript execution
- **Separate process with watchdog**: Overkill for simple timeout
- **JavaScriptCore private APIs**: Undocumented, might break in future macOS versions

---

## 3. Template Variable Parsing

### Decision
Use regex-based substitution with NSRegularExpression for template variable replacement. Cache compiled regex patterns.

### Rationale
- Template syntax is simple: `{variableName}` or `{query:paramName}` or `{pathComponent:0}`
- Regex handles this efficiently without building a full parser
- NSRegularExpression is fast and battle-tested
- Caching compiled patterns prevents repeated compilation overhead

### Implementation Notes
```swift
struct TemplateEngine {
    private static let cache = NSCache<NSString, NSRegularExpression>()

    // Regex pattern: {variableName} or {category:name}
    private static let pattern = #"\{([a-zA-Z][a-zA-Z0-9]*(?::[a-zA-Z0-9_-]+)?)\}"#

    static func substitute(
        template: String,
        components: URLComponents
    ) -> String {
        let regex = getCompiledRegex()
        let nsString = template as NSString
        let range = NSRange(location: 0, length: nsString.length)

        var result = template
        let matches = regex.matches(in: template, range: range).reversed()

        for match in matches {
            let matchRange = match.range(at: 1)
            let variable = nsString.substring(with: matchRange)

            if let value = resolveVariable(variable, from: components) {
                let fullMatchRange = match.range(at: 0)
                result = (result as NSString)
                    .replacingCharacters(in: fullMatchRange, with: value)
            }
        }

        return result
    }

    private static func resolveVariable(
        _ variable: String,
        from components: URLComponents
    ) -> String? {
        let parts = variable.split(separator: ":")

        switch parts[0] {
        case "scheme": return components.scheme
        case "host": return components.host
        case "port": return components.port.map(String.init)
        case "path": return components.path
        case "query":
            if parts.count == 2 {
                // {query:paramName}
                let paramName = String(parts[1])
                return components.queryItems?
                    .first(where: { $0.name == paramName })?
                    .value
            } else {
                // {query}
                return components.query.map { "?\($0)" }
            }
        case "fragment":
            return components.fragment.map { "#\($0)" }
        case "pathComponent":
            if parts.count == 2, let index = Int(parts[1]) {
                let pathComponents = components.path
                    .split(separator: "/")
                    .map(String.init)
                return pathComponents.indices.contains(index)
                    ? pathComponents[index]
                    : nil
            }
            return nil
        default: return nil
        }
    }

    private static func getCompiledRegex() -> NSRegularExpression {
        let key = pattern as NSString
        if let cached = cache.object(forKey: key) {
            return cached
        }

        let regex = try! NSRegularExpression(pattern: pattern)
        cache.setObject(regex, forKey: key)
        return regex
    }
}
```

**Performance**: Template substitution should complete in <10ms for typical templates (5-10 variables).

### Alternatives Considered
- **String.replacingOccurrences**: Not flexible enough for conditional logic (query:paramName)
- **Mustache/Handlebars library**: Overkill, adds dependency, slower
- **Custom lexer/parser**: More robust but unnecessary complexity for simple syntax
- **String interpolation**: Not applicable to user-provided templates

---

## 4. Template Pattern Caching

### Decision
Cache compiled regex pattern but NOT template strings themselves. Each template evaluation is stateless and fast enough without caching results.

### Rationale
- The regex pattern is constant (`\{([a-zA-Z]...)\}`) so caching it makes sense
- Template strings vary per transformation rule, but evaluating them is fast (<10ms)
- Caching template evaluation results would be complex (need cache key from URL components)
- Memory overhead not justified for ~1-5 transformations per user
- Keep implementation simple and stateless

### Implementation Notes
```swift
// ONLY cache the regex pattern (shown above)
private static let cache = NSCache<NSString, NSRegularExpression>()

// Do NOT cache template evaluation results
// Each call to substitute() is independent and fast
```

### Alternatives Considered
- **Cache template + URL component hash → result**: Too complex, cache invalidation issues
- **No caching at all**: Regex compilation is relatively expensive, worth caching
- **Pre-compile templates at save time**: Not applicable since input (URL) varies

---

## 5. URL Component Extraction

### Decision
Use `URLComponents` for all URL parsing. It handles percent-encoding, query parameters, and edge cases correctly.

### Rationale
- `URLComponents` is the standard Swift API for URL parsing
- Automatically handles percent-encoding/decoding
- Provides structured access to scheme, host, path, query items, fragment
- Well-tested by Apple and handles RFC 3986 compliance
- No need for manual string manipulation or regex parsing

### Implementation Notes
```swift
func extractComponents(from url: URL) -> URLComponents? {
    guard let components = URLComponents(url: url, resolvingAgainstBaseURL: true) else {
        return nil
    }

    return components
}

// Access components:
let scheme = components.scheme ?? ""
let host = components.host ?? ""
let path = components.path  // includes leading /
let query = components.query  // excludes ?
let fragment = components.fragment  // excludes #

// Query parameters
let queryItems = components.queryItems ?? []
for item in queryItems {
    print("\(item.name) = \(item.value ?? "")")
}

// Individual query param
let tokenValue = components.queryItems?
    .first(where: { $0.name == "token" })?
    .value
```

**Percent-encoding handling**: URLComponents automatically decodes percent-encoded values in `queryItems`, but preserves encoding in raw `query` string. We expose both to JavaScript/templates.

### Alternatives Considered
- **URL class only**: Less structured, harder to access query parameters
- **Manual string parsing**: Error-prone, doesn't handle edge cases (IPv6, IDN, etc.)
- **Third-party URL parsing library**: Unnecessary dependency

---

## 6. Circular Transformation Detection

### Decision
Compare transformed URL with original URL after transformation. If identical (and transformation was expected to change it), warn user. Do NOT prevent saving - let user decide.

### Rationale
- True circular transformation (A → B → A) requires tracking transformation chain, which is complex
- More likely scenario: user creates transformation that doesn't change URL
- Simple string comparison catches this common mistake
- Allow advanced users to intentionally create identity transformations (for logging, etc.)
- Infinite loops prevented by system architecture (transformation runs once per URL processing)

### Implementation Notes
```swift
func validateTransformation(
    _ transformation: CustomTransformation,
    testURL: URL
) -> ValidationResult {
    let original = testURL.absoluteString

    guard let transformed = applyTransformation(transformation, to: testURL) else {
        return .error("Transformation failed")
    }

    let result = transformed.absoluteString

    if original == result {
        return .warning(
            "Transformation produces unchanged URL. " +
            "If this is intentional, you can ignore this warning."
        )
    }

    // Additional check: if transformation should modify specific component
    // but didn't, warn user
    if transformation.type == .template {
        let hasVariables = transformation.logic.contains("{")
        if hasVariables && original == result {
            return .warning("Template contains variables but URL unchanged")
        }
    }

    return .success(result)
}
```

**Why not prevent infinite loops in chain?**
- Transformations apply once per URL processing flow
- URLProcessor doesn't re-process transformed URLs
- If user creates rule that matches transformed URL, that's a separate rule evaluation
- System architecture prevents infinite loops by design

### Alternatives Considered
- **Track transformation chain**: Complex, not worth it for single-transformation-per-rule model
- **Prevent saving identity transformations**: Too restrictive, breaks valid use cases (logging, testing)
- **Hash-based loop detection**: Overcomplicated for architecture that processes each URL once

---

## 7. CloudKit Sync for Transformations

### Decision
Store transformations as part of Rule model with new optional fields. JavaScript code stored as plain String in CloudKit. Reuse existing CloudKitSyncManager infrastructure.

### Rationale
- Rule already has CloudKit sync via CKRecord
- Adding optional transformation fields maintains backward compatibility
- CloudKit automatically handles string encoding (UTF-8)
- JavaScript code is typically small (<1KB), well within CloudKit field limits

### Implementation Notes
```swift
// Extend Rule model
struct Rule: Codable {
    // ... existing fields ...

    // New optional fields (nil for non-transformation rules)
    var transformation: TransformationType?
    var transformationLogic: String?

    enum TransformationType: String, Codable {
        case template
        case javascript
    }
}

// CloudKit sync (in CloudKitSyncManager)
func ruleToRecord(_ rule: Rule) -> CKRecord {
    let record = CKRecord(recordType: "Rule", recordID: ...)

    // ... existing fields ...

    // Add transformation fields if present
    if let transformation = rule.transformation {
        record["transformation"] = transformation.rawValue
    }
    if let logic = rule.transformationLogic {
        record["transformationLogic"] = logic
    }

    return record
}
```

**CloudKit limits**:
- String field limit: 1MB per field (plenty for JavaScript code)
- Asset recommended for >1KB, but JavaScript will typically be <500 bytes
- Use String field for simplicity unless user reports >1KB scripts

**Encoding considerations**:
- CloudKit handles UTF-8 automatically
- No special escaping needed for JavaScript (quotes, newlines, etc.)
- Template strings with curly braces `{}` stored as-is

### Alternatives Considered
- **Separate CKRecord type for transformations**: Unnecessary complexity, complicates sync
- **Store JavaScript as CKAsset**: Overkill for small strings, complicates CRUD operations
- **Base64 encode JavaScript**: Unnecessary, CloudKit handles strings correctly
- **Separate UserDefaults key**: Would require separate CloudKit record type

---

## 8. Performance Considerations

### Expected Performance Targets

| Operation | Target | Notes |
|-----------|--------|-------|
| Template substitution | <10ms p95 | Simple regex + string replacement |
| JavaScript execution | <100ms p95 | Most scripts will be <50ms |
| JavaScript timeout | 1000ms hard limit | Prevents infinite loops |
| URL validation | <5ms | URLComponents parsing |
| Total transformation overhead | <100ms p95 | Added to existing URL processing |

### Optimization Strategies

1. **Regex pattern caching**: Prevents recompiling same pattern (saves ~1-2ms per transformation)
2. **URLComponents reuse**: Parse URL once, pass components to both template and JS engines
3. **Lazy JavaScript context creation**: Only create JSContext when needed (not for template transformations)
4. **Main thread execution**: Transformations are fast enough, avoid thread overhead
5. **Early validation**: Catch errors during rule creation, not during URL processing

### Memory Considerations

- **NSCache for regex patterns**: Limited to 10 entries (one pattern currently)
- **JSContext per execution**: Created and destroyed each time (no context pooling needed)
- **Template strings**: Stored in Rule model (~100 bytes average)
- **JavaScript code**: Stored in Rule model (~500 bytes average)
- **Total memory overhead**: ~5KB per transformation rule

---

## 9. Error Handling Strategy

### Categories of Errors

1. **Validation errors** (at rule creation):
   - Invalid template syntax
   - JavaScript syntax errors
   - Empty transformation logic

2. **Runtime errors** (during URL processing):
   - JavaScript execution errors
   - JavaScript timeout
   - Invalid URL output
   - Circular transformation (warning only)

3. **Fallback behavior**:
   - All runtime errors → open original URL unchanged
   - Log error for debugging
   - Show user notification (optional, configurable)

### Implementation Pattern

```swift
enum TransformationError: LocalizedError {
    case invalidTemplate(String)
    case javascriptSyntaxError(String, line: Int)
    case javascriptRuntimeError(String)
    case timeout
    case invalidURLOutput(String)
    case circularTransformation

    var errorDescription: String? {
        switch self {
        case .invalidTemplate(let msg):
            return "Invalid template: \(msg)"
        case .javascriptSyntaxError(let msg, let line):
            return "JavaScript syntax error at line \(line): \(msg)"
        case .javascriptRuntimeError(let msg):
            return "JavaScript error: \(msg)"
        case .timeout:
            return "Transformation timed out after 1 second"
        case .invalidURLOutput(let output):
            return "Transformation produced invalid URL: \(output)"
        case .circularTransformation:
            return "Transformation produces unchanged URL"
        }
    }
}
```

---

## 10. Testing Strategy

### Unit Tests Required

1. **Template engine**:
   - Variable substitution for all types (scheme, host, path, query, fragment)
   - Query parameter extraction (`{query:token}`)
   - Path component extraction (`{pathComponent:0}`)
   - Missing variables (should substitute with empty string)
   - Malformed templates (unclosed braces, invalid syntax)

2. **JavaScript executor**:
   - Simple transformations (return modified string)
   - Complex transformations (conditional logic, string manipulation)
   - Error handling (syntax errors, runtime errors, exceptions)
   - Timeout enforcement (infinite loop detection)
   - Input object structure (verify all URL components available)

3. **Circular transformation detection**:
   - Identity transformation (URL unchanged)
   - Template with variables but no change
   - Valid transformations that intentionally preserve URL

4. **Integration tests**:
   - End-to-end transformation in URLProcessor flow
   - CloudKit sync of transformation rules
   - UI validation feedback
   - Fallback behavior on errors

### Test Data

```swift
// Template tests
let testCases = [
    (template: "{scheme}://example.com{path}",
     url: "https://twitter.com/user/status/123",
     expected: "https://example.com/user/status/123"),

    (template: "{scheme}://{host}{path}?clean=true",
     url: "https://site.com/page?utm_source=twitter",
     expected: "https://site.com/page?clean=true"),
]

// JavaScript tests
let jsTests = [
    (script: "return input.url.replace('twitter.com', 'nitter.net');",
     url: "https://twitter.com/user",
     expected: "https://nitter.net/user"),

    (script: "if (input.host.includes('reddit')) return input.url.replace('m.', 'old.'); return input.url;",
     url: "https://m.reddit.com/r/swift",
     expected: "https://old.reddit.com/r/swift"),
]
```

---

## Summary of Decisions

| Topic | Decision | Rationale |
|-------|----------|-----------|
| JS Sandboxing | Use JSContext isolation only | Built-in security sufficient, no explicit sandbox needed |
| JS Timeout | DispatchWorkItem with 1s limit | Clean cancellation without thread termination |
| Template Parsing | Regex-based substitution | Fast, simple, handles our syntax well |
| Template Caching | Cache regex pattern only | Template evaluation is fast enough without result caching |
| URL Parsing | URLComponents | Standard API, handles edge cases correctly |
| Circular Detection | String comparison with warning | Simple, non-blocking, covers common mistakes |
| CloudKit Sync | Store in Rule model as strings | Reuse existing infrastructure, simple encoding |

All decisions prioritize **simplicity**, **performance**, and **security** while maintaining compatibility with existing Proxly architecture.
