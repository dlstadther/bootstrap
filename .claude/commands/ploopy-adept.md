# Ploopy Adept ASCII Art Convention

Use this skill when documenting a Ploopy Adept trackball layout config.

A complete Adept config description has four components:
- **6 button labels** (one per button, drawn in position)
- **Rotation** in degrees (the physical rotation set in firmware)
- **Direction** of that rotation (CW or CCW)
- **DPI** value

---

## Button Numbering

Button 1 is the **lower-left** button. Buttons are numbered **clockwise**:

```
 B2  B3  B4  B5
 B1          B6
```

Physical positions:
- **B1** Lower-left (large)
- **B2** Upper-left (large, tall — shares the left edge with B1)
- **B3** Upper-center-left (smaller)
- **B4** Upper-center-right (smaller)
- **B5** Upper-right (large, tall — shares the right edge with B6)
- **B6** Lower-right (large)

---

## ASCII Art Templates

Choose the template based on rotation value:

### 0° — vertical orientation (`-15 <= rotation <= 15`)

```
 +------------+--------+--------+------------+
 |            |        |        |            |
 |    [B2]    |  [B3]  |  [B4]  |    [B5]    |
 |            +--------+--------+            |
 +------+-----+                 +-----+------+
 |      |                             |      |
 |  B1  |          ( O )              |  B6  |
 |      |                             |      |
 +------+-----------------------------+------+
```

Cell widths: B2/B5 = 12 chars, B3/B4 = 8 chars, B1/B6 = 6 chars. Center gap = 29 chars.

### 90° CCW — top of device points left (`rotation > 15`)

B2–B5 run down the left column (B5 at top); B6 is upper-right, B1 is lower-right.

```
 +--------+--------+
 |  [B5]  |  [B6]  |
 +--------+        |
 |  [B4]  |  ( O ) |
 +--------+        |
 |  [B3]  |        |
 +--------+--------+
 |  [B2]  |  [B1]  |
 +--------+--------+
```

### 90° CW — top of device points right (`rotation < -15`)

B2–B5 run down the right column (B2 at top); B1 is upper-left, B6 is lower-left.

```
 +--------+--------+
 |  [B1]  |  [B2]  |
 |        +--------+
 |  ( O ) |  [B3]  |
 |        +--------+
 |        |  [B4]  |
 +--------+--------+
 |  [B6]  |  [B5]  |
 +--------+--------+
```

---

## Config Documentation Block

Use this format when adding a layout to `ploopy/adept/README.md`:

```markdown
### <config-name>

| Field     | Value            |
|-----------|------------------|
| Rotation  | <degrees>°       |
| Direction | <CW / CCW / None>|
| DPI       | <value>          |

<ascii art from appropriate template above, labels filled in>

| Button | Position           | Keycode | Label |
|--------|--------------------|---------|-------|
| B1     | Lower-left         | ...     | ...   |
| B2     | Upper-left         | ...     | ...   |
| B3     | Upper-center-left  | ...     | ...   |
| B4     | Upper-center-right | ...     | ...   |
| B5     | Upper-right        | ...     | ...   |
| B6     | Lower-right        | ...     | ...   |
```

---

## Common Keycodes

| Keycode        | Label  | Description                       |
|----------------|--------|-----------------------------------|
| KC_MS_BTN1     | LClick | Left mouse button                 |
| KC_MS_BTN2     | RClick | Right mouse button                |
| KC_WWW_BACK    | Back   | Browser back                      |
| KC_WWW_FORWARD | Fwd    | Browser forward                   |
| CUSTOM(0)      | DPI+   | Increase DPI                      |
| CUSTOM(1)      | DragSc | Toggle drag scroll                |
| CUSTOM(2)      | Cust2  | Unknown custom action             |
| CUSTOM(3)      | Cust3  | Unknown custom action             |
| 0xc1           | MCtrl  | Mission Control (macOS)           |
| KC_TRNS        | ▽      | Transparent (pass to lower layer) |
