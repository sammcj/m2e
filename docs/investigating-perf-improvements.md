The performance analysis clearly shows the bottleneck:

Performance Analysis Results:

🔍 Major Bottleneck Identified:
- 90% of CPU time is spent in regex processing, specifically:
  - regexp.(*Regexp).tryBacktrack (81.30% of total time)
  - regexp/syntax.(*Inst).MatchRunePos (33.33% of time)
  - unicode.SimpleFold (13.82% of time)

📍 Source of the Problem:
- The issue is in the converter's contextual word pattern matching
- (*ContextualWordPatterns).IsExcluded accounts for 85.37% of time
- (*ContextAwareWordDetector).findPatternMatches is doing expensive regex operations

🎯 Specific Areas:
1. Comment ignore processing - heavy regex matching
2. Contextual word detection - complex pattern matching
3. Code block processing - multiple regex operations per file

💡 Performance Improvements Identified:

1. Regex Compilation: The regexes are likely being compiled repeatedly instead of being cached
2. Backtracking: Complex regex patterns are causing expensive backtracking
3. Redundant Processing: The same patterns are being matched multiple times per file

🚀 Potential Optimizations:
- Cache compiled regex patterns
- Simplify complex regex patterns
- Use more efficient string matching for simple cases
- Consider using Boyer-Moore or other string algorithms for basic word matching

⏺ The performance bottleneck is clear: 90% of the time is spent in regex processing within the converter's contextual word pattern matching system. The slow performance you experienced
is due to complex regex backtracking operations in the IsExcluded and findPatternMatches functions.

The main culprits are:
- Complex regex patterns that cause expensive backtracking
- Regex compilation happening repeatedly instead of being cached
- Multiple pattern matching operations per file

To significantly improve performance, the converter package would need optimization around regex usage - such as caching compiled patterns, simplifying complex regexes, or using more efficient string matching algorithms for simpler cases.
