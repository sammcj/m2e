# Unit Conversion Troubleshooting Guide

This guide helps you troubleshoot common issues with M2E's unit conversion feature.

## Quick Diagnostics

### Check if Unit Conversion is Working

Test with a simple example:
```bash
echo "The room is 12 feet wide" | m2e-cli -units
# Expected output: "The room is 3.7 metres wide"
```

If this doesn't work, unit conversion may be disabled or misconfigured.

### Check Configuration File

1. **Check if config file exists:**
   ```bash
   ls -la ~/.config/m2e/unit_config.json
   ```

2. **Validate JSON syntax:**
   ```bash
   cat ~/.config/m2e/unit_config.json | python -m json.tool
   ```

3. **View current configuration:**
   ```bash
   cat ~/.config/m2e/unit_config.json
   ```

## Common Issues

### 1. Units Not Converting

**Symptoms:** Imperial units remain unchanged in output

**Possible Causes:**
- Unit conversion is disabled globally
- Specific unit type is disabled
- Unit doesn't meet confidence threshold
- Unit matches an exclusion pattern

**Solutions:**

1. **Check global enable flag:**
   ```json
   {
     "enabled": true  // Must be true
   }
   ```

2. **Check enabled unit types:**
   ```json
   {
     "enabledUnitTypes": ["length", "mass", "volume", "temperature", "area"]
   }
   ```

3. **Lower confidence threshold:**
   ```json
   {
     "detection": {
       "minConfidence": 0.3  // Lower from default 0.5
     }
   }
   ```

4. **Check exclusion patterns:**
   Remove or modify patterns that might be excluding your text:
   ```json
   {
     "excludePatterns": [
       // Remove patterns that might match your text
     ]
   }
   ```

### 2. Wrong Unit Conversions

**Symptoms:** Units convert but to wrong values or units

**Possible Causes:**
- Incorrect precision settings
- Wrong unit type detection
- Inappropriate metric unit selection

**Solutions:**

1. **Adjust precision:**
   ```json
   {
     "precision": {
       "length": 2,    // More decimal places
       "mass": 1,
       "temperature": 0  // Whole numbers only
     }
   }
   ```

2. **Check conversion preferences:**
   ```json
   {
     "preferences": {
       "preferWholeNumbers": false,  // Disable rounding
       "maxDecimalPlaces": 3,        // Allow more precision
       "roundingThreshold": 0.05     // Stricter rounding
     }
   }
   ```

### 3. Unwanted Conversions

**Symptoms:** Idiomatic expressions or non-measurement text gets converted

**Possible Causes:**
- Missing exclusion patterns
- Confidence threshold too low
- Detection settings too permissive

**Solutions:**

1. **Add exclusion patterns:**
   ```json
   {
     "excludePatterns": [
       "miles?\\s+(?:away|apart|from\\s+home|ahead)",
       "inch\\s+by\\s+inch",
       "every\\s+inch",
       "tons?\\s+of\\s+(?:fun|work|stuff)",
       "your_custom_pattern_here"
     ]
   }
   ```

2. **Increase confidence threshold:**
   ```json
   {
     "detection": {
       "minConfidence": 0.7  // Higher threshold
     }
   }
   ```

3. **Reduce detection distance:**
   ```json
   {
     "detection": {
       "maxNumberDistance": 2  // Require closer number-unit proximity
     }
   }
   ```

### 4. Configuration File Errors

**Symptoms:** Error messages about invalid configuration

**Common JSON Errors:**
- Missing commas between properties
- Trailing commas (not allowed in JSON)
- Unquoted property names
- Invalid escape sequences in regex patterns

**Solutions:**

1. **Validate JSON syntax:**
   ```bash
   cat ~/.config/m2e/unit_config.json | python -m json.tool
   ```

2. **Common fixes:**
   - Add missing commas: `"property1": value,` (not `"property1": value`)
   - Remove trailing commas: `"lastProperty": value` (not `"lastProperty": value,`)
   - Quote all property names: `"enabled"` (not `enabled`)
   - Escape backslashes in regex: `"miles?\\s+"` (not `"miles?\s+"`)

3. **Reset to defaults:**
   ```bash
   rm ~/.config/m2e/unit_config.json
   # Configuration will be recreated with defaults
   ```

### 5. Code-Aware Issues

**Symptoms:** Units in code get converted when they shouldn't, or comments don't get converted

**Possible Causes:**
- File type not detected correctly
- Comment detection issues
- Code block detection problems

**Solutions:**

1. **For code files:** Only comments should be converted
   ```go
   // This 12 feet comment should convert → This 3.7 metres comment should convert
   const WIDTH = 12; // This should NOT convert the variable value
   ```

2. **For markdown files:** Code blocks should be preserved
   ```markdown
   The room is 12 feet wide. ← Should convert

   ```go
   // Width is 12 feet ← Should convert (comment)
   width := 12        ← Should NOT convert (code)
   ```
   ```

3. **Check file extensions:** Ensure files have correct extensions (.go, .js, .py, .md, etc.)

## Advanced Configuration

### Custom Unit Mappings

Add your own unit conversion rules:
```json
{
  "customMappings": {
    "foot": "metre",
    "lb": "kg",
    "custom_unit": "metric_equivalent"
  }
}
```

### Regex Pattern Examples

Common exclusion patterns for idiomatic expressions:
```json
{
  "excludePatterns": [
    "miles?\\s+(?:away|apart|ahead|behind)",
    "inch\\s+by\\s+inch",
    "every\\s+inch",
    "tons?\\s+of\\s+(?:fun|work|stuff|things)",
    "pounds?\\s+of\\s+(?:pressure|force)\\b(?!\\s*\\d)",
    "cold\\s+feet",
    "foot\\s+(?:in\\s+the\\s+door|the\\s+bill)",
    "pound\\s+(?:the\\s+pavement|the\\s+table)"
  ]
}
```

### Detection Fine-Tuning

Adjust detection sensitivity:
```json
{
  "detection": {
    "minConfidence": 0.6,           // Higher = fewer false positives
    "maxNumberDistance": 3,         // Max words between number and unit
    "detectCompoundUnits": true,    // "6-foot fence"
    "detectWrittenNumbers": true    // "five feet"
  }
}
```

## Testing Your Configuration

### Test Specific Cases

Create test files to verify your configuration:

1. **Create test file:**
   ```bash
   echo "The room is 12 feet wide. I'm miles away from home." > test.txt
   ```

2. **Test conversion:**
   ```bash
   m2e-cli -units -input test.txt
   # Expected: "The room is 3.7 metres wide. I'm miles away from home."
   ```

3. **Test code-aware processing:**
   ```bash
   echo '// Width is 12 feet
   const width = 12;' > test.go

   m2e-cli -units -input test.go
   # Expected: Comment converts, code doesn't
   ```

### Debugging Output

For detailed debugging, you can examine the configuration loading:
```bash
# This will show any configuration warnings
m2e-cli -units "test text" 2>&1 | grep -i warning
```

## Getting Help

If you're still having issues:

1. **Check the main README:** [README.md](../README.md)
2. **Verify your use case:** Ensure unit conversion is appropriate for your text
3. **Test with simple examples:** Start with basic conversions before complex cases
4. **Reset configuration:** Remove config file to use defaults as a baseline

## Configuration File Location

- **macOS/Linux:** `~/.config/m2e/unit_config.json`
- **Windows:** `%USERPROFILE%\.config\m2e\unit_config.json`

The configuration file is created automatically with default values when first needed.
