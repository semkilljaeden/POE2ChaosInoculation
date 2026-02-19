# POE2 混沌石自动制作工具

> 使用 OCR 词缀识别与浏览器图形界面，在《流放之路 2》中自动化混沌石刷词缀流程。

[English](README.md) | 简体中文

---

## 功能特性

- **OCR 词缀识别** — 使用 Tesseract OCR 实时读取物品提示框
- **批量制作** — 自动处理整列物品（待处理区 → 工作台 → 结果区）
- **灵活的词缀目标设置** — 内置 20+ 种词缀模板，支持自定义正则表达式
- **浏览器界面** — 本地 Web 界面，无需安装 Electron 或其他软件
- **分节配置编辑** — 单独修改任一配置项（坐标、词缀、提示框等），无需重新运行向导
- **双语支持** — 界面与 OCR 同时支持 English 和 简体中文

---

## 环境要求

| 依赖项 | 说明 |
|---|---|
| **Go 1.21+** | https://go.dev/dl/ |
| **Tesseract OCR** | https://github.com/UB-Mannheim/tesseract/wiki — 安装基础版即可 |
| **C 编译器** | Windows 下使用 MinGW-w64 或 TDM-GCC（robotgo 依赖） |

---

## 快速开始

```bash
git clone https://github.com/yourname/POE2ChaosInoculation
cd POE2ChaosInoculation

make run-web        # 编译并启动（含 Web 界面，推荐）
make run-debug      # 编译并启动（含调试日志）
make build          # 仅编译 → poe2crafter.exe
```

在浏览器中打开 **http://localhost:8080**。

> Web 资源（`index.html`、`style.css`、`app.js`）在编译时嵌入二进制文件。
> 修改源码或 Web 文件后，需重新运行 `make build`。

---

## 界面说明

### Dashboard — 实时制作监控

![Dashboard](docs/screenshots/dashboard_clean.png)

Dashboard 面板实时显示制作状态：

| 字段 | 说明 |
|---|---|
| **State（状态）** | 空闲 / 启动中 / 运行中 / 已停止 |
| **Item（物品）** | 当前批次中正在制作的物品编号 |
| **Roll（投掷）** | 当前物品的尝试次数 / 上限 |
| **Total Rolls（总投掷）** | 本次会话累计投掷次数 |
| **Speed（速度）** | 每分钟投掷次数 |
| **Duration（时长）** | 本次会话已用时间 |

**Parsed Mod Text** 面板显示最近一次识别到的词缀原始文本；**Tooltip** 面板显示游戏内提示框截图；**Mod Statistics** 表格（向下滚动可见）统计各词缀的出现频率与数值分布。

| 按钮 | 功能 |
|---|---|
| **Start（开始）** | 启动制作；5 秒倒计时后自动运行 |
| **Stop（停止）** | 结束本次会话 |

---

### Config — 当前配置

![Config](docs/screenshots/config_clean.png)

Config 标签页以分节形式展示所有配置项。每个分节右侧均有独立的 **Edit（编辑）** 按钮 — 点击后直接在页面内展开该节的编辑器，无需跳转页面或重新运行向导。

---

### 分节内联编辑

![Section Editor](docs/screenshots/section_editor.png)

点击任意分节的 **Edit** 按钮，即可展开与向导相同操作流程的内联表单：

- **Positions（坐标）** — 重新捕捉背包角点与混沌石位置
- **Item（物品）** — 设置物品占格宽高
- **Batch Crafting（批量制作）** — 设置工作台格、待处理区、结果区
- **Tooltip（提示框）** — 重新捕捉提示框角点并验证 OCR
- **Target Mods（目标词缀）** — 单独增删词缀，不影响其他配置
- **Options（选项）** — 每轮混沌石数量、调试日志、保存截图

点击 **Save Config** 保存，或 **Cancel** 放弃修改。

---

### 设置向导（首次使用）

![Wizard Modal](docs/screenshots/wizard_modal.png)

在 Config 标签页点击 **Setup Wizard（设置向导）** 打开 8 步引导弹窗：

| 步骤 | 操作说明 |
|---|---|
| **第 1 步 — 配置** | 加载已有配置文件，或从头开始 |
| **第 2 步 — 背包网格** | 捕捉背包左上角与右下角单元格位置 |
| **第 3 步 — 混沌石** | 将光标悬停在藏宝图中的混沌石上并捕捉位置 |
| **第 4 步 — 物品尺寸** | 设置物品占格的宽度与高度（如 2×3） |
| **第 5 步 — 批量区域** | 设定工作台格、待处理区和结果区的网格位置 |
| **第 6 步 — 提示框区域** | 捕捉提示框角点，然后点击 **Validate OCR** 验证 |
| **第 7 步 — 目标词缀** | 通过快捷模板下拉菜单或自定义正则添加词缀 |
| **第 8 步 — 确认保存** | 检查所有设置，点击 **Save Config** 或 **Save & Start** |

> **捕捉倒计时：** 点击任意 Capture 按钮后，屏幕会出现 5 秒倒计时，给你充足时间切换到游戏并移动光标至目标位置。

---

## 目标词缀参考

在 **Target Mods** 分节中使用以下关键字：

```
life <最小值>         → 最大生命值 +X
mana <最小值>         → 最大法力值 +X
str <最小值>          → 力量 +X
dex <最小值>          → 敏捷 +X
int <最小值>          → 智慧 +X
spirit <最小值>       → 精神 +X
fire-res <最小值>     → 火焰抗性 X%
cold-res <最小值>     → 冰霜抗性 X%
light-res <最小值>    → 闪电抗性 X%
chaos-res <最小值>    → 混沌抗性 X%
armor <最小值>        → 护甲 X
evasion <最小值>      → 闪避 X
es <最小值>           → 最大能量护盾 +X
movespeed <最小值>    → 移动速度 X%
attackspeed <最小值>  → 攻击速度 X%
castspeed <最小值>    → 施法速度 X%
crit-dmg <最小值>     → 暴击伤害加成 X%
spell-level <n>       → 所有法术技能等级 +N
proj-level <n>        → 所有远程技能等级 +N
```

**示例：**
```
life 80            → 接受生命值 ≥ 80 的物品
fire-res 35        → 接受火焰抗性 ≥ 35% 的物品
```

设置多个词缀时，物品必须**同时满足所有**条件。

---

## 批量制作布局

```
┌──────────────────────────── 背包（5×12） ─────────────────────────────┐
│  [工作台]     [·][·][·]  [待处理区  ···]  [结果区  ···]               │
│  [ 第2行 ]    [·][·][·]  [待制作物品]     [成功物品]                  │
│  [ 第5列 ]    [·][·][·]  [          ]     [        ]                  │
└───────────────────────────────────────────────────────────────────────┘
```

- **工作台（Workbench）** — 当前正在制作的物品所在的单个背包格
- **待处理区（Pending Area）** — 排队等待制作的物品；程序逐一取用
- **结果区（Result Area）** — 满足所有目标词缀的物品将自动移至此处

---

## 配置文件位置

```
Windows:  C:\Users\<用户名>\.poe2_crafter_config.json
Linux:    ~/.poe2_crafter_config.json
macOS:    ~/.poe2_crafter_config.json
```

成功配置后请备份此文件。使用向导中的 **Load Existing** 可随时恢复。

---

## 常见问题

| 现象 | 解决方法 |
|---|---|
| OCR 无法识别词缀 | 重新捕捉提示框角点；在 Tooltip 分节运行 **Validate OCR** |
| 修改分辨率后坐标错位 | 在 Positions 分节重新捕捉背包左上角和右下角 |
| 物品未移入结果区 | 检查 Batch Crafting 中的行列值是否与实际背包布局一致 |
| Web 界面无法加载 | 使用 `make run-web` 重新编译——Web 文件在编译时嵌入 |
| 中文文字无法识别 | 捕捉提示框前，将 **Game** 语言选择器切换为 简体中文 |

在 Options 分节启用 **OCR Debug Logging** 可将每次投掷的截图保存至 `snapshots/` 文件夹。

---

> **免责声明：** 自动化工具可能违反《流放之路 2》服务条款，使用风险由用户自行承担。本项目仅供学习研究使用。
