# Developer Quickstart: Custom URL Transformations

**Feature**: 001-custom-url-transform
**Date**: 2025-10-31
**Audience**: Developers implementing this feature

## Overview

This guide helps developers quickly understand and implement the custom URL transformation feature. It provides concrete examples, code snippets, and implementation order.

---

## Prerequisites

Before starting implementation:

1. **Read these documents** (in order):
   - `spec.md` - Feature requirements and user stories
   - `research.md` - Technical decisions and best practices
   - `data-model.md` - Data structures and validation rules
   - `contracts/*.swift` - API contracts and protocols

2. **Understand existing code**:
   - `Proxly/Models/Rule.swift` - Current rule model
   - `Proxly/Services/URLProcessor.swift` - URL processing flow
   - `Proxly/Services/URLProcessingEngine.swift` - Pattern matching engine
   - `Proxly/Services/PersistenceManager.swift` - Data persistence
   - `Proxly/Services/CloudKitSyncManager.swift` - CloudKit sync

3. **Set up development environment**:
   ```bash
   cd proxly
   open Proxly.xcodeproj
   ```

---

## Implementation Order

Follow this order to minimize dependencies and enable incremental testing:

### Phase 1: Data Models (Day 1)
1. Extend `Rule` model with transformation fields
2. Add `TransformationResult` struct
3. Add `URLComponentsContext` struct
4. Update CloudKit sync to handle new fields
5. Write unit tests for model encoding/decoding

### Phase 2: Template Engine (Day 2)
1. Implement `TemplateEngine.swift`
2. Add regex pattern caching
3. Implement variable resolution
4. Add template validation
5. Write comprehensive unit tests

### Phase 3: JavaScript Engine (Day 3-4)
1. Implement `JavaScriptEngine.swift`
2. Add JSContext creation and configuration
3. Implement timeout mechanism
4. Add syntax validation
5. Write unit tests including timeout scenarios

### Phase 4: Transformation Engine (Day 4)
1. Implement `TransformationEngine.swift`
2. Integrate template and JavaScript engines
3. Add error handling and fallback logic
4. Write integration tests

### Phase 5: URLProcessor Integration (Day 5)
1. Extend `URLProcessor` to check for transformations
2. Apply transformation before browser launch
3. Add transformation logging
4. Write end-to-end tests

### Phase 6: UI Components (Day 6-7)
1. Create transformation editor view
2. Implement real-time preview
3. Add validation feedback UI
4. Add import/export functionality
5. Write UI tests

### Phase 7: Localization & Polish (Day 8)
1. Extract all strings to Localizable.strings
2. Translate to 7 languages
3. Add accessibility labels
4. Test in Light and Dark mode

---

## Code Examples

### 1. Extending Rule Model

```swift
// In Proxly/Models/Rule.swift

struct Rule: Codable, Identifiable {
    // ... existing fields ...

    // NEW: Transformation fields
    var transformationType: TransformationType?
    var transformationLogic: String?

    enum TransformationType: String, Codable {
        case template
        case javascript
    }

    // Validation
    func validateTransformation() -> Bool {
        if transformationType != nil && transformationLogic == nil {
            return false
        }
        if transformationLogic != nil && transformationType == nil {
            return false
        }
        if let logic = transformationLogic, logic.isEmpty {
            return false
        }
        return true
    }
}
```

### 2. Template Engine Implementation

```swift
// Create file: Proxly/Services/TemplateEngine.swift

import Foundation

struct TemplateEngine: TemplateEngineProtocol {
    private static let cache = NSCache<NSString, NSRegularExpression>()
    private static let pattern = #"\{([a-zA-Z][a-zA-Z0-9]*(?::[a-zA-Z0-9_-]+)?)\}"#

    static func substitute(
        _ template: String,
        context: URLComponentsContext
    ) -> Result<URL, TransformationError> {
        // Get compiled regex from cache
        guard let regex = getCompiledRegex() else {
            return .failure(.invalidTemplate("Failed to compile regex"))
        }

        var result = template
        let nsString = template as NSString
        let range = NSRange(location: 0, length: nsString.length)

        // Find all matches in reverse order (to preserve indices)
        let matches = regex.matches(in: template, range: range).reversed()

        for match in matches {
            let matchRange = match.range(at: 1)
            let variable = nsString.substring(with: matchRange)

            // Resolve variable
            guard let value = resolveVariable(variable, from: context) else {
                return .failure(.missingVariable(variable))
            }

            // Replace in result string
            let fullMatchRange = match.range(at: 0)
            result = (result as NSString)
                .replacingCharacters(in: fullMatchRange, with: value)
        }

        // Validate URL
        guard let url = URL(string: result) else {
            return .failure(.invalidURLOutput(result))
        }

        return .success(url)
    }

    static func validate(_ template: String) -> Result<Void, TransformationError> {
        guard let regex = getCompiledRegex() else {
            return .failure(.invalidTemplate("Invalid regex pattern"))
        }

        let nsString = template as NSString
        let range = NSRange(location: 0, length: nsString.length)
        let matches = regex.matches(in: template, range: range)

        // Check all variables are valid
        for match in matches {
            let matchRange = match.range(at: 1)
            let variable = nsString.substring(with: matchRange)

            // Validate variable name
            if !isValidVariable(variable) {
                return .failure(.invalidTemplate("Invalid variable: {\(variable)}"))
            }
        }

        return .success(())
    }

    // MARK: - Private

    private static func resolveVariable(
        _ variable: String,
        from context: URLComponentsContext
    ) -> String? {
        let parts = variable.split(separator: ":")

        switch parts[0] {
        case "scheme": return context.scheme
        case "host": return context.host
        case "port": return context.port.map(String.init)
        case "path": return context.path
        case "query":
            if parts.count == 2 {
                return context.queryParam(String(parts[1]))
            } else {
                return context.query.map { "?\($0)" }
            }
        case "fragment":
            return context.fragment.map { "#\($0)" }
        case "pathComponent":
            if parts.count == 2, let index = Int(parts[1]) {
                return context.pathComponent(at: index)
            }
            return nil
        default: return nil
        }
    }

    private static func isValidVariable(_ variable: String) -> Bool {
        let validNames = ["scheme", "host", "port", "path", "query", "fragment", "pathComponent"]
        let parts = variable.split(separator: ":")
        return validNames.contains(String(parts[0]))
    }

    private static func getCompiledRegex() -> NSRegularExpression? {
        let key = pattern as NSString
        if let cached = cache.object(forKey: key) {
            return cached
        }

        guard let regex = try? NSRegularExpression(pattern: pattern) else {
            return nil
        }

        cache.setObject(regex, forKey: key)
        return regex
    }
}
```

### 3. JavaScript Engine Implementation

```swift
// Create file: Proxly/Services/JavaScriptEngine.swift

import Foundation
import JavaScriptCore

struct JavaScriptEngine: JavaScriptEngineProtocol {
    static func execute(
        _ script: String,
        context: URLComponentsContext,
        timeout: TimeInterval = 1.0
    ) -> Result<URL, TransformationError> {
        let jsContext = JSContext()!
        var result: Result<URL, TransformationError>?

        // Set up error handler
        jsContext.exceptionHandler = { _, exception in
            let message = exception?.toString() ?? "Unknown error"
            result = .failure(.javascriptRuntimeError(message))
        }

        // Prepare input object
        jsContext.setObject(
            context.toDictionary(),
            forKeyedSubscript: "input" as NSString
        )

        // Execute with timeout
        let workItem = DispatchWorkItem {
            guard let jsResult = jsContext.evaluateScript(script) else {
                result = .failure(.javascriptRuntimeError("Script returned nil"))
                return
            }

            guard let urlString = jsResult.toString() else {
                result = .failure(.javascriptRuntimeError("Script did not return string"))
                return
            }

            guard let url = URL(string: urlString) else {
                result = .failure(.invalidURLOutput(urlString))
                return
            }

            result = .success(url)
        }

        DispatchQueue.global(qos: .userInitiated).async(execute: workItem)

        let timeoutResult = workItem.wait(timeout: .now() + timeout)

        if timeoutResult == .timedOut {
            workItem.cancel()
            return .failure(.timeout)
        }

        return result ?? .failure(.javascriptRuntimeError("No result"))
    }

    static func validateSyntax(_ script: String) -> Result<Void, TransformationError> {
        let jsContext = JSContext()!
        var syntaxError: (String, Int?)?

        jsContext.exceptionHandler = { _, exception in
            syntaxError = (
                exception?.toString() ?? "Syntax error",
                exception?.objectForKeyedSubscript("line").toInt32()
            )
        }

        // Try to compile (not execute)
        _ = jsContext.evaluateScript("(function() { \(script) })")

        if let (message, line) = syntaxError {
            return .failure(.javascriptSyntaxError(message, line: line))
        }

        return .success(())
    }
}
```

### 4. URLProcessor Integration

```swift
// In Proxly/Services/URLProcessor.swift

extension URLProcessor {
    func processURL(_ url: URL, opener: OpenerInfo?) async -> URL {
        // ... existing code to find matching rule ...

        guard let rule = URLProcessingEngine.findBestRule(
            for: url,
            rules: rules,
            focusMode: focusMode,
            opener: opener
        ) else {
            // No rule matched, use fallback
            return url
        }

        // NEW: Apply transformation if rule has one
        let finalURL: URL
        if rule.transformationType != nil {
            let result = applyTransformation(for: rule, to: url)
            if result.success, let transformed = result.transformedURL {
                finalURL = transformed
                AppLogger.log(
                    "Applied transformation: \(url) -> \(finalURL)",
                    level: .info
                )
            } else {
                finalURL = url
                AppLogger.log(
                    "Transformation failed, using original URL: \(result.errorMessage ?? "unknown")",
                    level: .warning
                )
            }
        } else {
            finalURL = url
        }

        // ... existing code to launch browser ...
        await BrowserLauncher.launch(
            url: finalURL,
            browserTarget: rule.browserTarget
        )

        return finalURL
    }

    private func applyTransformation(
        for rule: Rule,
        to url: URL
    ) -> TransformationResult {
        TransformationEngine.apply(transformation: rule, to: url)
    }
}
```

### 5. UI Editor View

```swift
// Create file: Proxly/UI/TransformationEditorView.swift

import SwiftUI

struct TransformationEditorView: View {
    @StateObject var viewModel: TransformationEditorViewModel
    @Environment(\.dismiss) var dismiss

    var body: some View {
        VStack(spacing: 16) {
            // Header
            Text("URL Transformation")
                .font(.headline)

            // Type picker
            Picker("Type", selection: $viewModel.transformationType) {
                Text("Template").tag(Rule.TransformationType.template)
                Text("JavaScript").tag(Rule.TransformationType.javascript)
            }
            .pickerStyle(.segmented)

            // Logic editor
            if viewModel.transformationType == .template {
                templateEditor
            } else {
                javascriptEditor
            }

            // Test URL
            TextField("Test URL", text: $viewModel.testURL)
                .textFieldStyle(.roundedBorder)

            // Preview
            if let result = viewModel.previewResult {
                TransformationPreviewView(result: result)
            }

            // Validation errors
            if !viewModel.validationErrors.isEmpty {
                VStack(alignment: .leading, spacing: 4) {
                    ForEach(viewModel.validationErrors, id: \.self) { error in
                        Text(error)
                            .font(.caption)
                            .foregroundColor(.red)
                    }
                }
            }

            // Actions
            HStack {
                Button("Cancel") {
                    dismiss()
                }
                Spacer()
                Button("Save") {
                    viewModel.save()
                    dismiss()
                }
                .disabled(!viewModel.canSave)
            }
        }
        .padding()
        .frame(width: 600, height: 500)
    }

    var templateEditor: some View {
        VStack(alignment: .leading) {
            Text("Template (use {scheme}, {host}, {path}, {query}, etc.)")
                .font(.caption)
                .foregroundColor(.secondary)

            TextEditor(text: $viewModel.transformationLogic)
                .font(.system(.body, design: .monospaced))
                .frame(height: 150)
                .border(Color.gray.opacity(0.3))
        }
    }

    var javascriptEditor: some View {
        VStack(alignment: .leading) {
            Text("JavaScript (input.url, input.scheme, input.host, etc.)")
                .font(.caption)
                .foregroundColor(.secondary)

            TextEditor(text: $viewModel.transformationLogic)
                .font(.system(.body, design: .monospaced))
                .frame(height: 200)
                .border(Color.gray.opacity(0.3))
        }
    }
}
```

---

## Testing Strategy

### Unit Tests

```swift
// Tests/TransformationEngineTests.swift

import XCTest
@testable import Proxly

class TemplateEngineTests: XCTestCase {
    func testBasicSubstitution() {
        let url = URL(string: "https://twitter.com/user")!
        let components = URLComponents(url: url, resolvingAgainstBaseURL: true)!
        let context = URLComponentsContext(url: url, components: components)

        let template = "{scheme}://nitter.net{path}"
        let result = TemplateEngine.substitute(template, context: context)

        XCTAssertEqual(
            try? result.get().absoluteString,
            "https://nitter.net/user"
        )
    }

    func testQueryParameterExtraction() {
        let url = URL(string: "https://example.com?token=abc123")!
        let components = URLComponents(url: url, resolvingAgainstBaseURL: true)!
        let context = URLComponentsContext(url: url, components: components)

        let template = "{scheme}://{host}/?key={query:token}"
        let result = TemplateEngine.substitute(template, context: context)

        XCTAssertEqual(
            try? result.get().absoluteString,
            "https://example.com/?key=abc123"
        )
    }
}

class JavaScriptEngineTests: XCTestCase {
    func testBasicExecution() {
        let url = URL(string: "https://twitter.com/user")!
        let components = URLComponents(url: url, resolvingAgainstBaseURL: true)!
        let context = URLComponentsContext(url: url, components: components)

        let script = "return input.url.replace('twitter.com', 'nitter.net');"
        let result = JavaScriptEngine.execute(script, context: context)

        XCTAssertEqual(
            try? result.get().absoluteString,
            "https://nitter.net/user"
        )
    }

    func testTimeout() {
        let url = URL(string: "https://example.com")!
        let components = URLComponents(url: url, resolvingAgainstBaseURL: true)!
        let context = URLComponentsContext(url: url, components: components)

        let script = "while(true) {}" // Infinite loop
        let result = JavaScriptEngine.execute(script, context: context, timeout: 0.1)

        if case .failure(.timeout) = result {
            XCTAssert(true)
        } else {
            XCTFail("Expected timeout error")
        }
    }
}
```

---

## Common Pitfalls

### 1. Forgetting to validate before saving
Always validate transformation logic before saving the rule:
```swift
guard viewModel.validate() else {
    showError("Invalid transformation")
    return
}
```

### 2. Not handling nil transformations
Check if transformation exists before applying:
```swift
guard rule.transformationType != nil else {
    return url // No transformation
}
```

### 3. Blocking main thread with JavaScript
JavaScript execution is synchronous but fast enough (<100ms). If you find it blocking UI, consider dispatching to background queue:
```swift
// Already handled in JavaScriptEngine implementation
```

### 4. Forgetting CloudKit sync
Remember to update CloudKitSyncManager to include new fields:
```swift
record["transformationType"] = rule.transformationType?.rawValue
record["transformationLogic"] = rule.transformationLogic
```

### 5. Missing localization
Extract all strings:
```swift
Text(NSLocalizedString("transformation_template_help", comment: ""))
```

---

## Performance Benchmarks

Test your implementation against these targets:

```swift
// Performance tests
func testTemplatePerformance() {
    measure {
        // Should complete in <10ms
        _ = TemplateEngine.substitute(template, context: context)
    }
}

func testJavaScriptPerformance() {
    measure {
        // Should complete in <100ms for simple scripts
        _ = JavaScriptEngine.execute(script, context: context)
    }
}
```

Expected results:
- Template substitution: **3-5ms** average
- JavaScript execution: **20-50ms** average
- End-to-end transformation: **25-60ms** average

---

## Debugging Tips

### Enable verbose logging
```swift
AppLogger.setLevel(.debug)
AppLogger.log("Transformation input: \(url)", level: .debug)
AppLogger.log("Transformation output: \(result)", level: .debug)
```

### Test transformations in isolation
```swift
let url = URL(string: "https://example.com/path?foo=bar")!
let components = URLComponents(url: url, resolvingAgainstBaseURL: true)!
let context = URLComponentsContext(url: url, components: components)

// Test template
let templateResult = TemplateEngine.substitute("{scheme}://{host}", context: context)
print(templateResult)

// Test JavaScript
let jsResult = JavaScriptEngine.execute("return input.url.toUpperCase();", context: context)
print(jsResult)
```

### Inspect JavaScript context
```swift
let jsContext = JSContext()!
jsContext.setObject(context.toDictionary(), forKeyedSubscript: "input" as NSString)

// Check what's available
print(jsContext.evaluateScript("Object.keys(input)"))
// Output: ["url", "scheme", "host", "path", "query", "queryParams", "pathComponents"]
```

---

## Next Steps

After completing implementation:

1. Run `/speckit.tasks` to generate task breakdown
2. Implement features in priority order (P1 → P2 → P3)
3. Write tests as you go (TDD recommended)
4. Test on real URLs with actual transformations
5. Get user feedback on UI/UX
6. Iterate based on feedback

For questions, refer to:
- **Technical details**: research.md
- **Data structures**: data-model.md
- **API contracts**: contracts/README.md
- **User requirements**: spec.md
