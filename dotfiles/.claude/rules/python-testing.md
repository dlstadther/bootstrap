---
paths:
  - "**/test_*.py"
  - "**/*_test.py"
  - "**/tests/**/*.py"
  - "**/test/**/*.py"
---

Always use `pytest` (not `unittest`) for Python unit tests.

When writing more than one test case for the same pattern, use `@pytest.mark.parametrize` with `pytest.param(..., id="<descriptive-id>")` for every case — never repeat test functions or use bare tuples without IDs.

Example:
```python
import pytest

@pytest.mark.parametrize("input,expected", [
    pytest.param(1, 2, id="positive integer"),
    pytest.param(-1, 0, id="negative integer"),
    pytest.param(0, 1, id="zero"),
])
def test_increment(input, expected):
    assert increment(input) == expected
```
