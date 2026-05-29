# Ploopy Adept Trackball (Madromys)

Firmware and Via keymap configurations for the Ploopy Adept Trackball (Madromys firmware).

## Firmware

`ploopyco_madromys_rev1_001_viam.uf2` — Via-compatible firmware for Rev1 hardware.

To flash: hold the button on the underside of the device while plugging in USB, then drag the `.uf2` file onto the drive that appears.

## Button Layout Convention

The Adept has 6 buttons. **Button 1** is the lower-left; buttons are numbered **clockwise**, making **Button 6** the lower-right.

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

## Layout Configs

### madromys_00

| Field     | Value |
|-----------|-------|
| Rotation  | 0°    |
| Direction | None  |
| DPI       | 1600  |

```
 +------------+--------+--------+------------+
 |            |        |        |            |
 |    Back    |  Fwd   | Cust2  |   MCtrl    |
 |            +--------+--------+            |
 +------+-----+                 +-----+------+
 |      |                             |      |
 |LClick|          ( O )              |RClick|
 |      |                             |      |
 +------+-----------------------------+------+
```

| Button | Position          | Keycode            | Label  |
|--------|-------------------|--------------------|--------|
| B1     | Lower-left        | KC_MS_BTN1         | LClick |
| B2     | Upper-left        | KC_WWW_BACK        | Back   |
| B3     | Upper-center-left | KC_WWW_FORWARD     | Fwd    |
| B4     | Upper-center-right| CUSTOM(2)          | Cust2  | <!-- TODO: identify keycode label -->
| B5     | Upper-right       | 0xc1               | MCtrl  |
| B6     | Lower-right       | KC_MS_BTN2         | RClick |

---

### madromys_01

| Field     | Value |
|-----------|-------|
| Rotation  | 90°   |
| Direction | CCW   |
| DPI       | 1600  |

```
 +--------+--------+
 |  Fwd   |        |
 +--------+  MCtrl |
 |  Back  |        |
 +--------+  ( O ) |
 | DragSc |        |
 +--------+--------+
 | LClick | RClick |
 +--------+--------+
```

| Button | Position           | Keycode        | Label  |
|--------|--------------------|----------------|--------|
| B1     | Lower-left         | KC_MS_BTN2     | RClick |
| B2     | Upper-left         | KC_MS_BTN1     | LClick |
| B3     | Upper-center-left  | CUSTOM(1)      | DragSc |
| B4     | Upper-center-right | KC_WWW_BACK    | Back   |
| B5     | Upper-right        | KC_WWW_FORWARD | Fwd    |
| B6     | Lower-right        | 0xc1           | MCtrl  |

---

## Resources

- [VIA](https://usevia.app/) — standard keymap configuration
- [VIA (Plodah build)](https://via.plodah.uk/) — Plodah's fork with Ploopy extras
- [plodah/ploopy_viamenus](https://github.com/plodah/ploopy_viamenus) — Plodah's firmware additions

## Common Keycodes

| Keycode        | Label  | Description                      |
|----------------|--------|----------------------------------|
| KC_MS_BTN1     | LClick | Left mouse button                |
| KC_MS_BTN2     | RClick | Right mouse button               |
| KC_WWW_BACK    | Back   | Browser back                     |
| KC_WWW_FORWARD | Fwd    | Browser forward                  |
| CUSTOM(0)      | DPI+   | Increase DPI                     |
| CUSTOM(1)      | DragSc | Toggle drag scroll               |
| CUSTOM(2)      | Cust2  | Unknown custom action            |
| CUSTOM(3)      | Cust3  | Unknown custom action            |
| 0xc1           | MCtrl  | Mission Control (macOS)          |
| KC_TRNS        | ▽      | Transparent (pass to lower layer)|
