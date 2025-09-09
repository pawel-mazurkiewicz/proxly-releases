# Time-Based Rules Matching in Proxly

## Overview

Time-based rules in Proxly allow users to automatically route URLs to specific browsers based on the current time and day of the week. This feature is particularly useful for creating work/personal browsing boundaries or implementing focus-time workflows.

## Rule Structure

Time-based rules are stored with the following components:

- **Type**: `RuleType.time`
- **Condition**: A JSON string or legacy time range format
- **Browser Target**: The destination browser configuration
- **Priority**: Integer value for rule precedence
- **Enabled Status**: Boolean flag to activate/deactivate the rule

## Condition Formats

### Modern JSON Format (Recommended)

```json
{
  "time": "09:00-17:00",
  "days": [2, 3, 4, 5, 6]
}
```

**Components:**
- `time`: Time range in 24-hour format (`HH:mm-HH:mm`)
- `days`: Array of weekday numbers (1=Sunday, 2=Monday, ..., 7=Saturday)

### Legacy Format (Backward Compatibility)

```
09:00-17:00
```

**Note**: Legacy format assumes all days of the week are active.

## Matching Algorithm

The time-based rule matching process follows these steps:

### 1. Rule Filtering

```swift
let timeRules = rules.filter { $0.type == .time && $0.isEnabled }
    .sorted { $0.priority > $1.priority }
```

- Only enabled time-based rules are considered
- Rules are sorted by priority (highest first)

### 2. Time and Day Validation

For each time rule, the system validates:

#### Day Matching
```swift
let calendar = Calendar.current
let currentWeekday = calendar.component(.weekday, from: Date())
guard timeCondition.days.contains(currentWeekday) else { return false }
```

- Gets current weekday (1-7, where 1=Sunday)
- Checks if current day is in the rule's allowed days array
- Returns false immediately if day doesn't match

#### Time Range Matching
```swift
private static func isCurrentTime(in timeRange: String) -> Bool {
    let components = timeRange.split(separator: "-").map(String.init)
    // Parse start and end times
    // Compare with current time
}
```

**Time Parsing:**
- Splits time range on `-` character
- Uses `DateFormatter` with `"HH:mm"` format
- Creates date objects for comparison

**Time Comparison Logic:**

1. **Standard Range** (e.g., `09:00-17:00`):
   ```swift
   return currentTime >= startDate && currentTime < endDate
   ```

2. **Overnight Range** (e.g., `22:00-06:00`):
   ```swift
   if startDate > endDate {
       return currentTime >= startDate || currentTime < endDate
   }
   ```

### 3. Rule Priority

When multiple time-based rules match:
- Rules are evaluated in priority order (highest first)
- First matching rule is selected and returned
- Lower priority rules are ignored

## Processing Flow

```
URL Request
    ↓
Domain Rules Check (Priority 1)
    ↓ (if no match)
Time Rules Check (Priority 2)
    ↓
For each time rule (by priority):
    ↓
Check if current day in allowed days
    ↓ (if yes)
Check if current time in time range
    ↓ (if yes)
Return matched rule
    ↓ (if no matches)
Focus Mode Rules Check (Priority 3)
    ↓ (if no match)
Apply fallback behavior
```

## Examples

### Weekday Work Hours
```json
{
  "time": "09:00-17:00",
  "days": [2, 3, 4, 5, 6]
}
```
- **Active**: Monday-Friday, 9 AM to 5 PM
- **Use Case**: Route work-related browsing to a dedicated browser profile

### Evening Personal Time
```json
{
  "time": "18:00-23:00",
  "days": [1, 2, 3, 4, 5, 6, 7]
}
```
- **Active**: Every day, 6 PM to 11 PM
- **Use Case**: Switch to personal browser for evening browsing

### Weekend Relaxation
```json
{
  "time": "10:00-22:00",
  "days": [1, 7]
}
```
- **Active**: Saturday and Sunday, 10 AM to 10 PM
- **Use Case**: Use a specific browser profile for weekend activities

### Overnight Range
```json
{
  "time": "22:00-06:00",
  "days": [1, 2, 3, 4, 5, 6, 7]
}
```
- **Active**: Every day, 10 PM to 6 AM (next day)
- **Use Case**: Late night browsing with specific settings

## Day Number Reference

| Day | Number | Constant |
|-----|--------|----------|
| Sunday | 1 | Calendar.current.component(.weekday, from: Date()) |
| Monday | 2 | |
| Tuesday | 3 | |
| Wednesday | 4 | |
| Thursday | 5 | |
| Friday | 6 | |
| Saturday | 7 | |

## Common Patterns

### Weekdays Only
```json
"days": [2, 3, 4, 5, 6]
```

### Weekends Only
```json
"days": [1, 7]
```

### Every Day
```json
"days": [1, 2, 3, 4, 5, 6, 7]
```

### Specific Days
```json
"days": [2, 4, 6]  // Monday, Wednesday, Friday
```

## Validation Rules

### Time Format
- Must be in `HH:mm-HH:mm` format
- Hours: 00-23
- Minutes: 00-59
- Start and end times cannot be identical

### Day Selection
- At least one day must be selected
- Day numbers must be between 1-7
- Duplicate days are automatically handled by using `Set<Int>`

### JSON Structure
- Must be valid JSON format
- Required fields: `time`, `days`
- `time` must be a string
- `days` must be an array of integers

## Error Handling

### Invalid Time Format
```swift
AppLogger.log("Invalid time range format: \(timeRange)", level: .error)
return false
```

### JSON Parsing Errors
```swift
AppLogger.log("Failed to decode time-based rule: \(condition) - Error: \(error.localizedDescription)", level: .error)
return false
```

### Missing Components
- Invalid time parsing results in rule being skipped
- Malformed JSON falls back to legacy format parsing
- Missing days array defaults to legacy behavior (all days)

## Performance Considerations

### Regex Caching
- Time-based rules don't use regex patterns
- No caching overhead for time calculations

### Date Calculations
- Date components are extracted once per evaluation
- Calendar operations are optimized by the system
- Time comparisons use efficient date arithmetic

### Rule Sorting
- Rules are sorted once when loaded
- Priority-based evaluation stops at first match
- No unnecessary rule evaluations after match found

## Integration with Other Rule Types

Time-based rules have **Priority 2** in the overall rule evaluation hierarchy:

1. **Domain Rules** (Priority 1) - Always checked first
2. **Time Rules** (Priority 2) - Checked if no domain match
3. **Focus Mode Rules** (Priority 3) - Checked if no time match

This ensures that specific domain rules can override time-based routing when needed.

## Migration and Backward Compatibility

The system supports both legacy and modern formats:

### Legacy Support
```swift
// Handle legacy format "HH:mm-HH:mm" (backwards compatibility)
// Assume all days are allowed for legacy rules
return isCurrentTime(in: condition)
```

### Modern Format Detection
```swift
if condition.starts(with: "{") {
    // Parse JSON format
} else {
    // Handle legacy format
}
```

This ensures existing time-based rules continue to work while new rules benefit from enhanced day-specific functionality.