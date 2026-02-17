# POE2 Chaos Crafter - Quick Start

**Simplified: ONE mod at a time! No complex setup!**

## ⚡ Super Quick Start

### Windows:
```cmd
setup_vscode.bat
```

### Linux/macOS:
```bash
./setup_vscode.sh
```

Then: **Open folder in VSCode → Press F5**

## 📋 What You Get

Just specify ONE mod like:
- `life 80` → Find +80 Life or better
- `fire-res 35` → Find 35% Fire Res or better
- `str 45` → Find +45 Strength or better

That's it! No multiple mod complexity.

## 🎮 Example

```
Enter mod: life 80

[1/1000] Crafting...
[2/1000] Crafting...
[47/1000] Crafting...

🎉 SUCCESS! Found: Life 80+ = 85
```

## 📦 Files You Need

```
poe2_crafter.go       ← Main program (single mod only)
go.mod                ← Dependencies  
.vscode/              ← VSCode config
setup_vscode.bat      ← Windows setup
setup_vscode.sh       ← Linux/Mac setup
VSCODE_SETUP.md       ← Detailed guide
```

## 🔧 Requirements

1. **Go** - https://go.dev/dl/
2. **Tesseract OCR** - https://github.com/UB-Mannheim/tesseract/wiki
   - Just install the basic package (tesseract.exe)
   - ✅ No development headers needed!
   - ✅ No Leptonica libraries needed!
3. **C Compiler** - MinGW (Windows) / gcc (Linux) / Xcode (Mac)
   - Only needed for robotgo (mouse/keyboard control)
   - Windows: MinGW-w64 or TDM-GCC
4. **VSCode** (optional) - https://code.visualstudio.com/

**What's Simplified:** The program now calls Tesseract via command-line instead of using CGO bindings. This eliminates the complex Tesseract/Leptonica header requirements!

## 🛠️ Manual Build (Alternative to VSCode)

If you prefer command-line:

**Windows:**
```bash
cd c:\development\go
set CGO_ENABLED=1
set PATH=C:\ProgramData\mingw64\mingw64\bin;%PATH%
go build poe2_crafter.go
poe2_crafter.exe
```

**Linux/macOS:**
```bash
go build poe2_crafter.go
./poe2_crafter
```

## 🚀 VSCode Usage

| Action | Shortcut |
|--------|----------|
| Build | `Ctrl+Shift+B` |
| Run | `F5` |
| Stop | `Shift+F5` |
| Terminal | `` Ctrl+` `` |

## 📝 Available Mods

```
life <min>        → +X to maximum Life
mana <min>        → +X to maximum Mana
str <min>         → +X to Strength
dex <min>         → +X to Dexterity
int <min>         → +X to Intelligence
fire-res <min>    → X% Fire Resistance
cold-res <min>    → X% Cold Resistance
light-res <min>   → X% Lightning Resistance
chaos-res <min>   → X% Chaos Resistance
armor <min>       → X Armour
evasion <min>     → X Evasion
es <min>          → +X Energy Shield
movespeed <min>   → X% Movement Speed
```

## 🐛 Debug Mode

Enable to see screenshots and OCR text:
```
Save debug screenshots? (y/n): y
```

Creates:
- `debug_0001.png`, `debug_0002.png`, etc.
- Shows exact OCR text output

## ⚠️ Disclaimer

Using automation violates POE2 ToS. Educational purposes only.

---

**See VSCODE_SETUP.md for detailed instructions**
