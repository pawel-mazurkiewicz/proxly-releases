// URLProcessor Contract Extensions
// Feature: 001-custom-url-transform
// Purpose: Defines how URLProcessor integrates with transformations

import Foundation

/// Extension to URLProcessor interface for transformation support
protocol URLProcessorTransformationProtocol {
    /// Process URL with transformation support
    /// - Parameters:
    ///   - url: The URL to process
    ///   - opener: Information about the app that opened the URL
    /// - Returns: The final URL to open (potentially transformed)
    func processURL(
        _ url: URL,
        opener: OpenerInfo?
    ) async -> URL

    /// Apply transformation if rule specifies one
    /// - Parameters:
    ///   - rule: The matched rule
    ///   - url: The URL to transform
    /// - Returns: Transformation result
    func applyTransformation(
        for rule: Rule,
        to url: URL
    ) -> TransformationResult
}

/// Integration flow documentation
///
/// URL Processing Pipeline with Transformations:
///
/// 1. URLProcessor.processURL(_:opener:) receives URL
/// 2. Decode proxly:// scheme if needed
/// 3. Check if rules are enabled
/// 4. URLProcessingEngine.findBestRule() finds matching rule
/// 5. **NEW**: If rule has transformation, apply it
///    - TransformationEngine.apply(transformation:to:) → TransformationResult
///    - If success: use transformed URL
///    - If failure: use original URL (fallback)
/// 6. BrowserLauncher.launch() with final URL
/// 7. Log transformation result for debugging
///
/// Error Handling:
/// - Transformation errors DO NOT block URL opening
/// - Fall back to original URL on any transformation error
/// - Log errors for user debugging
/// - Optionally show notification (user preference)

/// Logging protocol for transformation events
protocol TransformationLoggerProtocol {
    /// Log successful transformation
    func logSuccess(
        rule: Rule,
        originalURL: URL,
        transformedURL: URL,
        executionTime: Double
    )

    /// Log transformation failure
    func logFailure(
        rule: Rule,
        originalURL: URL,
        error: TransformationError,
        executionTime: Double
    )

    /// Log transformation timeout
    func logTimeout(
        rule: Rule,
        originalURL: URL
    )

    /// Retrieve recent transformation logs
    func recentLogs(limit: Int) -> [TransformationLog]
}

/// Transformation log entry
struct TransformationLog: Identifiable, Codable {
    let id: UUID
    let timestamp: Date
    let ruleName: String
    let originalURL: String
    let transformedURL: String?
    let success: Bool
    let errorMessage: String?
    let executionTimeMs: Double

    init(
        ruleName: String,
        originalURL: String,
        transformedURL: String?,
        success: Bool,
        errorMessage: String?,
        executionTimeMs: Double
    ) {
        self.id = UUID()
        self.timestamp = Date()
        self.ruleName = ruleName
        self.originalURL = originalURL
        self.transformedURL = transformedURL
        self.success = success
        self.errorMessage = errorMessage
        self.executionTimeMs = executionTimeMs
    }
}
