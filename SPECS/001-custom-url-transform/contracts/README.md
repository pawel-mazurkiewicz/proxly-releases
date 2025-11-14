# API Contracts: Custom URL Transformations

**Feature**: 001-custom-url-transform
**Date**: 2025-10-31

## Overview

This directory contains Swift protocol definitions that serve as contracts between components in the custom URL transformation feature. Since Proxly is a native macOS app (not a web service), these contracts define Swift interfaces rather than REST/GraphQL schemas.

## Contract Files

### 1. TransformationEngine.swift

Defines the core transformation engine interfaces:
- `TransformationEngineProtocol`: Main interface for applying transformations
- `TemplateEngineProtocol`: Template-based transformation contract
- `JavaScriptEngineProtocol`: JavaScript-based transformation contract
- `TransformationError`: Error types for transformation failures
- `TransformationResult`: Result object containing transformation outcome
- `URLComponentsContext`: Parsed URL components for transformation input

**Implementers**:
- `Proxly/Services/TransformationEngine.swift` (implements `TransformationEngineProtocol`)
- `Proxly/Services/TemplateEngine.swift` (implements `TemplateEngineProtocol`)
- `Proxly/Services/JavaScriptEngine.swift` (implements `JavaScriptEngineProtocol`)

**Key Design Decisions**:
- All engines are stateless (static methods)
- Engines use `Result` type for explicit error handling
- Context object provides unified URL component access
- Timeout is configurable but defaults to 1 second

### 2. TransformationUI.swift

Defines UI component contracts:
- `TransformationEditorViewModelProtocol`: View model for transformation editor
- `TransformationPreviewProtocol`: Preview component interface
- `TransformationEditorState`: UI state machine
- `ValidationResult`: Structured validation feedback

**Implementers**:
- `Proxly/ViewModels/TransformationEditorViewModel.swift`
- `Proxly/UI/Components/TransformationPreview.swift`

**Key Design Decisions**:
- Real-time validation with live preview
- Separate errors from warnings (warnings don't block saving)
- Line/column information for JavaScript errors
- SwiftUI `ObservableObject` pattern for reactive updates

### 3. URLProcessor.swift

Defines integration contracts with existing URLProcessor:
- `URLProcessorTransformationProtocol`: Extensions to URLProcessor interface
- `TransformationLoggerProtocol`: Logging interface for transformation events
- `TransformationLog`: Log entry structure

**Implementers**:
- `Proxly/Services/URLProcessor.swift` (extended with transformation support)
- `Proxly/Services/TransformationLogger.swift` (new)

**Key Design Decisions**:
- Transformations integrate seamlessly into existing URL processing flow
- Errors never block URL opening (fallback to original URL)
- Logging is optional but useful for debugging
- Logs stored in memory (last 100 entries) for performance

## Contract Usage Examples

### Using TransformationEngine

```swift
// Apply transformation
let rule = Rule(
    transformationType: .template,
    transformationLogic: "{scheme}://nitter.net{path}{query}"
)

let url = URL(string: "https://twitter.com/user/status/123")!
let result = TransformationEngine.apply(transformation: rule, to: url)

if result.success, let transformedURL = result.transformedURL {
    // Use transformed URL
    print("Opening: \(transformedURL)")
} else {
    // Fall back to original
    print("Using original URL due to error: \(result.errorMessage ?? "")")
}
```

### Using TemplateEngine

```swift
// Create context
let url = URL(string: "https://example.com/path?foo=bar")!
let components = URLComponents(url: url, resolvingAgainstBaseURL: true)!
let context = URLComponentsContext(url: url, components: components)

// Substitute template
let template = "{scheme}://{host}/modified{path}"
let result = TemplateEngine.substitute(template, context: context)

switch result {
case .success(let transformedURL):
    print("Result: \(transformedURL)")
case .failure(let error):
    print("Error: \(error.localizedDescription)")
}
```

### Using JavaScriptEngine

```swift
// Execute JavaScript
let script = """
if (input.host.includes('reddit')) {
    return input.url.replace('m.reddit', 'old.reddit');
}
return input.url;
"""

let context = URLComponentsContext(url: url, components: components)
let result = JavaScriptEngine.execute(script, context: context, timeout: 1.0)

switch result {
case .success(let transformedURL):
    print("Transformed: \(transformedURL)")
case .failure(.timeout):
    print("JavaScript execution timed out")
case .failure(let error):
    print("Error: \(error.localizedDescription)")
}
```

### Using TransformationEditorViewModel

```swift
// In SwiftUI view
@StateObject var viewModel = TransformationEditorViewModel(rule: rule)

var body: some View {
    VStack {
        // Editor fields
        Picker("Type", selection: $viewModel.transformationType) {
            Text("Template").tag(Rule.TransformationType.template)
            Text("JavaScript").tag(Rule.TransformationType.javascript)
        }

        TextEditor(text: $viewModel.transformationLogic)
            .onChange(of: viewModel.transformationLogic) { _ in
                viewModel.updatePreview()
            }

        // Test URL field
        TextField("Test URL", text: $viewModel.testURL)
            .onChange(of: viewModel.testURL) { _ in
                viewModel.updatePreview()
            }

        // Preview
        if let result = viewModel.previewResult {
            TransformationPreview(result: result)
        }

        // Validation errors
        ForEach(viewModel.validationErrors, id: \.self) { error in
            Text(error).foregroundColor(.red)
        }

        // Save button
        Button("Save") {
            viewModel.save()
        }
        .disabled(!viewModel.canSave)
    }
}
```

## Testing Contracts

All protocols should have corresponding test implementations:

### MockTransformationEngine

```swift
class MockTransformationEngine: TransformationEngineProtocol {
    var applyCallCount = 0
    var applyResult: TransformationResult?

    static func apply(
        transformation rule: Rule,
        to url: URL
    ) -> TransformationResult {
        applyCallCount += 1
        return applyResult ?? .success(
            original: url,
            transformed: url,
            executionTime: 0.0
        )
    }
}
```

### MockTemplateEngine

```swift
class MockTemplateEngine: TemplateEngineProtocol {
    static var substituteResult: Result<URL, TransformationError> = .success(
        URL(string: "https://example.com")!
    )

    static func substitute(
        _ template: String,
        context: URLComponentsContext
    ) -> Result<URL, TransformationError> {
        return substituteResult
    }

    static func validate(_ template: String) -> Result<Void, TransformationError> {
        return .success(())
    }
}
```

## Contract Versioning

Contracts follow semantic versioning:
- **Major version**: Breaking changes to protocols (method signature changes, removed methods)
- **Minor version**: Added methods with default implementations
- **Patch version**: Documentation updates, clarifications

Current version: **1.0.0**

## Contract Compliance Checklist

When implementing these contracts:

- [ ] All protocol methods implemented
- [ ] Error cases handled gracefully
- [ ] Stateless implementation (for engines)
- [ ] Thread-safe (if used from multiple threads)
- [ ] Unit tests covering all protocol methods
- [ ] Documentation strings added to implementation
- [ ] Performance targets met (see research.md)
- [ ] Accessibility labels for UI components
- [ ] Localized strings for user-facing messages

## Contract Evolution

To modify contracts:

1. Propose changes in feature planning phase
2. Check backward compatibility impact
3. Update all implementers simultaneously
4. Bump contract version appropriately
5. Update this README with migration notes
6. Add deprecation warnings for removed methods (if applicable)

For questions about contracts, see:
- **Implementation details**: data-model.md
- **Technical decisions**: research.md
- **Integration flow**: plan.md
