# Write Test Prompt

Use this prompt as a scaffold:

```text
Write tests for the requested Go service behavior.

Requirements:
- choose the narrowest useful test level
- cover success, business rejection, and dependency failure where relevant
- prefer table-driven tests when they improve clarity
- avoid deep framework mocks unless there is no better seam
- explain what behavior the test protects
```
