# Native App URL Transformation – Task Tracker

Version: 2025-10-08  
Source Spec: `SPECS/NativeAppURLTransformation.md`

> Update this list as work progresses. Keep tasks atomic, tick status, and record dependencies/notes for future collaborators.

## Status Legend
- ⬜️ Not Started
- 🔄 In Progress
- ✅ Complete
- ⚠️ Blocked

## Core Engineering

| ID | Status | Area | Task | Owner | Notes / Dependencies |
| --- | --- | --- | --- | --- | --- |
| ENG-001 | ✅ | Models | Extend `Rule` with `nativeAppTransformation`, add Codable migrations |  | Initial model implementation landed 2025-10-09; requires schema validation against existing persisted rules |
| ENG-002 | ✅ | Models | Implement `NativeAppTransformation`, `TransformationPattern`, `PatternMatcher`, `TransformationLogic` types |  | Base enums/structs added 2025-10-09; align enums with spec tier definitions |
| ENG-003 | ✅ | Services | Create `NativeAppTransformationEngine` actor with public API (`transform`, `canTransform`, `refreshCaches`) |  | Actor scaffolding added 2025-10-09 with DI for `AppEnumerationServicing` |
| ENG-004 | ✅ | Services | Implement matching, caching, and validation logic (including rate limiting) |  | Matching, caching, validation, rate limiting implemented 2025-10-09; parameter map/template support requires further tuning via tests |
| ENG-005 | ✅ | Services | Add new methods to `AppEnumerationService` (`isSchemeRegistered`, `appSupportsScheme`, `applicationInfo`) and protocol |  | Async protocol + lookup caches implemented 2025-10-09; verify persistence rebuild on upgrade |
| ENG-006 | ✅ | Routing | Integrate engine call into `URLProcessor.processURLCore` after rule selection |  | Engine integration added 2025-10-09; fallback handling pending ENG-007 |
| ENG-007 | ✅ | Routing | Implement fallback behaviours (`useRuleBrowser`, `useDefaultBrowser`, `promptUser`) for skipped transformations |  | Fallback handler integrated in URLProcessor 2025-10-09; verify notifications/UX with UI work |
| ENG-008 | ✅ | Logging | Add structured logging for transformation attempts and outcomes |  | Attempt/outcome logs added across URLProcessor/engine 2025-10-09 |
| ENG-009 | ✅ | Presets | Build preset loader (read-only) for native app patterns from `docs/patterns/native-app-presets.json` |  | Loader + sample JSON added 2025-10-09; expand dataset via DOC-002 |
| ENG-010 | ✅ | Persistence | Verify rule persistence upgrades; add regression tests loading/saving rules with transformations |  | Rule/Persistence tests updated 2025-10-09 to cover transformations |

## User Interface

| ID | Status | Area | Task | Owner | Notes / Dependencies |
| --- | --- | --- | --- | --- | --- |
| UI-001 | ⬜️ | Rule Editor | Add “Native App” section with enable toggle, app picker, pattern summary |  | Section implemented 2025-10-09 with preset import and accessibility |
| UI-002 | ⬜️ | Rule Editor | Hook UI bindings to mutate `nativeAppTransformation` safely (no direct `@Published private(set)` writes) |  | Update helpers ensure safe mutations 2025-10-09 |
| UI-003 | ⬜️ | Rule Editor | Implement inline validation and error surfacing for pattern configuration |  | Inline validation + localized messaging added 2025-10-09 |
| UI-004 | ⬜️ | Tester | Add sheet/modal to test transformations against sample URLs |  | Engine must expose `transform` (ENG-003) |
| UI-005 | ⬜️ | Notifications | Update user notifications for fallback scenarios (app missing / failure) |  | Depends on ENG-007 |

## Documentation & Presets

| ID | Status | Area | Task | Owner | Notes / Dependencies |
| --- | --- | --- | --- | --- | --- |
| DOC-001 | ⬜️ | Docs | Update rule editor documentation with native app workflow in `docs/Rules-and-Processing-Engine-Guide.md` - document new feature thoroughly |  | After UI-001 |
| DOC-002 | ⬜️ | Docs | Author `docs/patterns/native-app-presets.json` covering Figma, Notion, Spotify, Linear, Discord, Trello, Slack, Zoom, VS Code, Obsidian |  | Source examples from appendix in spec |
| DOC-003 | ⬜️ | Help | Add troubleshooting section for missing native apps / invalid schemes |  | After ENG-007 |

## Testing

| ID | Status | Area | Task | Owner | Notes / Dependencies |
| --- | --- | --- | --- | --- | --- |
| TEST-001 | ⬜️ | Unit | Add `NativeAppTransformationEngineTests` covering all transformation tiers and validation |  | Depends on ENG-003/004 |
| TEST-002 | ⬜️ | Unit | Add persistence round-trip tests for rules with transformations |  | After ENG-010 |
| TEST-003 | ⬜️ | Integration | Exercise full URL processing pipeline with native transformations and fallbacks |  | Needs ENG-006/007 |
| TEST-004 | ⬜️ | UI | (Optional) SwiftUI snapshot or UI tests for rule editor changes |  | After UI tasks complete |

## Release Readiness

| ID | Status | Area | Task | Owner | Notes / Dependencies |
| --- | --- | --- | --- | --- | --- |
| REL-001 | ⬜️ | QA | Manual smoke checklist for each preset app (success + fallback) |  | After TEST-003 |
| REL-002 | ⬜️ | Metrics | Define logging review process to monitor transformation errors post-release |  | Coordinate with ENG-008 |
| REL-003 | ⬜️ | Launch | Update changelog / release notes highlighting native app support |  | Final step before ship |

## Backlog / Future Enhancements
- Community preset sharing and sync
- User-authored transformation editor
- Analytics for transformation usage
- Multi-native-app priorities per rule
