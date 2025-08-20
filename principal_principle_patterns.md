# Principal/Principle Detection Patterns

## Design Overview

Unlike existing contextual words (license/licence), principal/principle require semantic context detection rather than grammatical role detection. We need high-confidence patterns for specific technical domains.

## Principal Patterns (Person/Entity in Authority)

### AWS/Cloud Computing Contexts
- `AWS IAM principal` → always "principal"
- `service principal` → always "principal"
- `user principal` → always "principal" 
- `principal ARN` → always "principal"
- `authentication principal` → always "principal"
- `security principal` → always "principal" (when referring to identity)
- `Kerberos principal` → always "principal"
- `OAuth principal` → always "principal"

### Database/Authentication Contexts
- `database principal` → always "principal"
- `login principal` → always "principal"
- `principal name` → always "principal"
- `principal ID` → always "principal"

### Finance Contexts
- `principal amount` → always "principal"
- `loan principal` → always "principal"
- `principal payment` → always "principal"

## Principle Patterns (Fundamental Rules/Beliefs)

### Security/Design Contexts
- `principle of least privilege` → always "principle"
- `principle of least privileged` → always "principle" (correcting common error)
- `security principle` → always "principle" (when referring to guidelines)
- `design principle` → always "principle"
- `fundamental principle` → always "principle"
- `core principle` → always "principle"
- `guiding principle` → always "principle"
- `basic principle` → always "principle"

### Technical/Engineering Contexts
- `DRY principle` → always "principle"
- `SOLID principle` → always "principle"
- `engineering principle` → always "principle"
- `architectural principle` → always "principle"
- `programming principle` → always "principle"

## Pattern Implementation Strategy

1. **High Confidence Only**: Only convert when confidence > 90%
2. **Specific Contexts**: Target technical domains where usage is well-defined
3. **No Ambiguous Cases**: Leave unclear contexts unchanged
4. **Preserve Case**: Maintain original capitalisation

## Example Conversions

```
- Follow the principal of least privileged security.
+ Follow the principle of least privileged security.

- The AWS IAM principle has permission to access S3.
+ The AWS IAM principal has permission to access S3.

- Our service principle needs database access.
+ Our service principal needs database access.

- This violates the security principals we established.
+ This violates the security principles we established.
```

## Non-Conversions (Ambiguous Cases)
- "school principal" vs "school principle" → leave unchanged
- "principal concern" vs "principle concern" → leave unchanged
- General usage without clear technical context → leave unchanged