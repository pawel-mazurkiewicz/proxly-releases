# Data Model: Custom URL Transformations

**Feature**: 001-custom-url-transform
**Date**: 2025-10-31
**Status**: Complete

## Overview

This document defines the data structures needed for custom URL transformations. The design extends the existing `Rule` model with optional transformation fields to maintain backward compatibility and leverage existing CloudKit sync infrastructure.

---

## Entity Definitions

### 1. CustomTransformation (Embedded in Rule)

Represents transformation logic embedded within a Rule. Not a separate entity but additional optional fields on the existing `Rule` model.

**Location in Model**: `Proxly/Models/Rule.swift` (extend existing)

```swift
struct Rule: Codable, Identifiable {
    // === EXISTING FIELDS (not shown for brevity) ===
    let id: UUID
    var name: String
    var type: RuleType
    var domains: [String]?
    var condition: String?
    var browserTarget: BrowserTarget
    var priority: Int
    var isEnabled: Bool
    var openerBundleId: String?
    // CloudKit sync fields
    var recordID: String?
    var lastModified: Date
    var deviceID: String?

    // === NEW FIELDS FOR TRANSFORMATIONS ===

    /// Type of transformation to apply (nil if no transformation)
    var transformationType: TransformationType?

    /// The transformation logic (template string or JavaScript code)
    var transformationLogic: String?

    /// Enum defining transformation types
    enum TransformationType: String, Codable {
        case template   // Template-based with {variables}
        case javascript // JavaScript-based with full code
    }
}
```

**Field Descriptions**:

| Field | Type | Required | Description | Validation |
|-------|------|----------|-------------|------------|
| `transformationType` | `TransformationType?` | No | Specifies whether transformation uses template or JavaScript | Must be non-nil if `transformationLogic` is present |
| `transformationLogic` | `String?` | No | The actual transformation code/template | Non-empty if `transformationType` is set; max 10KB |

**Validation Rules**:
- If `transformationType` is set, `transformationLogic` MUST be non-empty
- If `transformationLogic` is set, `transformationType` MUST be set
- `transformationLogic` length MUST be ≤ 10,000 characters
- For `template` type: Must contain valid template syntax (validated by TemplateEngine)
- For `javascript` type: Must be valid JavaScript syntax (validated by JSContext)

**Relationships**:
- Rule has 0..1 transformation (optional)
- Transformation belongs to exactly 1 Rule
- Transformation inherits Rule's domain patterns, priority, and enabled state

**State Transitions**:
```
Rule (no transformation)
  ↓ User adds transformation
Rule (with transformation)
  ↓ User edits transformation
Rule (with updated transformation)
  ↓ User removes transformation
Rule (no transformation)
```

---

### 2. TransformationResult (Transient)

Represents the outcome of applying a transformation to a URL. This is a transient object (not persisted), used for passing results between components and for testing/preview.

**Location**: `Proxly/Models/TransformationResult.swift` (new file)

```swift
struct TransformationResult {
    /// The original URL before transformation
    let originalURL: URL

    /// The transformed URL (if successful)
    let transformedURL: URL?

    /// Whether transformation succeeded
    let success: Bool

    /// Error message if transformation failed
    let errorMessage: String?

    /// Execution time in milliseconds
    let executionTimeMs: Double

    /// Whether fallback to original URL was used
    let usedFallback: Bool

    /// Creates a successful transformation result
    static func success(
        original: URL,
        transformed: URL,
        executionTime: Double
    ) -> TransformationResult {
        TransformationResult(
            originalURL: original,
            transformedURL: transformed,
            success: true,
            errorMessage: nil,
            executionTimeMs: executionTime,
            usedFallback: false
        )
    }

    /// Creates a failed transformation result with fallback
    static func failure(
        original: URL,
        error: String,
        executionTime: Double
    ) -> TransformationResult {
        TransformationResult(
            originalURL: original,
            transformedURL: nil,
            success: false,
            errorMessage: error,
            executionTimeMs: executionTime,
            usedFallback: true
        )
    }
}
```

**Field Descriptions**:

| Field | Type | Description |
|-------|------|-------------|
| `originalURL` | `URL` | The URL before transformation was applied |
| `transformedURL` | `URL?` | The resulting URL if transformation succeeded, nil if failed |
| `success` | `Bool` | True if transformation completed without errors |
| `errorMessage` | `String?` | Human-readable error description if transformation failed |
| `executionTimeMs` | `Double` | Time taken to execute transformation in milliseconds |
| `usedFallback` | `Bool` | True if error occurred and original URL should be used instead |

**Usage**:
- Returned by `TransformationEngine.apply()`
- Passed to URLProcessor for browser launching decision
- Used in UI for real-time preview feedback
- Logged for debugging transformation issues

---

### 3. URLComponentsContext (Transient)

Encapsulates parsed URL components for use in template substitution and JavaScript execution. Not persisted, created on-demand during transformation.

**Location**: `Proxly/Services/TransformationEngine.swift` (internal struct)

```swift
struct URLComponentsContext {
    let url: URL
    let components: URLComponents

    // Convenience accessors for common components
    var scheme: String { components.scheme ?? "" }
    var host: String { components.host ?? "" }
    var port: Int? { components.port }
    var path: String { components.path }
    var query: String? { components.query }
    var fragment: String? { components.fragment }
    var queryItems: [URLQueryItem] { components.queryItems ?? [] }

    // Path components (split by /)
    var pathComponents: [String] {
        path.split(separator: "/").map(String.init)
    }

    // Convert to dictionary for JavaScript input
    func toDictionary() -> [String: Any] {
        var dict: [String: Any] = [
            "url": url.absoluteString,
            "scheme": scheme,
            "host": host,
            "path": path
        ]

        if let query = query {
            dict["query"] = query
        }

        if let fragment = fragment {
            dict["fragment"] = fragment
        }

        if let port = port {
            dict["port"] = port
        }

        // Add query parameters as nested dictionary
        if !queryItems.isEmpty {
            var queryDict: [String: String] = [:]
            for item in queryItems {
                queryDict[item.name] = item.value ?? ""
            }
            dict["queryParams"] = queryDict
        }

        // Add path components as array
        dict["pathComponents"] = pathComponents

        return dict
    }

    // Get specific query parameter value
    func queryParam(_ name: String) -> String? {
        queryItems.first(where: { $0.name == name })?.value
    }

    // Get specific path component by index
    func pathComponent(at index: Int) -> String? {
        let components = pathComponents
        return components.indices.contains(index) ? components[index] : nil
    }
}
```

**Purpose**:
- Provides unified interface for accessing URL components
- Used by both TemplateEngine and JavaScriptEngine
- Ensures consistent URL parsing across transformation types
- Simplifies template variable resolution
- Converts to JavaScript-friendly dictionary for JSContext

---

## 4. TransformationEngine (Service, Stateless)

Not a data model but the stateless service that applies transformations. Mentioned here for completeness.

**Location**: `Proxly/Services/TransformationEngine.swift` (new file)

```swift
struct TransformationEngine {
    // Stateless struct - no stored properties

    /// Apply transformation to URL
    static func apply(
        transformation: Rule,
        to url: URL
    ) -> TransformationResult {
        guard let type = transformation.transformationType,
              let logic = transformation.transformationLogic else {
            // No transformation, return original
            return .success(
                original: url,
                transformed: url,
                executionTime: 0
            )
        }

        let context = URLComponentsContext(
            url: url,
            components: URLComponents(url: url, resolvingAgainstBaseURL: true)!
        )

        let startTime = Date()

        let result: Result<URL, TransformationError>

        switch type {
        case .template:
            result = TemplateEngine.substitute(logic, context: context)
        case .javascript:
            result = JavaScriptEngine.execute(logic, context: context)
        }

        let executionTime = Date().timeIntervalSince(startTime) * 1000

        switch result {
        case .success(let transformedURL):
            return .success(
                original: url,
                transformed: transformedURL,
                executionTime: executionTime
            )
        case .failure(let error):
            return .failure(
                original: url,
                error: error.localizedDescription,
                executionTime: executionTime
            )
        }
    }
}
```

---

## CloudKit Schema Extensions

### CKRecord: Rule

The existing "Rule" record type needs to support two new optional fields:

```swift
// In CloudKitSyncManager.swift

func ruleToRecord(_ rule: Rule) -> CKRecord {
    let recordID = CKRecord.ID(recordName: rule.recordID ?? UUID().uuidString)
    let record = CKRecord(recordType: "Rule", recordID: recordID)

    // === EXISTING FIELDS (not shown) ===
    record["name"] = rule.name
    record["type"] = rule.type.rawValue
    // ... etc ...

    // === NEW FIELDS ===
    if let transformationType = rule.transformationType {
        record["transformationType"] = transformationType.rawValue
    }

    if let transformationLogic = rule.transformationLogic {
        record["transformationLogic"] = transformationLogic
    }

    return record
}

func recordToRule(_ record: CKRecord) -> Rule {
    var rule = Rule(/* existing fields */)

    // === NEW FIELDS (optional) ===
    if let typeString = record["transformationType"] as? String,
       let type = Rule.TransformationType(rawValue: typeString) {
        rule.transformationType = type
    }

    if let logic = record["transformationLogic"] as? String {
        rule.transformationLogic = logic
    }

    return rule
}
```

**CloudKit Field Specifications**:

| Field Name | CKRecord Type | Indexed | Queryable | Notes |
|------------|---------------|---------|-----------|-------|
| `transformationType` | String | No | Yes | Values: "template" or "javascript" |
| `transformationLogic` | String | No | No | Max 10KB; UTF-8 encoded |

**Backward Compatibility**:
- Existing records without transformation fields will have `nil` for both fields
- Decoding gracefully handles missing fields (optional properties)
- Old app versions will ignore transformation fields (won't break)
- New app versions will preserve transformation fields when syncing

---

## Persistence Schema (UserDefaults)

Rules are stored as JSON-encoded array in UserDefaults under key "rules".

**JSON Structure** (single rule with transformation):

```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "name": "Strip Tracking Parameters",
  "type": "domain",
  "domains": ["*"],
  "browserTarget": {
    "type": "browser",
    "browserId": "com.apple.Safari"
  },
  "priority": 5,
  "isEnabled": true,
  "transformationType": "javascript",
  "transformationLogic": "// Remove utm_* params\nconst url = new URL(input.url);\nfor (const key of [...url.searchParams.keys()]) {\n  if (key.startsWith('utm_')) {\n    url.searchParams.delete(key);\n  }\n}\nreturn url.toString();",
  "lastModified": "2025-10-31T12:00:00Z",
  "deviceID": "device-fingerprint-123"
}
```

**Encoding/Decoding**:
```swift
// Encoding
let encoder = JSONEncoder()
encoder.dateEncodingStrategy = .iso8601
let data = try encoder.encode(rules)
UserDefaults.standard.set(data, forKey: "rules")

// Decoding
let decoder = JSONDecoder()
decoder.dateDecodingStrategy = .iso8601
if let data = UserDefaults.standard.data(forKey: "rules") {
    let rules = try decoder.decode([Rule].self, from: data)
}
```

---

## Migration Strategy

### From Current Schema to Transformation-Enabled Schema

**Migration Version**: 1.0.0 → 1.1.0

**Changes**:
- Add two optional fields to Rule model
- No data migration needed (fields are optional)
- CloudKit automatically handles new fields on sync

**Migration Code** (none needed, but validation check):

```swift
// In PersistenceManager
func validateRules() {
    let rules = loadRules()

    for rule in rules {
        // Validate transformation consistency
        if rule.transformationType != nil && rule.transformationLogic == nil {
            print("Warning: Rule \(rule.id) has transformationType but no logic")
        }

        if rule.transformationLogic != nil && rule.transformationType == nil {
            print("Warning: Rule \(rule.id) has logic but no transformationType")
        }
    }
}
```

**Rollback Strategy**:
- If rolling back from 1.1.0 to 1.0.0, transformation fields are ignored
- Rules continue to work as domain/time/focus rules
- Transformations won't be applied but rules still match URLs

---

## Validation Rules Summary

### Rule-Level Validation

- `transformationType` and `transformationLogic` must both be present or both be absent
- `transformationLogic` must not be empty if `transformationType` is set
- `transformationLogic` length ≤ 10,000 characters
- Template syntax must be valid (checked by TemplateEngine.validate())
- JavaScript syntax must be valid (checked by JSContext compilation)

### Runtime Validation

- Transformed URL must be valid (passes URLComponents parsing)
- JavaScript execution must complete within 1 second
- JavaScript must return a string
- Circular transformation check (URL unchanged) shows warning but doesn't block

### UI Validation

Real-time validation in rule editor:
- Syntax errors highlighted with line numbers (JavaScript)
- Invalid template variables shown with suggestions
- Preview shows transformation result before saving
- Test URLs can be entered to verify transformation logic

---

## Performance Characteristics

### Storage

| Metric | Value | Notes |
|--------|-------|-------|
| Average transformation size | 500 bytes | Typical JavaScript or template |
| Max transformation size | 10 KB | Hard limit enforced |
| Rules per user (estimated) | 1-5 transformations | Most users will have few |
| Total storage overhead | ~2.5 KB | Average case (5 rules × 500 bytes) |

### Memory

| Component | Memory Usage |
|-----------|--------------|
| Rule model (with transformation) | ~600 bytes |
| URLComponentsContext | ~200 bytes |
| TransformationResult | ~150 bytes |
| Regex pattern cache | ~500 bytes |
| JSContext (temporary) | ~2 MB (created per execution) |

### Execution Time

| Operation | Target | Typical |
|-----------|--------|---------|
| Template substitution | <10ms p95 | 3-5ms |
| JavaScript execution | <100ms p95 | 20-50ms |
| URL validation | <5ms | 1-2ms |
| Total transformation | <100ms p95 | 25-60ms |

---

## Summary

The data model extends the existing `Rule` entity with two optional fields (`transformationType`, `transformationLogic`), maintaining full backward compatibility while enabling powerful URL transformation capabilities. The design prioritizes:

- **Simplicity**: Minimal schema changes, reuse existing infrastructure
- **Backward compatibility**: Old versions ignore new fields gracefully
- **Performance**: Lightweight objects, stateless processing
- **Flexibility**: Supports both template and JavaScript transformations
- **Testability**: Transient result objects enable comprehensive testing

All validation rules are clearly defined and enforced at both rule creation time and runtime, with graceful error handling and fallback behavior.
