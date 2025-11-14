// TransformationUI Contract
// Feature: 001-custom-url-transform
// Purpose: Defines the interface for UI components related to transformations

import Foundation
import SwiftUI

/// Protocol for transformation editor view model
protocol TransformationEditorViewModelProtocol: ObservableObject {
    /// The rule being edited
    var rule: Rule { get set }

    /// Selected transformation type (template or javascript)
    var transformationType: Rule.TransformationType { get set }

    /// Transformation logic (template or JS code)
    var transformationLogic: String { get set }

    /// Test URL for preview
    var testURL: String { get set }

    /// Preview result
    var previewResult: TransformationResult? { get }

    /// Validation errors
    var validationErrors: [String] { get }

    /// Whether save button should be enabled
    var canSave: Bool { get }

    /// Update preview with current logic and test URL
    func updatePreview()

    /// Validate transformation logic
    func validate() -> Bool

    /// Save transformation to rule
    func save()

    /// Remove transformation from rule
    func removeTransformation()
}

/// Protocol for transformation preview component
protocol TransformationPreviewProtocol {
    /// Display transformation result
    /// - Parameter result: The transformation result to display
    func display(result: TransformationResult)

    /// Show loading state while transformation executes
    func showLoading()

    /// Show error state
    /// - Parameter error: Error message to display
    func showError(_ error: String)
}

/// View state for transformation editor
enum TransformationEditorState {
    case editing
    case previewing
    case validating
    case saving
    case error(String)
}

/// Validation result for transformation logic
struct ValidationResult {
    let isValid: Bool
    let errors: [ValidationError]
    let warnings: [ValidationWarning]

    struct ValidationError {
        let message: String
        let line: Int?
        let column: Int?
    }

    struct ValidationWarning {
        let message: String
        let suggestion: String?
    }

    static let valid = ValidationResult(
        isValid: true,
        errors: [],
        warnings: []
    )

    static func invalid(errors: [ValidationError]) -> ValidationResult {
        ValidationResult(
            isValid: false,
            errors: errors,
            warnings: []
        )
    }

    static func validWithWarnings(warnings: [ValidationWarning]) -> ValidationResult {
        ValidationResult(
            isValid: true,
            errors: [],
            warnings: warnings
        )
    }
}
