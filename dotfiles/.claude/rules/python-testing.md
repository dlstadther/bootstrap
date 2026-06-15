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

## Naming Conventions

Name tests after what they exercise, not after how they're grouped.

- **Simple functions** — `test_<function_under_test>`:
  ```python
  def test_normalize_path(): ...
  ```
- **More complex cases** — `test_<function_under_test>__<action>__<expectation>`, where `<action>` is the input or scenario and `<expectation>` is the result:
  ```python
  def test_normalize_path__trailing_slash__stripped(): ...
  def test_normalize_path__empty_input__raises_valueerror(): ...
  ```
- **Classes and their methods** — name the test class `Test<ClassName>` and put method tests inside it:
  ```python
  class TestPathResolver:
      def test_resolve(self): ...
      def test_resolve__missing_file__returns_none(self): ...
  ```

Never encode the method or scenario into the test class name. Use `TestPathResolver` — never `TestPathResolverResolveMissingFile`. The class names the unit under test; the action and expectation belong in the method name.
