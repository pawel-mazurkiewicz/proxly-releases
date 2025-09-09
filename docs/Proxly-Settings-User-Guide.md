# Proxly Settings User Guide

## Overview

Proxly's settings allow you to customize how the app behaves, manage browser profiles, control notifications, and configure automatic updates. This guide covers all available settings and how to use them effectively.

*[Screenshot: Settings window overview showing all sections]*

## General Settings

### Enable Rules
**What it does**: Master switch that turns Proxly's URL routing on or off.

- **Enabled**: Proxly will process URLs according to your rules
- **Disabled**: All URLs open with your fallback browser (Proxly is essentially bypassed)

**When to use**: 
- Temporarily disable when troubleshooting browser issues
- Turn off during presentations to avoid unexpected browser switching

*[Screenshot: Enable Rules toggle in the General section]*

### Launch at Login
**What it does**: Automatically starts Proxly when you log into your Mac.

**Recommended**: Keep this enabled so Proxly is always ready to handle URLs.

**Note**: Proxly runs quietly in the background and uses minimal system resources.

*[Screenshot: Launch at Login toggle]*

### Show Tooltips
**What it does**: Displays helpful hints when you hover over settings and buttons.

**Recommended for new users**: Keep enabled while learning Proxly's features, then disable if you prefer a cleaner interface.

*[Screenshot: Tooltip example appearing over a setting]*

## Profile Features

### Enable Profile Features
**What it does**: Allows Proxly to open URLs in specific browser profiles (like Chrome's work profile vs personal profile).

**Requirements**: 
- Supported browsers (Chrome, Edge, Brave, Firefox)
- Browser profiles must be set up in the browser first

**When enabled**: You'll see profile options when creating rules and can manage profile permissions.

*[Screenshot: Profile Features section with toggle enabled]*

### Manage Profile Permissions
**What it does**: Opens a dedicated window to grant Proxly access to read browser profiles.

**Why needed**: macOS security requires explicit permission to access browser profile information.

**Steps**:
1. Click "Manage Permissions"
2. Follow the prompts to grant folder access
3. Select your browser's profile directory when prompted

*[Screenshot: Profile permissions management window]*

## Browser Visibility

### Manage Browser Visibility
**What it does**: Controls which browsers appear in Proxly's browser selection menus.

**Use cases**:
- Hide browsers you never use to reduce clutter
- Show only work-approved browsers in a corporate environment
- Temporarily hide browsers that are having issues

**How to use**:
1. Click "Manage Browser Visibility"
2. Toggle browsers on/off as needed
3. Hidden browsers won't appear in rule creation or selection panels

*[Screenshot: Browser visibility management interface with browser list and toggles]*

## Default Behavior

### When No Rule Matches
**What it does**: Determines what happens when a URL doesn't match any of your rules.

**Options**:
- **Show Selection Panel**: Opens a popup letting you choose which browser to use
- **Open in Fallback Browser**: Automatically opens in your designated fallback browser

**Recommendation**: 
- New users: Start with "Show Selection Panel" to learn which URLs need rules
- Experienced users: Use "Open in Fallback Browser" for seamless operation

*[Screenshot: Default behavior picker showing both options]*

### Fallback Browser
**What it does**: The browser used when no rules match and you've chosen "Open in Fallback Browser".

**How to choose**:
- Select your most commonly used browser
- Consider using your system's default browser
- Click the refresh button if you've recently installed new browsers

*[Screenshot: Fallback browser dropdown menu with browser icons and names]*

### Notify on Fallback
**What it does**: Shows a notification when URLs are opened with the fallback browser.

**Useful for**:
- Identifying URLs that need rules
- Monitoring when rules aren't working as expected
- Debugging rule configurations

**Note**: Requires notification permissions. Proxly will prompt you to enable these in System Preferences.

*[Screenshot: Notification permission dialog and example fallback notification]*

## Logging

### Enable Logging
**What it does**: Records Proxly's activity for troubleshooting and debugging.

**What gets logged**:
- Which rules matched for specific URLs
- Browser launch successes and failures
- Profile detection and access attempts
- Error messages and warnings

**When to enable**:
- Troubleshooting rule matching issues
- Debugging browser profile problems
- When reporting bugs to support

**Privacy note**: Logs are stored locally on your Mac and contain URLs you visit.

*[Screenshot: Logging toggle and explanation of what gets logged]*

## Updates

### Automatic Update Checks
**What it does**: Periodically checks for new versions of Proxly.

**Benefits**:
- Get new features and bug fixes automatically
- Receive security updates promptly
- Stay compatible with new browser versions

*[Screenshot: Automatic updates toggle]*

### Update Check Frequency
**What it does**: Controls how often Proxly checks for updates.

**Options**:
- **Daily**: Checks every day (recommended for beta users)
- **Weekly**: Checks once per week (recommended for most users)
- **Monthly**: Checks once per month (for stable environments)

*[Screenshot: Update frequency dropdown menu]*

### Current Version & Manual Check
**What it does**: Shows your current Proxly version and allows manual update checks.

**Use the manual check when**:
- You want to check for updates immediately
- Troubleshooting update-related issues
- Verifying you have the latest version

*[Screenshot: Version information and "Check Now" button]*

## Advanced Tips

### Optimal Settings for Different Users

#### New Users
- Enable Rules: ✓
- Launch at Login: ✓
- Show Tooltips: ✓
- Default Behavior: Show Selection Panel
- Notify on Fallback: ✓
- Logging: ✓ (temporarily, for learning)

#### Power Users
- Enable Rules: ✓
- Launch at Login: ✓
- Show Tooltips: ✗
- Default Behavior: Open in Fallback Browser
- Notify on Fallback: ✗
- Logging: ✗ (unless troubleshooting)

#### Corporate/Team Environments
- Profile Features: ✓
- Browser Visibility: Managed (hide personal browsers)
- Automatic Updates: ✓ (weekly)
- Logging: ✓ (for support)

### Troubleshooting Common Issues

#### Rules Not Working
1. Check that "Enable Rules" is turned on
2. Enable logging to see what's happening
3. Verify your fallback browser is set correctly
4. Check rule priorities and conditions

#### Browser Profiles Not Detected
1. Ensure "Profile Features" is enabled
2. Click "Manage Profile Permissions"
3. Grant folder access when prompted
4. Restart Proxly after granting permissions

#### Missing Browsers in Lists
1. Click the refresh button next to browser selections
2. Check "Manage Browser Visibility" settings
3. Ensure browsers are properly installed
4. Restart Proxly if browsers still don't appear

#### Notifications Not Working
1. Check that "Notify on Fallback" is enabled
2. Grant notification permissions in System Preferences
3. Test with a URL that should trigger fallback behavior

### Performance Considerations

#### Memory Usage
- Proxly uses minimal memory when running in the background
- Logging increases memory usage slightly
- Profile detection requires periodic disk access

#### Battery Impact
- Negligible battery impact during normal operation
- Update checks use minimal network data
- No continuous background processing

#### Startup Time
- Launch at login adds ~1-2 seconds to login time
- Profile scanning may add slight delay on first launch
- Subsequent launches are nearly instantaneous

## Privacy and Security

### Data Collection
- Proxly does not send usage data to external servers
- All settings and rules are stored locally on your Mac
- Logs contain URLs you visit but remain on your device

### Permissions Required
- **Folder Access**: To read browser profiles (optional)
- **Notifications**: To show fallback notifications (optional)
- **Launch at Login**: To start automatically (optional)

### Network Access
- Only used for checking updates
- No analytics or tracking
- No personal data transmitted

---

*For more help with specific features, see our guides on Creating Rules, Managing Browser Profiles, and Troubleshooting Common Issues.*