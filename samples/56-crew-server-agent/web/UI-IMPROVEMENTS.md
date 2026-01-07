# UI Improvements - Tool Validation Visibility

## Problem Fixed

**Issue**: After sending a message that triggers a tool call, the validation notification card would appear in the chat flow. As the streaming response continued, the notification would scroll out of view, making it impossible to access the Validate/Cancel buttons.

**Symptom**:
- User sends: "Say hello to Alice"
- Notification appears in chat
- Chat continues to scroll as messages arrive
- Notification is pushed off screen
- User cannot validate the tool call

## Solution

### Moved Notifications to Dedicated Overlay

The operation notifications are now displayed in a dedicated overlay area positioned **between the chat container and the input bar**.

### Layout Structure

```
┌─────────────────────────────┐
│ Header (Status Bar)         │
├─────────────────────────────┤
│                             │
│ Chat Container              │
│ (scrollable)                │
│                             │
│ - Messages                  │
│ - Responses                 │
│                             │
├─────────────────────────────┤
│ ⏳ Operations Overlay       │ ← ALWAYS VISIBLE
│                             │
│ ┌─────────────────────────┐ │
│ │ 🔔 Tool Call Validation │ │
│ │ [Validate] [Cancel]     │ │
│ └─────────────────────────┘ │
├─────────────────────────────┤
│ Input Bar                   │
│ [Type message...]  [Send]   │
└─────────────────────────────┘
```

### Key Features

#### 1. **Always Visible**
- Fixed position above input bar
- Doesn't scroll with chat content
- Always accessible regardless of chat length

#### 2. **Non-Intrusive**
- Hidden when empty (`:empty` CSS rule)
- Semi-transparent background
- Subtle shadow for depth

#### 3. **Scrollable if Needed**
- Max height: 40% of viewport
- Multiple notifications stack vertically
- Internal scroll if many operations

#### 4. **Visual Hierarchy**
```css
.operations-overlay {
    background: rgba(26, 26, 26, 0.98);
    border-top: 1px solid #404040;
    max-height: 40vh;
    overflow-y: auto;
}
```

## Files Modified

### 1. [web/js/app.js](web/js/app.js#L341-L349)

**Before** (in chat-container):
```vue
<div class="chat-container">
    <chat-message ... />
    <operation-controls ... />  <!-- Mixed with messages -->
</div>
```

**After** (separate overlay):
```vue
<div class="chat-container">
    <chat-message ... />
</div>

<div class="operations-overlay">
    <operation-controls ... />  <!-- Separate, always visible -->
</div>
```

### 2. [web/index.html](web/index.html#L234-L253)

Added CSS:
```css
.operations-overlay {
    background: rgba(26, 26, 26, 0.98);
    border-top: 1px solid #404040;
    padding: 1rem 1.5rem;
    flex-shrink: 0;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    max-height: 40vh;
    overflow-y: auto;
}

.operations-overlay:empty {
    display: none;  /* Hidden when no operations */
}

.operations-overlay .operation-controls {
    margin: 0;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
}
```

## User Experience

### Before
```
User: "Say hello to Alice"
  ↓
Notification appears in chat ✓
  ↓
Chat scrolls with new messages ↓↓↓
  ↓
Notification scrolls out of view ✗
  ↓
User can't click buttons ✗
```

### After
```
User: "Say hello to Alice"
  ↓
Notification appears in overlay ✓
  ↓
Chat scrolls with new messages ↓↓↓
  ↓
Notification stays visible ✓
  ↓
User clicks Validate ✓
```

## Testing

1. **Open** http://localhost:3000
2. **Send**: "Say hello to Alice"
3. **Observe**:
   - Notification appears in dedicated area above input
   - Chat continues scrolling
   - Buttons remain accessible
4. **Click** "Validate"
5. **Verify**:
   - Card turns green
   - Disappears after 3 seconds
   - Input remains usable

## Edge Cases Handled

### Multiple Operations
If multiple tools are triggered:
```
┌─────────────────────────────┐
│ Operations Overlay          │
│ ┌─────────────────────────┐ │
│ │ ⏳ Tool 1: say_hello    │ │
│ │ [Validate] [Cancel]     │ │
│ └─────────────────────────┘ │
│ ┌─────────────────────────┐ │
│ │ ⏳ Tool 2: calculate    │ │
│ │ [Validate] [Cancel]     │ │
│ └─────────────────────────┘ │
└─────────────────────────────┘
```

### No Operations
When no operations pending:
- Overlay is completely hidden (`:empty`)
- No visual clutter
- Input bar appears normally

### Long Messages
If chat gets very long:
- Chat container scrolls independently
- Operations overlay stays fixed
- Scrollbar only on chat, not overlay

## Benefits

✅ **Accessibility**: Buttons always reachable
✅ **Usability**: No hunting for notifications
✅ **Clarity**: Separate from chat flow
✅ **Responsive**: Works on different screen sizes
✅ **Clean**: Hidden when not needed

## Future Enhancements (Optional)

1. **Slide Animation**: Animate overlay appearance
2. **Sound Alert**: Play sound when operation needs attention
3. **Badge Count**: Show count in input bar when collapsed
4. **Keyboard Shortcuts**:
   - `V` to validate
   - `C` to cancel
5. **Auto-scroll to Chat Bottom**: When operation appears, scroll chat to show context
