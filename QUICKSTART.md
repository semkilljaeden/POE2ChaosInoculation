# POE2 Chaos Crafter — Quick Start Guide

> **Automate Chaos Orb crafting in Path of Exile 2 with OCR-powered mod detection.**

---

## Requirements

| Requirement | Notes |
|---|---|
| **Go 1.21+** | https://go.dev/dl/ |
| **Tesseract OCR** | https://github.com/UB-Mannheim/tesseract/wiki — install basic package |
| **C Compiler** | Windows: MinGW-w64 or TDM-GCC (needed by robotgo) |

Config is saved to `~/.poe2_crafter_config.json` automatically.

---

## 1. Build & Launch

```bash
cd c:\development\POE2ChaosInoculation

make run-web        # build + launch with web GUI  (recommended)
make run-debug      # build + launch with debug logging enabled
make build          # compile only → poe2crafter.exe
```

Then open your browser at **http://localhost:8080**

> The web files (`index.html`, `style.css`, `app.js`) are embedded in the binary at compile time.
> You **must** rebuild with `make build` after any source or web file changes.

---

## 2. Interface Overview

```
┌─────────────────────────────────────────────────────────────┐
│  POE2 Chaos Crafter              UI [en▾]  Game [en▾]  ●   │
├──────────────┬──────────────────────────────────────────────┤
│  Dashboard   │  Config                                       │
└──────────────┴──────────────────────────────────────────────┘
```

| Tab | Purpose |
|---|---|
| **Dashboard** | Live crafting monitor — status, snapshots, mod stats, history |
| **Config** | View & edit configuration sections; launch the Setup Wizard |

---

## 3. First-Time Setup (Setup Wizard)

Click **Config → Setup Wizard** to open the 8-step wizard modal.

```
┌──────────────────────────────────────────────────────────┐
│                    POE2 Chaos Crafter               ✕    │
│                                                          │
│   ①──②──③──④──⑤──⑥──⑦──⑧                              │
│                                                          │
│   Step 1: Configuration                                  │
│   Load existing config or start fresh?                   │
│                                                          │
│   [ Load Existing ]   [ Start Fresh ]                    │
└──────────────────────────────────────────────────────────┘
```

### Step-by-step

| Step | What to do |
|---|---|
| **1 — Config** | Choose **Load Existing** (reuse saved config) or **Start Fresh** |
| **2 — Backpack Grid** | Click **Capture** for the **top-left cell** of your backpack, then the **bottom-right cell**. Switch to the game within the 5-second countdown |
| **3 — Chaos Orb** | Click **Capture** and hover over your Chaos Orb in the stash |
| **4 — Item Dimensions** | Pick how many grid cells wide & tall the item is (e.g. 2×3) |
| **5 — Batch Areas** | Enter grid row/col for **Workbench**, **Pending Area** (items to craft), and **Result Area** (finished items) |
| **6 — Tooltip Area** | Hover an item so the tooltip appears, **Capture** the top-left and bottom-right corners of the tooltip, then click **Validate OCR** to confirm text is readable |
| **7 — Target Mods** | Add mods using the quick-template dropdown or type custom mod strings |
| **8 — Review & Save** | Check the summary, then click **Save Config** or **Save & Start** |

> **Capture countdown:** When you click Capture, you have **5 seconds** to switch to the game window and position your cursor. The countdown appears as a large overlay in the center of the screen.

---

## 4. Editing Config Sections (Returning Users)

After the first setup you don't need the wizard. Each config section on the **Config tab** has its own **Edit** button:

```
┌─────────────────────────────────────────────────────┐
│  Current Configuration      [ Reload ] [ Setup Wizard ]│
│                                                        │
│  ┌─ Positions ──────────────────────── [ Edit ] ─┐   │
│  │  Chaos Orb         (1234, 567)                 │   │
│  │  Backpack TL       (100, 200)                  │   │
│  │  Backpack BR       (500, 450)                  │   │
│  └───────────────────────────────────────────────┘   │
│                                                        │
│  ┌─ Item ───────────────────────────── [ Edit ] ─┐   │
│  │  Item Size         2 x 3 cells                │   │
│  └───────────────────────────────────────────────┘   │
│                                                        │
│  ┌─ Target Mods ────────────────────── [ Edit ] ─┐   │
│  │  1. Life 80+                                   │   │
│  │  2. Fire Res 35%+                              │   │
│  └───────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────┘
```

Clicking **Edit** expands an inline editor for just that section — same UI as the wizard, without leaving the page:

```
  ┌─ Positions ──────────────────────── [ Edit ] ─┐
  │  Chaos Orb         (1234, 567)                 │
  │  Backpack TL       (100, 200)                  │
  │  Backpack BR       (500, 450)                  │
  │  ···············································│  ← dashed separator
  │  Backpack Top-Left:  (100, 200)  [ Capture ]   │
  │  Backpack Bot-Right: (500, 450)  [ Capture ]   │
  │  Chaos Orb:          (1234, 567) [ Capture ]   │
  │                                                 │
  │  [ Save Config ]  [ Cancel ]                   │
  └─────────────────────────────────────────────────┘
```

| Section | What you can change |
|---|---|
| **Positions** | Re-capture Backpack TL/BR corners and Chaos Orb location |
| **Item** | Width & height in grid cells |
| **Batch Crafting** | Workbench slot, Pending Area, Result Area (row/col + size) |
| **Tooltip** | Re-capture tooltip corners + re-validate OCR |
| **Target Mods** | Add/remove mods without touching anything else |
| **Options** | Chaos per round, debug logging, save snapshots |

---

## 5. Target Mod Reference

Use these keywords in the **Target Mods** section (quick-template dropdown or custom input):

```
life <min>         → +X to Maximum Life
mana <min>         → +X to Maximum Mana
str <min>          → +X to Strength
dex <min>          → +X to Dexterity
int <min>          → +X to Intelligence
spirit <min>       → +X to Spirit
fire-res <min>     → X% Fire Resistance
cold-res <min>     → X% Cold Resistance
light-res <min>    → X% Lightning Resistance
chaos-res <min>    → X% Chaos Resistance
armor <min>        → X Armour
evasion <min>      → X Evasion
es <min>           → +X to Maximum Energy Shield
movespeed <min>    → X% Movement Speed
attackspeed <min>  → X% Attack Speed
castspeed <min>    → X% Cast Speed
crit-dmg <min>     → X% Critical Damage Bonus
spell-level <n>    → +N to Level of all Spell Skills
proj-level <n>     → +N to Level of all Projectile Skills
```

**Examples:**
```
life 80            → accept items with Life ≥ 80
fire-res 35        → accept items with Fire Res ≥ 35%
life 60            → (multiple mods) add more rows for AND logic
```

> Multiple mods = **ALL** must be met on the same item.

---

## 6. Running a Crafting Session

1. Switch to the **Dashboard** tab
2. Have POE2 open with items in the Pending Area and Chaos Orbs in the stash
3. Click **Start** — a 5-second countdown gives you time to alt-tab to the game

```
┌─────────────────────────────────────────────────────┐
│  Crafting Status                                     │
│  State:       Running                                │
│  Item:        #4                                     │
│  Roll:        12 / 50                                │
│  Total Rolls: 63                                     │
│  Speed:       24.3 / min                             │
│  Duration:    2m 36s                                 │
│                                                      │
│  [ Start ]  [ Pause ]  [ Stop ]                      │
└─────────────────────────────────────────────────────┘
```

| Button | Action |
|---|---|
| **Start** | Begin crafting; 5-second countdown, then auto-plays |
| **Pause** | Freeze after current roll; click again to Resume |
| **Stop** | End the session |

The **Live Game Snapshot** panel updates every roll. The **Mod Statistics** table tracks how often each mod appears and at what values across all rolls.

---

## 7. Batch Crafting Layout

```
┌──────────────────────────── Backpack (5×12) ─────────────────────────────┐
│  [Workbench]  [·][·][·]  [Pending Area  ···]  [Result Area  ···]         │
│  [ row 0 ]    [·][·][·]  [items to craft ·]   [successful  ·]            │
│  [ col 0 ]    [·][·][·]  [             ···]   [items       ·]            │
└──────────────────────────────────────────────────────────────────────────┘
```

- **Workbench** — the single grid cell where the current item sits while being crafted
- **Pending Area** — items waiting to be crafted (bot picks from here one at a time)
- **Result Area** — items that met target mods are moved here automatically

---

## 8. Language Support

Select independently in the header:

| Selector | Purpose |
|---|---|
| **UI** | Interface language (English / 简体中文) |
| **Game** | OCR language matching your POE2 client (English / 简体中文) |

> Changing Game language clears Target Mods — re-add them after switching.

---

## 9. Troubleshooting

| Symptom | Fix |
|---|---|
| OCR not detecting mods | Re-capture tooltip corners; run **Validate OCR** in the Tooltip section |
| Wrong positions after resolution change | Re-capture Backpack TL/BR in the Positions section |
| Items not moving to result area | Check Batch Crafting row/col values match actual backpack layout |
| Web UI not loading | Rebuild with `make run-web`; web files are embedded at compile time |
| Chinese text not recognized | Set **Game** language selector to 简体中文 before capturing tooltip |

Enable **OCR Debug Logging** (Options section) to write per-roll screenshots to the `snapshots/` folder.

---

## 10. Config File Location

```
Windows:  C:\Users\<you>\.poe2_crafter_config.json
Linux:    ~/.poe2_crafter_config.json
macOS:    ~/.poe2_crafter_config.json
```

Back this file up after a successful setup — use **Load Existing** in the wizard to restore it.

---

> **Disclaimer:** Automation tools may violate the Path of Exile 2 Terms of Service. Use at your own risk. This project is for educational purposes.
