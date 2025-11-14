# Quickstart: Testing Browser & Profile Ordering

**Feature**: 002-browser-order-shortcuts
**Created**: 2025-11-03
**Audience**: Developers implementing or testing the browser ordering feature

## Prerequisites

- Xcode 16+ installed
- macOS 14.0+ target device
- Multiple browsers installed (Chrome, Safari, Firefox, Brave, etc.)
- At least one browser with multiple profiles configured (e.g., Chrome with "Work" and "Personal" profiles)

## Quick Setup

### 1. Branch Setup

```bash
cd /Users/pawelma/code/browser-chooser/proxly
git checkout 002-browser-order-shortcuts
git pull origin 002-browser-order-shortcuts
```

### 2. Build and Run

```bash
# Open project in Xcode
open Proxly.xcodeproj

# Select "Proxly" scheme
# Press Cmd+R to build and run
```

### 3. Enable Profile Features (for P2 testing)

```bash
# In Proxly app:
# Settings → Advanced → Enable Profile Features toggle
```

---

## Testing Scenarios

### Scenario 1: Basic Browser Reordering (P1)

**Objective**: Verify drag-and-drop reordering works and persists

**Steps**:
1. Open Proxly app
2. Navigate to Settings → Browser Visibility
3. Click "Manage Browsers" button
4. **Expected**: See list of installed browsers in alphabetical order (initial state)
5. Drag Safari browser card to the top position (above all other browsers)
6. **Expected**: Safari moves to position 1, other browsers shift down
7. Close Manage Browsers popup
8. Trigger a URL (e.g., click a link in Mail.app)
9. When browser selection panel appears, press keyboard key "1"
10. **Expected**: Safari opens (not the original first browser)

**Verification**:
- Order change reflects immediately in UI
- Keyboard shortcut "1" now maps to Safari
- Order persists after app restart (quit Proxly, relaunch, verify order unchanged)

---

### Scenario 2: Order Persistence Across Refresh

**Objective**: Verify custom order survives browser list refresh

**Steps**:
1. Set custom browser order: Firefox (1), Chrome (2), Safari (3), Brave (4)
2. Verify order in browser selection panel (trigger URL, check keyboard shortcuts)
3. Go to Settings → Browser Visibility → click "Refresh Browser List"
4. **Expected**: Order remains Firefox, Chrome, Safari, Brave (no change)
5. Trigger URL again, verify keyboard shortcuts still match custom order

**Verification**:
- Refresh does not reset order to alphabetical
- No duplicate entries appear
- Newly detected browsers (if any) append to end

---

### Scenario 3: New Browser Detection

**Objective**: Verify newly installed browsers append to end of custom order

**Setup**:
1. Uninstall Brave Browser (if installed): `rm -rf /Applications/Brave\ Browser.app`
2. Set custom order: Safari (1), Chrome (2), Firefox (3)
3. Install Brave Browser: Download from https://brave.com/download/
4. In Proxly, click "Refresh Browser List"

**Expected Behavior**:
- Brave appears at position 4 (after Firefox)
- Existing order unchanged: Safari (1), Chrome (2), Firefox (3), Brave (4)
- Keyboard shortcut "4" now maps to Brave

---

### Scenario 4: Browser Uninstall Handling

**Objective**: Verify custom order gracefully handles browser removal

**Setup**:
1. Set custom order: Safari (1), Chrome (2), Firefox (3), Brave (4)
2. Uninstall Firefox: `rm -rf /Applications/Firefox.app`
3. Restart Proxly app (or trigger browser detection)

**Expected Behavior**:
- Firefox removed from order
- Remaining browsers renumbered: Safari (1), Chrome (2), Brave (3)
- Keyboard shortcuts adjust automatically
- No error messages or crashes

---

### Scenario 5: Profile-Level Ordering (P2)

**Objective**: Verify profile reordering and interleaving works

**Prerequisites**:
- Chrome installed with at least 2 profiles ("Work" and "Personal")
- Profile Features enabled in Settings

**Steps**:
1. Open Manage Browsers popup
2. Expand Chrome browser entry (click chevron icon)
3. **Expected**: See "Work" and "Personal" profiles listed under Chrome
4. Drag profiles to create order: Chrome Work (1), Safari (2), Chrome Personal (3), Firefox (4)
5. Close popup
6. Trigger URL with browser selection panel
7. Press key "1": **Expected** Chrome opens with Work profile
8. Trigger URL again, press key "2": **Expected** Safari opens
9. Trigger URL again, press key "3": **Expected** Chrome opens with Personal profile

**Verification**:
- Profiles can be interleaved with browsers (not grouped under browser)
- Keyboard shortcuts respect profile-level order
- Profile selection reflects custom order

---

### Scenario 6: Visibility Integration

**Objective**: Verify custom order respects browser visibility settings

**Steps**:
1. Set custom order: Safari (1), Chrome (2), Firefox (3), Brave (4)
2. Hide Chrome: Settings → Browser Visibility → toggle off Chrome
3. Trigger URL with browser selection panel
4. **Expected**: Only Safari (1), Firefox (2), Brave (3) visible
5. Keyboard shortcuts: "1" = Safari, "2" = Firefox, "3" = Brave (skip hidden Chrome)
6. Open Manage Browsers popup
7. **Expected**: All browsers shown in order including hidden Chrome (dimmed/grayed out)

**Verification**:
- Hidden browsers excluded from selection panel
- Hidden browsers retained in custom order (can be reordered)
- Keyboard shortcuts skip hidden browsers (sequential numbering of visible only)

---

### Scenario 7: Reset to Default

**Objective**: Verify reset button restores alphabetical order

**Steps**:
1. Set custom order: Firefox (1), Safari (2), Chrome (3)
2. Open Manage Browsers popup
3. Click "Reset Order" button
4. **Expected**: Order resets to alphabetical (Brave, Chrome, Firefox, Safari)
5. Close popup
6. Verify keyboard shortcuts now match alphabetical order

**Verification**:
- Reset button appears in Manage Browsers UI
- Reset restores alphabetical order by display name
- Reset persists (survives app restart)

---

## Manual Testing Checklist

Use this checklist for regression testing before releases:

### P1: Browser Ordering
- [ ] Drag-and-drop reordering works smoothly
- [ ] Visual drag feedback (preview, drop indicators) appears
- [ ] Order persists across app restarts
- [ ] Order survives "Refresh Browser List" operation
- [ ] Newly detected browsers append to end
- [ ] Uninstalled browsers removed from order
- [ ] Keyboard shortcuts 1-9 map to custom order
- [ ] Browser selection panel respects custom order
- [ ] RuleEditView browser picker respects custom order
- [ ] Reset to default button works

### P2: Profile Ordering
- [ ] Profile expansion/collapse works (if profiles exist)
- [ ] Profile drag-and-drop reordering works
- [ ] Profiles can be interleaved with browsers
- [ ] Keyboard shortcuts map to browser+profile combinations
- [ ] Profile selection sheet respects profile order
- [ ] Profiles toggle off when Profile Features disabled

### Integration
- [ ] Hidden browsers excluded from selection panel
- [ ] Hidden browsers retained in custom order
- [ ] Keyboard shortcuts skip hidden browsers
- [ ] VoiceOver announces position numbers correctly
- [ ] Keyboard-based reordering works (accessibility feature)

### Edge Cases
- [ ] Order with 1 browser only (no reordering possible)
- [ ] Order with 10+ browsers (only first 9 get keyboard shortcuts)
- [ ] Browser with no profiles (profiles section not shown)
- [ ] All browsers hidden except one (order still applies)
- [ ] Empty browser list (rare, but should not crash)

---

## Simulating Edge Cases

### Browser Install/Uninstall

**Install browser temporarily**:
```bash
# Download browser .dmg
# Drag to /Applications
# In Proxly: Refresh Browser List
```

**Uninstall browser safely**:
```bash
# Move to trash (easily recoverable)
mv "/Applications/Brave Browser.app" ~/.Trash/

# In Proxly: Restart app or refresh browser list
```

**Restore browser**:
```bash
mv ~/.Trash/Brave\ Browser.app /Applications/
```

---

### Profile Creation/Deletion

**Chrome profiles**:
1. Open Chrome
2. Click profile icon → Add → Create new profile
3. Name it (e.g., "Test Profile")
4. In Proxly: Open Manage Browsers, expand Chrome, verify "Test Profile" appears

**Delete Chrome profile**:
1. Open Chrome
2. Click profile icon → Manage profiles → ⋮ (three dots) → Delete

**Firefox profiles**:
1. Open Firefox
2. Type `about:profiles` in address bar
3. Click "Create a New Profile"

---

### Stress Testing

**Test with many browsers** (10+ installed):
- Install: Chrome, Safari, Firefox, Brave, Edge, Opera, Vivaldi, Arc, Sidekick, etc.
- Verify: Only first 9 get keyboard shortcuts, rest shown in order but no shortcuts

**Test with many profiles** (5+ profiles in one browser):
- Create 5 Chrome profiles
- Reorder them extensively
- Verify: All profiles appear in custom order, keyboard shortcuts work correctly

---

## Debugging Tips

### View Persisted Order Data

```bash
# Dump UserDefaults for browserOrder key
defaults read com.pawelmaz.Proxly browserOrder

# Expected output: JSON string with orderedEntries array
```

### Enable Logging

```swift
// In BrowserOrderService or PersistenceManager
AppLogger.log("Browser order loaded: \(state.orderedEntries.count) entries", level: .debug)
AppLogger.log("Browser order saved: \(state)", level: .debug)
```

**View logs**:
```bash
# Console.app → filter for "Proxly" or "browser order"
# Or: tail -f ~/Library/Logs/Proxly/app.log  (if file logging enabled)
```

---

### Reset State Manually

If testing gets into a bad state, reset UserDefaults:

```bash
# Delete browser order data
defaults delete com.pawelmaz.Proxly browserOrder

# Restart Proxly app
# Order will reinitialize to alphabetical on next launch
```

---

### Xcode Debugging

**Breakpoints**:
- `BrowserOrderService.updateOrder(_:)` - Track when order changes
- `PersistenceManager.saveBrowserOrder(_:)` - Verify persistence
- `BrowserOrderService.loadOrder()` - Verify loading on startup

**LLDB Commands**:
```lldb
# Print current order state
po BrowserOrderService.shared.currentOrder

# Print persisted data
po UserDefaults.standard.data(forKey: "browserOrder")
```

---

## Automated Testing

### Unit Tests

Run all tests:
```bash
xcodebuild test -project Proxly.xcodeproj -scheme Proxly -destination 'platform=macOS'
```

Run specific test suite:
```bash
xcodebuild test -project Proxly.xcodeproj \
  -scheme Proxly \
  -destination 'platform=macOS' \
  -only-testing:ProxlyTests/BrowserOrderServiceTests
```

### Key Test Files

**Unit Tests**:
- `ProxlyTests/BrowserOrderStateTests.swift` - Data model validation
- `ProxlyTests/BrowserOrderServiceTests.swift` - Service logic (merge, append, remove)
- `ProxlyTests/PersistenceManagerTests.swift` - Persistence roundtrip

**UI Tests** (if implemented):
- `ProxlyUITests/BrowserManagementViewTests.swift` - Drag-and-drop interaction
- `ProxlyUITests/SelectionPanelTests.swift` - Keyboard shortcut mapping

---

## Performance Testing

### Measure UI Responsiveness

**Objective**: Verify reordering updates UI within 100ms (SC-007)

**Manual test**:
1. Open Manage Browsers popup
2. Drag browser card to new position
3. Observe animation smoothness (should be instant, no lag)

**Profiling** (optional):
1. In Xcode: Product → Profile (Cmd+I)
2. Select "Time Profiler" instrument
3. Drag browsers around in Manage Browsers UI
4. **Expected**: `updateOrder(_:)` method < 5ms, UI updates < 100ms

---

### Measure Persistence Performance

**Objective**: Verify save/load operations < 5ms

**Approach**:
1. Add performance test in `BrowserOrderServiceTests.swift`:
```swift
func testOrderPersistencePerformance() {
    let orders: [[String]] = [
        (0..<50).map { _ in UUID().uuidString }  // 50 browsers (max)
    ]

    measure {
        for order in orders {
            BrowserOrderService.shared.updateOrder(order)
            let loaded = BrowserOrderService.shared.loadOrder()
            XCTAssertEqual(loaded.count, order.count)
        }
    }
}
```

2. Run performance test: Select test → Cmd+U → View baseline

---

## Accessibility Testing

### VoiceOver Testing

**Enable VoiceOver**: Cmd+F5

**Test scenarios**:
1. Navigate to Manage Browsers popup with VoiceOver enabled
2. Focus on a browser card
3. **Expected**: Hear "Position 1 of 4: Safari, swipe up or down to reorder"
4. Swipe up (VO+Up): **Expected** Safari moves up one position
5. Swipe down (VO+Down): **Expected** Safari moves down one position

**Verification**:
- Position announcements are accurate
- Reorder hints are clear
- Keyboard navigation works (Tab, Shift+Tab)
- Drag handles have accessible labels

---

### Keyboard Navigation Testing

**Test without mouse**:
1. Open Manage Browsers popup
2. Press Tab to focus first browser
3. Press Tab to cycle through browsers
4. Press Space to select/activate
5. Use arrow keys + modifier (e.g., Cmd+Up/Down) to reorder
6. **Expected**: Full functionality available without mouse

---

## Localization Testing

### Test All 7 Languages

**Supported languages**: English, German, Spanish, French, Polish, Dutch, Italian

**Steps**:
1. Change system language: System Settings → Language & Region
2. Restart Proxly app
3. Open Manage Browsers popup
4. **Verify**:
   - "Reset Order" button translated
   - Position indicators translated (e.g., "Position 1 von 4" in German)
   - Drag hints translated
   - No English strings visible (unless intentional fallback)

**Key strings to verify**:
- `browser_management_reset_order` - Reset button
- `browser_management_position_label` - Position indicator
- `browser_management_reorder_hint` - Drag hint

---

## Common Issues and Solutions

### Issue: Drag-and-drop not working

**Symptoms**: Can't drag browser cards, no visual feedback

**Possible causes**:
- Browser model doesn't conform to `Transferable`
- `.draggable()` modifier not applied
- `.dropDestination()` missing or misconfigured

**Solution**:
1. Verify `Browser` conforms to `Transferable`:
```swift
extension Browser: Transferable {
    static var transferRepresentation: some TransferRepresentation {
        CodableRepresentation(contentType: .json)
    }
}
```
2. Check `BrowserManagementView` for `.draggable(browser)` on each card

---

### Issue: Order not persisting

**Symptoms**: Custom order resets to alphabetical after app restart

**Possible causes**:
- `PersistenceManager.saveBrowserOrder()` not called
- UserDefaults key mismatch
- JSON encoding error

**Solution**:
1. Add logging to `saveBrowserOrder()`:
```swift
AppLogger.log("Saving browser order: \(order)", level: .debug)
```
2. Verify UserDefaults data:
```bash
defaults read com.pawelmaz.Proxly browserOrder
```
3. Check Console.app for encoding errors

---

### Issue: Keyboard shortcuts don't match order

**Symptoms**: Pressing "1" opens wrong browser

**Possible causes**:
- `SelectionPanel` not using `BrowserOrderService`
- Visibility filtering incorrect
- Profile features affecting order unexpectedly

**Solution**:
1. Add logging in `SelectionPanel`:
```swift
let orderedBrowsers = BrowserOrderService.shared.getOrderedBrowsers()
AppLogger.log("Selection panel order: \(orderedBrowsers.map { $0.displayName })", level: .debug)
```
2. Verify visibility filter applied correctly
3. Check if profile features enabled (affects order display)

---

## Additional Resources

- **Specification**: [spec.md](./spec.md) - Full feature requirements
- **Research**: [research.md](./research.md) - Technical decisions
- **Data Model**: [data-model.md](./data-model.md) - Entity definitions
- **Plan**: [plan.md](./plan.md) - Implementation phases
- **Constitution**: [../../.specify/memory/constitution.md](../../.specify/memory/constitution.md) - Architecture principles

---

## Feedback and Reporting

**Report issues**:
- Create GitHub issue: https://github.com/pawel-mazurkiewicz/proxly/issues
- Tag with: `feature:browser-ordering`, `bug` or `enhancement`
- Include: Proxly version, macOS version, browser versions, steps to reproduce

**Feature requests**:
- Add to spec as "Future Enhancement" section
- Discuss architectural impact before implementation
