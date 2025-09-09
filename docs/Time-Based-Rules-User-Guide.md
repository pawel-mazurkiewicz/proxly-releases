# Time-Based Rules User Guide

## What are Time-Based Rules?

Time-based rules automatically open URLs in specific browsers based on the time of day and day of the week. This helps you maintain work-life balance by automatically switching between different browser profiles or applications during different parts of your day.

## Common Use Cases

### Work Hours Separation
Automatically use your work browser profile during business hours (9 AM - 5 PM, Monday-Friday) and switch to your personal browser in the evenings and weekends.

### Focus Time
Route all browsing to a distraction-free browser during your designated focus hours.

### Weekend Relaxation
Use a different browser setup for weekend browsing with your favorite bookmarks and extensions.

### Night Mode
Switch to a browser with dark themes and reading-focused extensions during evening hours.

## How to Create a Time-Based Rule

### Step 1: Add a New Rule
1. Open Proxly settings
2. Go to the Rules section
3. Click "Add Rule"
4. Select "Time" as the rule type

*[Screenshot: Rules list with "Add Rule" button highlighted]*

### Step 2: Set Your Time Range
Use the time picker to select when this rule should be active:

- **Start Time**: When the rule begins (e.g., 9:00 AM)
- **End Time**: When the rule ends (e.g., 5:00 PM)

*[Screenshot: Time range picker showing start and end time selectors]*

**Tip**: You can create overnight rules (like 10:00 PM to 6:00 AM) for late-night browsing preferences.

### Step 3: Choose Active Days
Select which days of the week this rule applies:

- Click individual day buttons to toggle them on/off
- Use quick selection buttons for common patterns:
  - **Weekdays**: Monday through Friday
  - **Weekends**: Saturday and Sunday
  - **Every Day**: All seven days

*[Screenshot: Day selection interface with weekday buttons and quick selection options]*

### Step 4: Select Your Browser
Choose which browser or browser profile to use when this rule is active:

- Pick from your installed browsers
- Select specific browser profiles if available
- Configure custom launch arguments if needed

*[Screenshot: Browser selection dropdown with various browser options]*

### Step 5: Set Priority and Save
- Give your rule a descriptive name
- Set the priority (higher numbers = higher priority)
- Enable the rule
- Click "Save"

*[Screenshot: Rule configuration summary with name, priority, and save button]*

## Understanding Rule Priority

When multiple time-based rules could apply at the same time, Proxly uses priority to decide which one to use:

- **Higher priority numbers win** (Priority 10 beats Priority 5)
- If priorities are equal, the more specific rule is used
- Domain-specific rules always override time-based rules

*[Screenshot: Rules list showing different priorities]*

## Example Configurations

### Work Schedule (9-5, Weekdays)
- **Time**: 9:00 AM to 5:00 PM
- **Days**: Monday, Tuesday, Wednesday, Thursday, Friday
- **Browser**: Chrome with Work Profile
- **Use Case**: All work-related browsing during business hours

*[Screenshot: Work schedule rule configuration]*

### Evening Personal Time
- **Time**: 6:00 PM to 11:00 PM  
- **Days**: Every day
- **Browser**: Firefox with Personal Profile
- **Use Case**: Personal browsing after work

*[Screenshot: Evening rule configuration]*

### Weekend Relaxation
- **Time**: 10:00 AM to 10:00 PM
- **Days**: Saturday, Sunday
- **Browser**: Safari
- **Use Case**: Casual weekend browsing

*[Screenshot: Weekend rule configuration]*

### Late Night Focus
- **Time**: 11:00 PM to 6:00 AM
- **Days**: Every day
- **Browser**: Brave with ad blocking
- **Use Case**: Distraction-free late night browsing

*[Screenshot: Overnight rule configuration]*

## Managing Your Time-Based Rules

### Viewing Active Rules
The rules list shows which time-based rules are currently active with a visual indicator.

*[Screenshot: Rules list with active rule highlighted]*

### Editing Rules
1. Click on any rule to edit it
2. Modify the time range, days, or browser
3. Save your changes

### Temporarily Disabling Rules
Toggle the enable/disable switch to temporarily turn off a rule without deleting it.

*[Screenshot: Rule with enable/disable toggle]*

### Rule Conflicts
If you have overlapping time-based rules, the one with higher priority will be used. Proxly will show a warning if rules conflict.

*[Screenshot: Warning message about conflicting rules]*

## Tips for Effective Time-Based Rules

### Start Simple
Begin with basic work/personal separation before creating complex schedules.

### Use Descriptive Names
Name your rules clearly (e.g., "Work Hours - Chrome", "Weekend - Safari") so you can easily identify them.

### Set Appropriate Priorities
- Work rules: High priority (8-10)
- Personal rules: Medium priority (5-7)  
- Experimental rules: Low priority (1-4)

### Test Your Rules
After creating a rule, test it by opening a URL during the specified time to make sure it works as expected.

### Regular Review
Periodically review your time-based rules to ensure they still match your schedule and needs.

## Troubleshooting

### Rule Not Working?
1. Check that the rule is enabled
2. Verify the current time falls within your specified range
3. Confirm today is one of your selected days
4. Make sure no higher-priority domain rules are overriding it

### Wrong Browser Opening?
1. Check if a domain-specific rule is taking precedence
2. Verify your browser selection in the rule settings
3. Ensure the selected browser is still installed

### Overnight Rules Issues?
Make sure your end time is earlier than your start time (e.g., 22:00 to 06:00) for rules that span midnight.

*[Screenshot: Troubleshooting checklist or help panel]*

## Advanced Features

### Multiple Time Blocks
Create separate rules for different parts of the day:
- Morning routine (7-9 AM)
- Work hours (9 AM-5 PM)  
- Evening personal time (6-11 PM)

### Seasonal Adjustments
Temporarily disable or modify rules during vacations or schedule changes.

### Integration with Focus Modes
Time-based rules work alongside macOS Focus modes for even more automated browsing control.

---

*Need more help? Check out our other guides on Domain Rules and Browser Profile Management.*