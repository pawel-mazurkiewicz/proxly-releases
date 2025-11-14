// TransformationEngine Contract
// Feature: 001-custom-url-transform
// Purpose: Defines the interface for URL transformation services

import Foundation

/// Protocol defining the contract for transformation engines
/// Implementations must be stateless and thread-safe
protocol TransformationEngineProtocol {
    /// Apply transformation to a URL
    /// - Parameters:
    ///   - rule: The rule containing transformation logic
    ///   - url: The URL to transform
    /// - Returns: Result containing transformed URL or transformation details
    static func apply(
        transformation rule: Rule,
        to url: URL
    ) -> TransformationResult
}

/// Protocol for template-based transformation engine
protocol TemplateEngineProtocol {
    /// Substitute template variables with actual URL components
    /// - Parameters:
    ///   - template: Template string with {variable} placeholders
    ///   - context: URL components context
    /// - Returns: Result containing transformed URL string or error
    static func substitute(
        _ template: String,
        context: URLComponentsContext
    ) -> Result<URL, TransformationError>

    /// Validate template syntax
    /// - Parameter template: Template string to validate
    /// - Returns: Result with nil for success or error description
    static func validate(_ template: String) -> Result<Void, TransformationError>
}

/// Protocol for JavaScript-based transformation engine
protocol JavaScriptEngineProtocol {
    /// Execute JavaScript transformation code
    /// - Parameters:
    ///   - script: JavaScript code to execute
    ///   - context: URL components context
    ///   - timeout: Maximum execution time in seconds (default: 1.0)
    /// - Returns: Result containing transformed URL or error
    static func execute(
        _ script: String,
        context: URLComponentsContext,
        timeout: TimeInterval
    ) -> Result<URL, TransformationError>

    /// Validate JavaScript syntax without executing
    /// - Parameter script: JavaScript code to validate
    /// - Returns: Result with nil for success or error with line number
    static func validateSyntax(_ script: String) -> Result<Void, TransformationError>
}

/// Transformation error types
enum TransformationError: Error, LocalizedError {
    case invalidTemplate(String)
    case missingVariable(String)
    case javascriptSyntaxError(String, line: Int?)
    case javascriptRuntimeError(String)
    case timeout
    case invalidURLOutput(String)
    case circularTransformation
    case emptyTransformationLogic

    var errorDescription: String? {
        switch self {
        case .invalidTemplate(let msg):
            return "Invalid template: \(msg)"
        case .missingVariable(let varName):
            return "Missing or undefined variable: {\(varName)}"
        case .javascriptSyntaxError(let msg, let line):
            if let line = line {
                return "JavaScript syntax error at line \(line): \(msg)"
            }
            return "JavaScript syntax error: \(msg)"
        case .javascriptRuntimeError(let msg):
            return "JavaScript execution error: \(msg)"
        case .timeout:
            return "Transformation timed out after 1 second"
        case .invalidURLOutput(let output):
            return "Transformation produced invalid URL: \(output)"
        case .circularTransformation:
            return "Transformation produces unchanged URL"
        case .emptyTransformationLogic:
            return "Transformation logic cannot be empty"
        }
    }
}

/// Result of applying a transformation
struct TransformationResult {
    let originalURL: URL
    let transformedURL: URL?
    let success: Bool
    let errorMessage: String?
    let executionTimeMs: Double
    let usedFallback: Bool

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

/// Context containing parsed URL components for transformation
struct URLComponentsContext {
    let url: URL
    let components: URLComponents

    var scheme: String { components.scheme ?? "" }
    var host: String { components.host ?? "" }
    var port: Int? { components.port }
    var path: String { components.path }
    var query: String? { components.query }
    var fragment: String? { components.fragment }
    var queryItems: [URLQueryItem] { components.queryItems ?? [] }

    var pathComponents: [String] {
        path.split(separator: "/").map(String.init)
    }

    /// Convert to dictionary for JavaScript context
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

        if !queryItems.isEmpty {
            var queryDict: [String: String] = [:]
            for item in queryItems {
                queryDict[item.name] = item.value ?? ""
            }
            dict["queryParams"] = queryDict
        }

        dict["pathComponents"] = pathComponents

        return dict
    }

    func queryParam(_ name: String) -> String? {
        queryItems.first(where: { $0.name == name })?.value
    }

    func pathComponent(at index: Int) -> String? {
        let components = pathComponents
        return components.indices.contains(index) ? components[index] : nil
    }
}
