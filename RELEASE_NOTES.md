# Release Notes

## v1.0.0 — 2026-02-22

First public release of **POE2 Chaos Crafter** — a browser-based automation tool for Chaos Orb crafting in Path of Exile 2.

---

### What's New

#### Web GUI
- Local web interface served at `http://localhost:8080` — no Electron, no extra install
- **Dashboard** tab with real-time crafting status: state, item number, roll count, total rolls, speed (rolls/min), session duration
- **Live Game Snapshot** panel — auto-refreshes after every roll, giving a live view of the game at the current roll rate
- **Tooltip** panel — shows the OCR tooltip screenshot captured each roll
- **Mod Statistics** table — tracks min/max/avg/probability for every observed mod
- **Success sound** — ascending C-major arpeggio plays in the browser when a target mod is found

#### Config Tab
- View all configuration sections at a glance
- **Per-section inline editing** — click Edit on any section (Positions, Item, Batch Crafting, Tooltip, Target Mods, Options) to modify it in-place without re-running the wizard
- **Setup Wizard** button opens the full 8-step guided modal for first-time setup or complete reconfiguration

#### Session Controls
- **Start** / **Stop** buttons — no pause (keeps sessions clean and deterministic)
- 5-second countdown before crafting begins

#### i18n
- UI language: English / 简体中文
- Game OCR language: English / 简体中文

---

### Requirements

| | |
|---|---|
| OS | Windows 10 / 11 (64-bit) |
| Tesseract OCR | Must be installed separately — [download](https://github.com/UB-Mannheim/tesseract/wiki) |
| Config file | Created on first run via Setup Wizard |

---

### How to Run

Double-click **`run.bat`** — or run from a terminal:

```bat
poe2crafter.exe --web
```

Open **http://localhost:8080** in your browser, then use the Setup Wizard to configure screen positions and target mods.

---

### Known Limitations

- Windows only (robotgo screen capture and mouse automation)
- Requires Tesseract OCR to be in PATH
- High DPI / display scaling may affect coordinate capture accuracy — capture with the game running at target resolution
