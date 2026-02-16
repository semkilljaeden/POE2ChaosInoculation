// ===== Internationalization (i18n) =====
let currentLang = localStorage.getItem('poe2crafter-lang') || 'en';
let gameLang = localStorage.getItem('poe2crafter-game-lang') || 'en';

const i18n = {
    en: {
        'title': 'POE2 Chaos Crafter',
        'connected': 'Connected',
        'disconnected': 'Disconnected',
        'tab.dashboard': 'Dashboard',
        'tab.wizard': 'Wizard',
        'tab.config': 'Config',
        'status.title': 'Crafting Status',
        'status.state': 'State:',
        'status.item': 'Item:',
        'status.roll': 'Roll:',
        'status.totalRolls': 'Total Rolls:',
        'status.speed': 'Speed:',
        'status.duration': 'Duration:',
        'btn.start': 'Start',
        'btn.pause': 'Pause',
        'btn.resume': 'Resume',
        'btn.stop': 'Stop',
        'btn.refresh': 'Refresh',
        'btn.back': 'Back',
        'btn.next': 'Next',
        'btn.capture': 'Capture',
        'btn.add': 'Add',
        'panel.snapshot': 'Live Game Snapshot',
        'panel.ocrText': 'Parsed Mod Text',
        'panel.tooltip': 'Tooltip',
        'panel.modStats': 'Mod Statistics',
        'panel.history': 'Round History',
        'table.mod': 'Mod',
        'table.count': 'Count',
        'table.min': 'Min',
        'table.max': 'Max',
        'table.avg': 'Avg',
        'table.prob': 'Prob%',
        'empty.noData': 'No data yet',
        'empty.noRounds': 'No rounds yet',
        'empty.waiting': 'Waiting for crafting to start...',
        'empty.noConfig': 'No configuration found. Use the Wizard to create one.',
        'empty.loading': 'Loading...',
        'empty.noMods': 'No mods added yet',
        'empty.noTargetMods': 'No target mods configured',
        'state.idle': 'Idle',
        'state.starting': 'Starting...',
        'state.running': 'Running',
        'state.paused': 'Paused',
        'state.stopped': 'Stopped',
        'state.startingIn': 'Starting in {n}...',
        'wiz.step1.title': 'Step 1: Configuration',
        'wiz.step1.desc': 'Load existing config or start fresh?',
        'wiz.step1.load': 'Load Existing',
        'wiz.step1.fresh': 'Start Fresh',
        'wiz.step2.title': 'Step 2: Backpack Grid',
        'wiz.step2.desc': 'Capture the top-left and bottom-right corners of your backpack grid (5 rows x 12 columns).',
        'wiz.topLeft': 'Top-Left Corner:',
        'wiz.bottomRight': 'Bottom-Right Corner:',
        'wiz.notSet': 'Not set',
        'wiz.step3.title': 'Step 3: Chaos Orb Position',
        'wiz.step3.desc': 'Capture where the chaos orb is in your stash.',
        'wiz.chaosOrb': 'Chaos Orb:',
        'wiz.step4.title': 'Step 4: Item Dimensions',
        'wiz.step4.desc': 'How many grid cells does the item occupy?',
        'wiz.width': 'Width (cells):',
        'wiz.height': 'Height (cells):',
        'wiz.step5.title': 'Step 5: Batch Crafting Areas',
        'wiz.step5.desc': 'Specify workbench, pending area, and result area positions (in grid cell coordinates row, col).',
        'wiz.workbench': 'Workbench',
        'wiz.pendingArea': 'Pending Area',
        'wiz.resultArea': 'Result Area',
        'wiz.row': 'Row (0-4):',
        'wiz.col': 'Col (0-11):',
        'wiz.topLeftRow': 'Top-Left Row:',
        'wiz.topLeftCol': 'Top-Left Col:',
        'wiz.step6.title': 'Step 6: Tooltip Area',
        'wiz.step6.desc': 'Hover over an item to show the tooltip, then capture its corners.',
        'wiz.validateOCR': 'Validate OCR',
        'wiz.step7.title': 'Step 7: Target Mods',
        'wiz.step7.desc': 'Select which mods to search for. Format:',
        'wiz.step7.format': 'mod_name min_value',
        'wiz.quickTemplate': '-- Quick Template --',
        'wiz.minValue': 'Min value',
        'wiz.addCustom': 'Add Custom',
        'wiz.customPlaceholder': 'e.g. life 80, fire-res 30',
        'wiz.step8.title': 'Step 8: Options & Review',
        'wiz.chaosPerRound': 'Chaos Orbs per Round:',
        'wiz.ocrDebug': 'Enable OCR debug logging',
        'wiz.saveSnapshots': 'Save all snapshots',
        'wiz.review': 'Review',
        'wiz.saveConfig': 'Save Config',
        'wiz.saveAndStart': 'Save & Start',
        'wiz.switchToGame': 'Switch to game window now! 5...',
        'cfg.title': 'Current Configuration',
        'cfg.reload': 'Reload',
        'cfg.editWizard': 'Edit in Wizard',
        'cfg.positions': 'Positions',
        'cfg.chaosOrb': 'Chaos Orb',
        'cfg.bpTopLeft': 'Backpack Top-Left',
        'cfg.bpBottomRight': 'Backpack Bottom-Right',
        'cfg.item': 'Item',
        'cfg.itemSize': 'Item Size',
        'cfg.batchCrafting': 'Batch Crafting',
        'cfg.workbench': 'Workbench',
        'cfg.pendingArea': 'Pending Area',
        'cfg.pendingSize': 'Pending Size',
        'cfg.resultArea': 'Result Area',
        'cfg.resultSize': 'Result Size',
        'cfg.tooltip': 'Tooltip',
        'cfg.offsetFromItem': 'Offset from Item',
        'cfg.size': 'Size',
        'cfg.targetMods': 'Target Mods',
        'cfg.options': 'Options',
        'cfg.chaosPerRound': 'Chaos per Round',
        'cfg.ocrDebug': 'OCR Debug Logging',
        'cfg.saveSnapshots': 'Save All Snapshots',
        'cfg.enabled': 'Enabled',
        'cfg.disabled': 'Disabled',
        'lang.ui': 'UI',
        'lang.game': 'Game',
        'cfg.gameLanguage': 'Game Language',
        'toast.gameLangChanged': 'Game language changed. Re-add target mods if needed.',
        'toast.targetFound': 'Target found: {mod} = {value}!',
        'toast.sessionEnded': 'Crafting session ended',
        'toast.startFailed': 'Failed to start crafting',
        'toast.pauseFailed': 'Failed to toggle pause',
        'toast.stopFailed': 'Failed to stop crafting',
        'toast.configLoaded': 'Config loaded',
        'toast.freshConfig': 'Starting fresh config',
        'toast.captureFailed': 'Capture failed',
        'toast.selectMod': 'Select a mod template',
        'toast.enterMin': 'Enter a minimum value',
        'toast.enterMod': 'Enter a mod string',
        'toast.invalidMod': 'Invalid mod format. Try: life 80',
        'toast.modAdded': 'Added: {desc}',
        'toast.modParseFailed': 'Failed to parse mod',
        'toast.configSaved': 'Configuration saved!',
        'toast.saveFailed': 'Failed to save config',
        'toast.captureCorners': 'Capture tooltip corners first',
        'toast.validationFailed': 'Validation failed',
        'toast.configLoadError': 'Error loading config.',
        'toast.noConfig': 'No existing config found.',
        'toast.configLoadSuccess': 'Config loaded! Navigate steps to review/edit.',
        'toast.saveError': 'Save error',
        'ocr.detected': 'OCR detected {n} line(s) of text.',
        'ocr.noText': 'No text detected. Try recapturing the tooltip area.',
        'round.success': 'SUCCESS',
        'round.noMatch': 'No match',
        'cells': 'cells',
    },
    'zh-CN': {
        'title': 'POE2 混沌工匠',
        'connected': '已连接',
        'disconnected': '未连接',
        'tab.dashboard': '仪表盘',
        'tab.wizard': '向导',
        'tab.config': '配置',
        'status.title': '制作状态',
        'status.state': '状态：',
        'status.item': '物品：',
        'status.roll': '次数：',
        'status.totalRolls': '总次数：',
        'status.speed': '速度：',
        'status.duration': '耗时：',
        'btn.start': '开始',
        'btn.pause': '暂停',
        'btn.resume': '继续',
        'btn.stop': '停止',
        'btn.refresh': '刷新',
        'btn.back': '上一步',
        'btn.next': '下一步',
        'btn.capture': '捕获',
        'btn.add': '添加',
        'panel.snapshot': '游戏实时截图',
        'panel.ocrText': '识别的词缀文字',
        'panel.tooltip': '提示框',
        'panel.modStats': '词缀统计',
        'panel.history': '轮次历史',
        'table.mod': '词缀',
        'table.count': '次数',
        'table.min': '最小',
        'table.max': '最大',
        'table.avg': '平均',
        'table.prob': '概率%',
        'empty.noData': '暂无数据',
        'empty.noRounds': '暂无轮次',
        'empty.waiting': '等待开始制作...',
        'empty.noConfig': '未找到配置，请使用向导创建。',
        'empty.loading': '加载中...',
        'empty.noMods': '尚未添加词缀',
        'empty.noTargetMods': '未配置目标词缀',
        'state.idle': '空闲',
        'state.starting': '启动中...',
        'state.running': '运行中',
        'state.paused': '已暂停',
        'state.stopped': '已停止',
        'state.startingIn': '{n}秒后开始...',
        'wiz.step1.title': '第1步：配置',
        'wiz.step1.desc': '加载现有配置还是重新开始？',
        'wiz.step1.load': '加载现有',
        'wiz.step1.fresh': '重新开始',
        'wiz.step2.title': '第2步：背包网格',
        'wiz.step2.desc': '捕获背包网格的左上角和右下角（5行×12列）。',
        'wiz.topLeft': '左上角：',
        'wiz.bottomRight': '右下角：',
        'wiz.notSet': '未设置',
        'wiz.step3.title': '第3步：混沌石位置',
        'wiz.step3.desc': '捕获混沌石在仓库中的位置。',
        'wiz.chaosOrb': '混沌石：',
        'wiz.step4.title': '第4步：物品尺寸',
        'wiz.step4.desc': '物品占几个格子？',
        'wiz.width': '宽度（格）：',
        'wiz.height': '高度（格）：',
        'wiz.step5.title': '第5步：批量制作区域',
        'wiz.step5.desc': '指定工作台、待处理区域和结果区域的位置（行列坐标）。',
        'wiz.workbench': '工作台',
        'wiz.pendingArea': '待处理区域',
        'wiz.resultArea': '结果区域',
        'wiz.row': '行（0-4）：',
        'wiz.col': '列（0-11）：',
        'wiz.topLeftRow': '左上行：',
        'wiz.topLeftCol': '左上列：',
        'wiz.step6.title': '第6步：提示框区域',
        'wiz.step6.desc': '悬停在物品上显示提示框，然后捕获其角落。',
        'wiz.validateOCR': '验证OCR',
        'wiz.step7.title': '第7步：目标词缀',
        'wiz.step7.desc': '选择要搜索的词缀，格式：',
        'wiz.step7.format': '词缀名 最小值',
        'wiz.quickTemplate': '-- 快速模板 --',
        'wiz.minValue': '最小值',
        'wiz.addCustom': '自定义添加',
        'wiz.customPlaceholder': '如 life 80, fire-res 30',
        'wiz.step8.title': '第8步：选项与检查',
        'wiz.chaosPerRound': '每轮混沌石数量：',
        'wiz.ocrDebug': '启用OCR调试日志',
        'wiz.saveSnapshots': '保存所有快照',
        'wiz.review': '检查',
        'wiz.saveConfig': '保存配置',
        'wiz.saveAndStart': '保存并开始',
        'wiz.switchToGame': '请立即切换到游戏窗口！5...',
        'cfg.title': '当前配置',
        'cfg.reload': '重新加载',
        'cfg.editWizard': '在向导中编辑',
        'cfg.positions': '坐标位置',
        'cfg.chaosOrb': '混沌石',
        'cfg.bpTopLeft': '背包左上角',
        'cfg.bpBottomRight': '背包右下角',
        'cfg.item': '物品',
        'cfg.itemSize': '物品尺寸',
        'cfg.batchCrafting': '批量制作',
        'cfg.workbench': '工作台',
        'cfg.pendingArea': '待处理区域',
        'cfg.pendingSize': '待处理区域大小',
        'cfg.resultArea': '结果区域',
        'cfg.resultSize': '结果区域大小',
        'cfg.tooltip': '提示框',
        'cfg.offsetFromItem': '物品偏移',
        'cfg.size': '大小',
        'cfg.targetMods': '目标词缀',
        'cfg.options': '选项',
        'cfg.chaosPerRound': '每轮混沌石',
        'cfg.ocrDebug': 'OCR调试日志',
        'cfg.saveSnapshots': '保存所有快照',
        'cfg.enabled': '已启用',
        'cfg.disabled': '已禁用',
        'lang.ui': '界面',
        'lang.game': '游戏',
        'cfg.gameLanguage': '游戏语言',
        'toast.gameLangChanged': '游戏语言已更改，请重新添加目标词缀。',
        'toast.targetFound': '找到目标：{mod} = {value}！',
        'toast.sessionEnded': '制作会话已结束',
        'toast.startFailed': '启动制作失败',
        'toast.pauseFailed': '切换暂停失败',
        'toast.stopFailed': '停止制作失败',
        'toast.configLoaded': '配置已加载',
        'toast.freshConfig': '开始新配置',
        'toast.captureFailed': '捕获失败',
        'toast.selectMod': '请选择词缀模板',
        'toast.enterMin': '请输入最小值',
        'toast.enterMod': '请输入词缀字符串',
        'toast.invalidMod': '无效的词缀格式，试试：life 80',
        'toast.modAdded': '已添加：{desc}',
        'toast.modParseFailed': '解析词缀失败',
        'toast.configSaved': '配置已保存！',
        'toast.saveFailed': '保存配置失败',
        'toast.captureCorners': '请先捕获提示框角落',
        'toast.validationFailed': '验证失败',
        'toast.configLoadError': '加载配置出错。',
        'toast.noConfig': '未找到现有配置。',
        'toast.configLoadSuccess': '配置已加载！浏览步骤以检查/编辑。',
        'toast.saveError': '保存出错',
        'ocr.detected': 'OCR检测到 {n} 行文字。',
        'ocr.noText': '未检测到文字，请重新捕获提示框区域。',
        'round.success': '成功',
        'round.noMatch': '未匹配',
        'cells': '格',
    }
};

function t(key, params) {
    const dict = i18n[currentLang] || i18n['en'];
    let text = dict[key] || i18n['en'][key] || key;
    if (params) {
        for (const [k, v] of Object.entries(params)) {
            text = text.replace(`{${k}}`, v);
        }
    }
    return text;
}

function applyTranslations() {
    document.querySelectorAll('[data-i18n]').forEach(el => {
        const key = el.getAttribute('data-i18n');
        el.textContent = t(key);
    });
    document.querySelectorAll('[data-i18n-placeholder]').forEach(el => {
        const key = el.getAttribute('data-i18n-placeholder');
        el.placeholder = t(key);
    });
    document.title = t('title');
}

function setLanguage(lang) {
    currentLang = lang;
    localStorage.setItem('poe2crafter-lang', lang);
    document.getElementById('lang-select').value = lang;
    applyTranslations();
    // Re-render dynamic content
    if (modTemplates.length > 0) {
        modTemplates = [];
        initWizardModTemplates();
    }
}

function setGameLanguage(lang) {
    gameLang = lang;
    localStorage.setItem('poe2crafter-game-lang', lang);
    document.getElementById('game-lang-select').value = lang;
    // Re-render mod templates to show names in game language
    if (modTemplates.length > 0) {
        modTemplates = [];
        initWizardModTemplates();
    }
    showToast(t('toast.gameLangChanged'), 'info');
}

// ===== WebSocket Connection =====
let ws = null;
let wsReconnectTimer = null;
let craftStartTime = null;
let durationTimer = null;

function connectWebSocket() {
    const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:';
    const url = `${protocol}//${location.host}/ws`;

    ws = new WebSocket(url);

    ws.onopen = () => {
        document.getElementById('ws-status').textContent = t('connected');
        document.getElementById('ws-status').className = 'ws-connected';
        if (wsReconnectTimer) {
            clearTimeout(wsReconnectTimer);
            wsReconnectTimer = null;
        }
    };

    ws.onclose = () => {
        document.getElementById('ws-status').textContent = t('disconnected');
        document.getElementById('ws-status').className = 'ws-disconnected';
        wsReconnectTimer = setTimeout(connectWebSocket, 2000);
    };

    ws.onerror = () => {
        ws.close();
    };

    ws.onmessage = (event) => {
        try {
            const msg = JSON.parse(event.data);
            handleWSMessage(msg);
        } catch (e) {
            console.error('WS parse error:', e);
        }
    };
}

function handleWSMessage(msg) {
    switch (msg.type) {
        case 'state_change':
            updateCraftState(msg.data.state);
            break;
        case 'roll_attempted':
            updateRollInfo(msg.data);
            break;
        case 'tooltip_captured':
            refreshTooltipImage();
            break;
        case 'mods_tracked':
            updateModsTracked(msg.data);
            break;
        case 'target_found':
            handleTargetFound(msg.data);
            break;
        case 'item_started':
            updateItemStarted(msg.data);
            break;
        case 'item_completed':
            updateItemCompleted(msg.data);
            break;
        case 'snapshot_updated':
            refreshSnapshot();
            break;
        case 'session_ended':
            handleSessionEnded(msg.data);
            break;
        case 'craft_countdown':
            updateCraftCountdown(msg.data);
            break;
        case 'capture_countdown':
            updateCaptureCountdown(msg.data);
            break;
        case 'capture_result':
            handleCaptureResult(msg.data);
            break;
    }
}

// ===== State Management =====
function updateCraftState(state) {
    const el = document.getElementById('craft-state');
    el.className = 'value state-' + state;

    const btnStart = document.getElementById('btn-start');
    const btnPause = document.getElementById('btn-pause');
    const btnStop = document.getElementById('btn-stop');

    switch (state) {
        case 'idle':
        case 'stopped':
            el.textContent = t('state.' + state);
            btnStart.disabled = false;
            btnPause.disabled = true;
            btnStop.disabled = true;
            if (durationTimer) { clearInterval(durationTimer); durationTimer = null; }
            break;
        case 'countdown':
            btnStart.disabled = true;
            btnPause.disabled = true;
            btnStop.disabled = false;
            el.textContent = t('state.starting');
            el.className = 'value state-countdown';
            break;
        case 'running':
            el.textContent = t('state.running');
            btnStart.disabled = true;
            btnPause.disabled = false;
            btnStop.disabled = false;
            btnPause.textContent = t('btn.pause');
            if (!craftStartTime) craftStartTime = Date.now();
            startDurationTimer();
            break;
        case 'paused':
            el.textContent = t('state.paused');
            btnStart.disabled = true;
            btnPause.disabled = false;
            btnStop.disabled = false;
            btnPause.textContent = t('btn.resume');
            break;
    }
}

function updateCraftCountdown(data) {
    const el = document.getElementById('craft-state');
    el.textContent = t('state.startingIn', { n: data.secondsLeft });
    el.className = 'value state-countdown';
}

function startDurationTimer() {
    if (durationTimer) return;
    durationTimer = setInterval(() => {
        if (craftStartTime) {
            const elapsed = Math.floor((Date.now() - craftStartTime) / 1000);
            const min = Math.floor(elapsed / 60);
            const sec = elapsed % 60;
            document.getElementById('craft-duration').textContent =
                min > 0 ? `${min}m ${sec}s` : `${sec}s`;
        }
    }, 1000);
}

function updateRollInfo(data) {
    document.getElementById('craft-roll').textContent = `${data.attemptNum}/${data.maxAttempts}`;
    document.getElementById('craft-total').textContent = data.totalRolls;
    document.getElementById('craft-speed').textContent = `${data.rollsPerMin.toFixed(1)}/min`;
}

function updateItemStarted(data) {
    document.getElementById('craft-item').textContent = `#${data.itemNumber}`;
}

function updateItemCompleted(data) {
    const history = document.getElementById('round-history');
    if (history.querySelector('.empty-msg')) {
        history.innerHTML = '';
    }

    const badge = document.createElement('span');
    badge.className = data.success ? 'round-badge round-success' : 'round-badge round-fail';
    badge.textContent = `#${data.itemNumber}: ${data.success ? t('round.success') : t('round.noMatch')}`;
    history.appendChild(badge);
    history.scrollTop = history.scrollHeight;
}

function updateModsTracked(data) {
    const ocrEl = document.getElementById('ocr-text');
    if (data.ocrText) {
        ocrEl.textContent = data.ocrText;
    }

    if (data.modStats) {
        updateModStatsTable(data.modStats, data.totalRolls);
    }

    refreshTooltipImage();
}

function updateModStatsTable(modStats, totalRolls) {
    const tbody = document.getElementById('mod-stats-body');

    const entries = Object.entries(modStats).map(([name, stat]) => ({
        name: stat.ModName || name,
        count: stat.Count,
        min: stat.MinValue,
        max: stat.MaxValue,
        avg: stat.AvgValue,
        prob: totalRolls > 0 ? (stat.Count / totalRolls * 100) : 0
    }));

    entries.sort((a, b) => b.count - a.count);

    if (entries.length === 0) {
        tbody.innerHTML = `<tr><td colspan="6" class="empty-msg">${t('empty.noData')}</td></tr>`;
        return;
    }

    tbody.innerHTML = entries.map(e => `
        <tr>
            <td>${e.name}</td>
            <td>${e.count}</td>
            <td>${e.min}</td>
            <td>${e.max}</td>
            <td>${e.avg.toFixed(1)}</td>
            <td>${e.prob.toFixed(1)}%</td>
        </tr>
    `).join('');
}

function handleTargetFound(data) {
    showToast(t('toast.targetFound', { mod: data.modName, value: data.value }), 'success');
}

function handleSessionEnded(data) {
    craftStartTime = null;
    if (durationTimer) { clearInterval(durationTimer); durationTimer = null; }
    showToast(t('toast.sessionEnded'), 'info');
}

// ===== Crafting Controls =====
async function startCrafting() {
    try {
        craftStartTime = Date.now();
        document.getElementById('craft-total').textContent = '0';
        document.getElementById('craft-roll').textContent = '0/0';
        document.getElementById('craft-speed').textContent = '0/min';
        document.getElementById('craft-item').textContent = '#0';
        document.getElementById('round-history').innerHTML = `<span class="empty-msg">${t('empty.noRounds')}</span>`;
        document.getElementById('mod-stats-body').innerHTML = `<tr><td colspan="6" class="empty-msg">${t('empty.noData')}</td></tr>`;
        document.getElementById('ocr-text').textContent = t('state.starting');

        const resp = await fetch('/api/craft/start', { method: 'POST' });
        const data = await resp.json();
        if (data.error) {
            showToast(data.error, 'error');
        }
    } catch (e) {
        showToast(t('toast.startFailed'), 'error');
    }
}

async function pauseCrafting() {
    try {
        const resp = await fetch('/api/craft/pause', { method: 'POST' });
        const data = await resp.json();
        if (data.error) {
            showToast(data.error, 'error');
        }
    } catch (e) {
        showToast(t('toast.pauseFailed'), 'error');
    }
}

async function stopCrafting() {
    try {
        const resp = await fetch('/api/craft/stop', { method: 'POST' });
        const data = await resp.json();
        if (data.error) {
            showToast(data.error, 'error');
        }
    } catch (e) {
        showToast(t('toast.stopFailed'), 'error');
    }
}

// ===== Snapshot Refresh =====
function refreshSnapshot() {
    const img = document.getElementById('live-snapshot');
    img.src = `/api/snapshot/screen?t=${Date.now()}`;
}

function refreshTooltipImage() {
    const img = document.getElementById('tooltip-img');
    img.src = `/api/snapshot/current-tooltip?t=${Date.now()}`;
}

// ===== Tab Navigation =====
function switchTab(tabName) {
    document.querySelectorAll('.tab-btn').forEach(btn => {
        btn.classList.toggle('active', btn.dataset.tab === tabName);
    });

    document.querySelectorAll('.tab-content').forEach(content => {
        content.classList.toggle('active', content.id === `tab-${tabName}`);
    });

    if (tabName === 'config') loadAndShowConfig();
    if (tabName === 'wizard') initWizardModTemplates();
}

document.querySelectorAll('.tab-btn').forEach(btn => {
    btn.addEventListener('click', () => switchTab(btn.dataset.tab));
});

// ===== Config Tab =====
async function loadAndShowConfig() {
    const container = document.getElementById('config-display');
    try {
        const resp = await fetch('/api/config');
        if (!resp.ok) {
            container.innerHTML = `<p class="empty-msg">${t('empty.noConfig')}</p>`;
            return;
        }
        const cfg = await resp.json();
        container.innerHTML = formatConfigHTML(cfg);
    } catch (e) {
        container.innerHTML = `<p class="empty-msg">${t('toast.configLoadError')} ${e.message}</p>`;
    }
}

function formatConfigHTML(cfg) {
    function pixelToCell(pos) {
        if (!pos || !cfg.BackpackTopLeft || !cfg.BackpackBottomRight) return '';
        if (cfg.BackpackTopLeft.X === 0 && cfg.BackpackBottomRight.X === 0) return '';
        const cellW = (cfg.BackpackBottomRight.X - cfg.BackpackTopLeft.X) / 12;
        const cellH = (cfg.BackpackBottomRight.Y - cfg.BackpackTopLeft.Y) / 5;
        if (cellW <= 0 || cellH <= 0) return '';
        const col = Math.round((pos.X - cfg.BackpackTopLeft.X - cellW / 2) / cellW);
        const row = Math.round((pos.Y - cfg.BackpackTopLeft.Y - cellH / 2) / cellH);
        return ` (row ${Math.max(0, row)}, col ${Math.max(0, col)})`;
    }

    function row(label, value) {
        return `<div class="config-row"><span class="cfg-label">${label}</span><span class="cfg-value">${value}</span></div>`;
    }

    let html = '';

    html += `<div class="config-section"><h3>${t('cfg.positions')}</h3>`;
    html += row(t('cfg.chaosOrb'), `(${cfg.ChaosPos?.X || 0}, ${cfg.ChaosPos?.Y || 0})`);
    html += row(t('cfg.bpTopLeft'), `(${cfg.BackpackTopLeft?.X || 0}, ${cfg.BackpackTopLeft?.Y || 0})`);
    html += row(t('cfg.bpBottomRight'), `(${cfg.BackpackBottomRight?.X || 0}, ${cfg.BackpackBottomRight?.Y || 0})`);
    html += '</div>';

    html += `<div class="config-section"><h3>${t('cfg.item')}</h3>`;
    html += row(t('cfg.itemSize'), `${cfg.ItemWidth || 1} x ${cfg.ItemHeight || 1} ${t('cells')}`);
    html += '</div>';

    html += `<div class="config-section"><h3>${t('cfg.batchCrafting')}</h3>`;
    const wbPos = cfg.WorkbenchTopLeft;
    html += row(t('cfg.workbench'), `(${wbPos?.X || 0}, ${wbPos?.Y || 0})${pixelToCell(wbPos)}`);
    const pendPos = cfg.PendingAreaTopLeft;
    html += row(t('cfg.pendingArea'), `(${pendPos?.X || 0}, ${pendPos?.Y || 0})${pixelToCell(pendPos)}`);
    html += row(t('cfg.pendingSize'), `${cfg.PendingAreaWidth || 0} x ${cfg.PendingAreaHeight || 0} ${t('cells')}`);
    const resPos = cfg.ResultAreaTopLeft;
    html += row(t('cfg.resultArea'), `(${resPos?.X || 0}, ${resPos?.Y || 0})${pixelToCell(resPos)}`);
    html += row(t('cfg.resultSize'), `${cfg.ResultAreaWidth || 0} x ${cfg.ResultAreaHeight || 0} ${t('cells')}`);
    html += '</div>';

    html += `<div class="config-section"><h3>${t('cfg.tooltip')}</h3>`;
    if (cfg.TooltipOffset) {
        html += row(t('cfg.offsetFromItem'), `(${cfg.TooltipOffset.X}, ${cfg.TooltipOffset.Y})`);
    }
    if (cfg.TooltipSize) {
        html += row(t('cfg.size'), `${cfg.TooltipSize.X} x ${cfg.TooltipSize.Y} px`);
    }
    html += '</div>';

    html += `<div class="config-section"><h3>${t('cfg.targetMods')}</h3>`;
    if (cfg.TargetMods && cfg.TargetMods.length > 0) {
        html += '<ul class="config-mod-list">';
        cfg.TargetMods.forEach((mod, i) => {
            html += `<li>${i + 1}. ${mod.Description}</li>`;
        });
        html += '</ul>';
    } else {
        html += `<span class="empty-msg">${t('empty.noTargetMods')}</span>`;
    }
    html += '</div>';

    html += `<div class="config-section"><h3>${t('cfg.options')}</h3>`;
    html += row(t('cfg.chaosPerRound'), cfg.ChaosPerRound || 10);
    html += row(t('cfg.gameLanguage'), cfg.GameLanguage === 'zh-CN' ? '简体中文' : 'English');
    html += row(t('cfg.ocrDebug'), cfg.Debug ? t('cfg.enabled') : t('cfg.disabled'));
    html += row(t('cfg.saveSnapshots'), cfg.SaveAllSnapshots ? t('cfg.enabled') : t('cfg.disabled'));
    html += '</div>';

    return html;
}

// ===== Wizard =====
let wizardStep = 1;
let wizardConfig = {
    ChaosPos: { X: 0, Y: 0 },
    ItemPos: { X: 0, Y: 0 },
    ItemWidth: 1,
    ItemHeight: 1,
    TooltipOffset: { X: 0, Y: 0 },
    TooltipSize: { X: 0, Y: 0 },
    TooltipRect: { Min: { X: 0, Y: 0 }, Max: { X: 0, Y: 0 } },
    BackpackTopLeft: { X: 0, Y: 0 },
    BackpackBottomRight: { X: 0, Y: 0 },
    WorkbenchTopLeft: { X: 0, Y: 0 },
    PendingAreaTopLeft: { X: 0, Y: 0 },
    PendingAreaWidth: 4,
    PendingAreaHeight: 5,
    ResultAreaTopLeft: { X: 0, Y: 0 },
    ResultAreaWidth: 4,
    ResultAreaHeight: 5,
    UseBatchMode: true,
    TargetMods: [],
    ChaosPerRound: 10,
    Delay: 75000000,
    Debug: false,
    SaveAllSnapshots: false,
    GameLanguage: 'en'
};

let modTemplates = [];

async function initWizardModTemplates() {
    if (modTemplates.length > 0) return;
    try {
        const resp = await fetch('/api/mod-templates');
        modTemplates = await resp.json();
        const select = document.getElementById('wiz-mod-template');
        select.innerHTML = `<option value="">${t('wiz.quickTemplate')}</option>`;
        modTemplates.forEach(tmpl => {
            const opt = document.createElement('option');
            opt.value = tmpl.key;
            const name = (gameLang === 'zh-CN' && tmpl.name_zh) ? tmpl.name_zh : tmpl.name;
            opt.textContent = `${name} (${tmpl.example})`;
            select.appendChild(opt);
        });
    } catch (e) {
        console.error('Failed to load mod templates:', e);
    }
}

function wizardGoTo(step) {
    wizardStep = step;
    document.querySelectorAll('.wizard-step').forEach(s => s.classList.remove('active'));
    document.getElementById(`wizard-step-${step}`).classList.add('active');

    document.querySelectorAll('.step-dot').forEach(dot => {
        const dotStep = parseInt(dot.dataset.step);
        dot.classList.toggle('active', dotStep === step);
        dot.classList.toggle('done', dotStep < step);
    });

    if (step === 8) updateWizardReview();
}

function wizardNext() {
    gatherWizardStepData(wizardStep);
    if (wizardStep < 8) wizardGoTo(wizardStep + 1);
}

function wizardPrev() {
    if (wizardStep > 1) wizardGoTo(wizardStep - 1);
}

async function wizardLoadConfig() {
    try {
        const resp = await fetch('/api/config');
        if (!resp.ok) {
            document.getElementById('wizard-config-status').innerHTML =
                `<span style="color: var(--danger)">${t('toast.noConfig')}</span>`;
            return;
        }
        const cfg = await resp.json();
        wizardConfig = cfg;
        populateWizardFromConfig(cfg);
        document.getElementById('wizard-config-status').innerHTML =
            `<span style="color: var(--success)">${t('toast.configLoadSuccess')}</span>`;
        showToast(t('toast.configLoaded'), 'success');
    } catch (e) {
        document.getElementById('wizard-config-status').innerHTML =
            `<span style="color: var(--danger)">${t('toast.configLoadError')}</span>`;
    }
}

function wizardFresh() {
    wizardConfig = {
        ChaosPos: { X: 0, Y: 0 },
        ItemPos: { X: 0, Y: 0 },
        ItemWidth: 1,
        ItemHeight: 1,
        TooltipOffset: { X: 0, Y: 0 },
        TooltipSize: { X: 0, Y: 0 },
        TooltipRect: { Min: { X: 0, Y: 0 }, Max: { X: 0, Y: 0 } },
        BackpackTopLeft: { X: 0, Y: 0 },
        BackpackBottomRight: { X: 0, Y: 0 },
        WorkbenchTopLeft: { X: 0, Y: 0 },
        PendingAreaTopLeft: { X: 0, Y: 0 },
        PendingAreaWidth: 4,
        PendingAreaHeight: 5,
        ResultAreaTopLeft: { X: 0, Y: 0 },
        ResultAreaWidth: 4,
        ResultAreaHeight: 5,
        UseBatchMode: true,
        TargetMods: [],
        ChaosPerRound: 10,
        Delay: 75000000,
        Debug: false,
        SaveAllSnapshots: false,
        GameLanguage: gameLang
    };
    wizardGoTo(2);
    showToast(t('toast.freshConfig'), 'info');
}

function populateWizardFromConfig(cfg) {
    // Sync game language from loaded config
    if (cfg.GameLanguage) {
        gameLang = cfg.GameLanguage;
        localStorage.setItem('poe2crafter-game-lang', gameLang);
        document.getElementById('game-lang-select').value = gameLang;
    }

    if (cfg.BackpackTopLeft) {
        document.getElementById('wiz-grid-tl').textContent = `(${cfg.BackpackTopLeft.X}, ${cfg.BackpackTopLeft.Y})`;
    }
    if (cfg.BackpackBottomRight) {
        document.getElementById('wiz-grid-br').textContent = `(${cfg.BackpackBottomRight.X}, ${cfg.BackpackBottomRight.Y})`;
    }

    if (cfg.ChaosPos) {
        document.getElementById('wiz-chaos').textContent = `(${cfg.ChaosPos.X}, ${cfg.ChaosPos.Y})`;
    }

    if (cfg.ItemWidth) document.getElementById('wiz-item-width').value = cfg.ItemWidth;
    if (cfg.ItemHeight) document.getElementById('wiz-item-height').value = cfg.ItemHeight;

    if (cfg.BackpackTopLeft && cfg.BackpackBottomRight && cfg.BackpackTopLeft.X !== 0) {
        const cellW = (cfg.BackpackBottomRight.X - cfg.BackpackTopLeft.X) / 12;
        const cellH = (cfg.BackpackBottomRight.Y - cfg.BackpackTopLeft.Y) / 5;

        if (cfg.WorkbenchTopLeft && cfg.WorkbenchTopLeft.X !== 0) {
            const wbCol = Math.round((cfg.WorkbenchTopLeft.X - cfg.BackpackTopLeft.X - cellW / 2) / cellW);
            const wbRow = Math.round((cfg.WorkbenchTopLeft.Y - cfg.BackpackTopLeft.Y - cellH / 2) / cellH);
            document.getElementById('wiz-wb-row').value = Math.max(0, wbRow);
            document.getElementById('wiz-wb-col').value = Math.max(0, wbCol);
        }

        if (cfg.PendingAreaTopLeft && cfg.PendingAreaTopLeft.X !== 0) {
            const pCol = Math.round((cfg.PendingAreaTopLeft.X - cfg.BackpackTopLeft.X - cellW / 2) / cellW);
            const pRow = Math.round((cfg.PendingAreaTopLeft.Y - cfg.BackpackTopLeft.Y - cellH / 2) / cellH);
            document.getElementById('wiz-pend-row').value = Math.max(0, pRow);
            document.getElementById('wiz-pend-col').value = Math.max(0, pCol);
        }

        if (cfg.ResultAreaTopLeft && cfg.ResultAreaTopLeft.X !== 0) {
            const rCol = Math.round((cfg.ResultAreaTopLeft.X - cfg.BackpackTopLeft.X - cellW / 2) / cellW);
            const rRow = Math.round((cfg.ResultAreaTopLeft.Y - cfg.BackpackTopLeft.Y - cellH / 2) / cellH);
            document.getElementById('wiz-res-row').value = Math.max(0, rRow);
            document.getElementById('wiz-res-col').value = Math.max(0, rCol);
        }
    }

    if (cfg.PendingAreaWidth) document.getElementById('wiz-pend-w').value = cfg.PendingAreaWidth;
    if (cfg.PendingAreaHeight) document.getElementById('wiz-pend-h').value = cfg.PendingAreaHeight;
    if (cfg.ResultAreaWidth) document.getElementById('wiz-res-w').value = cfg.ResultAreaWidth;
    if (cfg.ResultAreaHeight) document.getElementById('wiz-res-h').value = cfg.ResultAreaHeight;

    if (cfg.TooltipRect && cfg.TooltipRect.Min) {
        document.getElementById('wiz-tooltip-tl').textContent =
            `(${cfg.TooltipRect.Min.X}, ${cfg.TooltipRect.Min.Y})`;
        document.getElementById('wiz-tooltip-br').textContent =
            `(${cfg.TooltipRect.Max.X}, ${cfg.TooltipRect.Max.Y})`;
    }

    updateWizardModList();

    if (cfg.ChaosPerRound) document.getElementById('wiz-chaos-per-round').value = cfg.ChaosPerRound;
    document.getElementById('wiz-debug').checked = cfg.Debug || false;
    document.getElementById('wiz-snapshots').checked = cfg.SaveAllSnapshots || false;
}

function gatherWizardStepData(step) {
    switch (step) {
        case 4:
            wizardConfig.ItemWidth = parseInt(document.getElementById('wiz-item-width').value) || 1;
            wizardConfig.ItemHeight = parseInt(document.getElementById('wiz-item-height').value) || 1;
            break;
        case 5:
            wizardConfig._wbRow = parseInt(document.getElementById('wiz-wb-row').value) || 0;
            wizardConfig._wbCol = parseInt(document.getElementById('wiz-wb-col').value) || 0;
            wizardConfig._pendRow = parseInt(document.getElementById('wiz-pend-row').value) || 0;
            wizardConfig._pendCol = parseInt(document.getElementById('wiz-pend-col').value) || 0;
            wizardConfig.PendingAreaWidth = parseInt(document.getElementById('wiz-pend-w').value) || 4;
            wizardConfig.PendingAreaHeight = parseInt(document.getElementById('wiz-pend-h').value) || 5;
            wizardConfig._resRow = parseInt(document.getElementById('wiz-res-row').value) || 0;
            wizardConfig._resCol = parseInt(document.getElementById('wiz-res-col').value) || 0;
            wizardConfig.ResultAreaWidth = parseInt(document.getElementById('wiz-res-w').value) || 4;
            wizardConfig.ResultAreaHeight = parseInt(document.getElementById('wiz-res-h').value) || 5;
            break;
        case 8:
            wizardConfig.ChaosPerRound = parseInt(document.getElementById('wiz-chaos-per-round').value) || 10;
            wizardConfig.Debug = document.getElementById('wiz-debug').checked;
            wizardConfig.SaveAllSnapshots = document.getElementById('wiz-snapshots').checked;
            break;
    }
}

async function wizardCapture(field) {
    try {
        const resp = await fetch('/api/wizard/capture', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ field: field })
        });

        if (!resp.ok) {
            showToast(t('toast.captureFailed'), 'error');
            return;
        }

        const countdown = document.getElementById('capture-countdown');
        countdown.classList.remove('hidden');
        countdown.textContent = t('wiz.switchToGame');
    } catch (e) {
        showToast(t('toast.captureFailed') + ': ' + e.message, 'error');
    }
}

function updateCaptureCountdown(data) {
    const countdown = document.getElementById('capture-countdown');
    countdown.classList.remove('hidden');
    countdown.textContent = `${data.secondsLeft}...`;
}

function handleCaptureResult(data) {
    const countdown = document.getElementById('capture-countdown');
    countdown.classList.add('hidden');

    const fieldMap = {
        'grid-tl': () => {
            wizardConfig.BackpackTopLeft = { X: data.x, Y: data.y };
            document.getElementById('wiz-grid-tl').textContent = `(${data.x}, ${data.y})`;
        },
        'grid-br': () => {
            wizardConfig.BackpackBottomRight = { X: data.x, Y: data.y };
            document.getElementById('wiz-grid-br').textContent = `(${data.x}, ${data.y})`;
        },
        'chaos': () => {
            wizardConfig.ChaosPos = { X: data.x, Y: data.y };
            document.getElementById('wiz-chaos').textContent = `(${data.x}, ${data.y})`;
        },
        'tooltip-tl': () => {
            wizardConfig.TooltipRect = wizardConfig.TooltipRect || { Min: {}, Max: {} };
            wizardConfig.TooltipRect.Min = { X: data.x, Y: data.y };
            document.getElementById('wiz-tooltip-tl').textContent = `(${data.x}, ${data.y})`;
        },
        'tooltip-br': () => {
            wizardConfig.TooltipRect = wizardConfig.TooltipRect || { Min: {}, Max: {} };
            wizardConfig.TooltipRect.Max = { X: data.x, Y: data.y };
            document.getElementById('wiz-tooltip-br').textContent = `(${data.x}, ${data.y})`;
        }
    };

    if (fieldMap[data.field]) {
        fieldMap[data.field]();
        showToast(`${t('btn.capture')} ${data.field}: (${data.x}, ${data.y})`, 'success');
    }
}

async function wizardValidateTooltip() {
    const tl = wizardConfig.TooltipRect?.Min;
    const br = wizardConfig.TooltipRect?.Max;

    if (!tl || !br || tl.X === 0 || br.X === 0) {
        showToast(t('toast.captureCorners'), 'error');
        return;
    }

    try {
        document.getElementById('btn-validate-tooltip').disabled = true;
        const resp = await fetch('/api/wizard/validate-tooltip', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ x1: tl.X, y1: tl.Y, x2: br.X, y2: br.Y, gameLanguage: gameLang })
        });

        const result = await resp.json();
        const el = document.getElementById('tooltip-validation-result');

        if (result.success) {
            el.innerHTML = `<div style="color: var(--success); margin-top: 10px;">
                ${t('ocr.detected', { n: result.validLines })}
                <pre style="font-size: 0.75rem; margin-top: 6px; max-height: 100px; overflow-y: auto;">${result.ocrText}</pre>
            </div>`;
        } else {
            el.innerHTML = `<div style="color: var(--danger); margin-top: 10px;">
                ${t('ocr.noText')}
                ${result.error ? '<br>Error: ' + result.error : ''}
            </div>`;
        }
    } catch (e) {
        showToast(t('toast.validationFailed') + ': ' + e.message, 'error');
    } finally {
        document.getElementById('btn-validate-tooltip').disabled = false;
    }
}

function wizardAddModFromTemplate() {
    const select = document.getElementById('wiz-mod-template');
    const valueInput = document.getElementById('wiz-mod-value');
    const key = select.value;
    const value = parseInt(valueInput.value);

    if (!key) {
        showToast(t('toast.selectMod'), 'error');
        return;
    }
    if (!value || value < 1) {
        showToast(t('toast.enterMin'), 'error');
        return;
    }

    const input = `${key} ${value}`;
    addModToWizard(input);
    select.value = '';
    valueInput.value = '';
}

function wizardAddModCustom() {
    const input = document.getElementById('wiz-mod-custom').value.trim();
    if (!input) {
        showToast(t('toast.enterMod'), 'error');
        return;
    }
    addModToWizard(input);
    document.getElementById('wiz-mod-custom').value = '';
}

async function addModToWizard(input) {
    try {
        const resp = await fetch('/api/wizard/parse-mod', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ input: input, gameLanguage: gameLang })
        });

        if (!resp.ok) {
            showToast(t('toast.invalidMod'), 'error');
            return;
        }

        const mod = await resp.json();
        wizardConfig.TargetMods = wizardConfig.TargetMods || [];
        wizardConfig.TargetMods.push(mod);
        updateWizardModList();
        showToast(t('toast.modAdded', { desc: mod.Description }), 'success');
    } catch (e) {
        showToast(t('toast.modParseFailed'), 'error');
    }
}

function wizardRemoveMod(index) {
    wizardConfig.TargetMods.splice(index, 1);
    updateWizardModList();
}

function updateWizardModList() {
    const list = document.getElementById('wiz-mod-list');
    if (!wizardConfig.TargetMods || wizardConfig.TargetMods.length === 0) {
        list.innerHTML = `<span class="empty-msg">${t('empty.noMods')}</span>`;
        return;
    }

    list.innerHTML = wizardConfig.TargetMods.map((mod, i) => `
        <div class="mod-entry">
            <span class="mod-desc">${mod.Description}</span>
            <button class="mod-remove" onclick="wizardRemoveMod(${i})">x</button>
        </div>
    `).join('');
}

function updateWizardReview() {
    gatherWizardStepData(8);

    const lines = [];
    lines.push(`${t('wiz.step2.title').replace(/^.*?：?/, '')}: (${wizardConfig.BackpackTopLeft.X}, ${wizardConfig.BackpackTopLeft.Y}) to (${wizardConfig.BackpackBottomRight.X}, ${wizardConfig.BackpackBottomRight.Y})`);
    lines.push(`${t('cfg.chaosOrb')}: (${wizardConfig.ChaosPos.X}, ${wizardConfig.ChaosPos.Y})`);
    lines.push(`${t('cfg.itemSize')}: ${wizardConfig.ItemWidth}x${wizardConfig.ItemHeight} ${t('cells')}`);
    lines.push('');
    lines.push(`${t('cfg.workbench')}: row ${wizardConfig._wbRow || 0}, col ${wizardConfig._wbCol || 0}`);
    lines.push(`${t('cfg.pendingArea')}: row ${wizardConfig._pendRow || 0}, col ${wizardConfig._pendCol || 0} [${wizardConfig.PendingAreaWidth}x${wizardConfig.PendingAreaHeight}]`);
    lines.push(`${t('cfg.resultArea')}: row ${wizardConfig._resRow || 0}, col ${wizardConfig._resCol || 0} [${wizardConfig.ResultAreaWidth}x${wizardConfig.ResultAreaHeight}]`);
    lines.push('');

    if (wizardConfig.TooltipRect && wizardConfig.TooltipRect.Min) {
        lines.push(`${t('cfg.tooltip')}: (${wizardConfig.TooltipRect.Min.X}, ${wizardConfig.TooltipRect.Min.Y}) to (${wizardConfig.TooltipRect.Max.X}, ${wizardConfig.TooltipRect.Max.Y})`);
    }
    lines.push('');
    lines.push(`${t('cfg.targetMods')}:`);
    if (wizardConfig.TargetMods && wizardConfig.TargetMods.length > 0) {
        wizardConfig.TargetMods.forEach((mod, i) => {
            lines.push(`  ${i + 1}. ${mod.Description}`);
        });
    } else {
        lines.push('  (none)');
    }
    lines.push('');
    lines.push(`${t('cfg.chaosPerRound')}: ${wizardConfig.ChaosPerRound}`);

    document.getElementById('wizard-review').textContent = lines.join('\n');
}

function computeCellCenter(row, col) {
    const tlX = wizardConfig.BackpackTopLeft.X;
    const tlY = wizardConfig.BackpackTopLeft.Y;
    const brX = wizardConfig.BackpackBottomRight.X;
    const brY = wizardConfig.BackpackBottomRight.Y;

    if (tlX === 0 && brX === 0) return { X: 0, Y: 0 };

    const cellW = (brX - tlX) / 12;
    const cellH = (brY - tlY) / 5;

    return {
        X: Math.round(tlX + col * cellW + cellW / 2),
        Y: Math.round(tlY + row * cellH + cellH / 2)
    };
}

function prepareConfigForSave() {
    gatherWizardStepData(wizardStep);

    const wbPos = computeCellCenter(wizardConfig._wbRow || 0, wizardConfig._wbCol || 0);
    wizardConfig.WorkbenchTopLeft = wbPos;
    wizardConfig.ItemPos = wbPos;

    const pendPos = computeCellCenter(wizardConfig._pendRow || 0, wizardConfig._pendCol || 0);
    wizardConfig.PendingAreaTopLeft = pendPos;

    const resPos = computeCellCenter(wizardConfig._resRow || 0, wizardConfig._resCol || 0);
    wizardConfig.ResultAreaTopLeft = resPos;

    if (wizardConfig.TooltipRect && wizardConfig.TooltipRect.Min && wbPos.X > 0) {
        wizardConfig.TooltipOffset = {
            X: wizardConfig.TooltipRect.Min.X - wbPos.X,
            Y: wizardConfig.TooltipRect.Min.Y - wbPos.Y
        };
        wizardConfig.TooltipSize = {
            X: wizardConfig.TooltipRect.Max.X - wizardConfig.TooltipRect.Min.X,
            Y: wizardConfig.TooltipRect.Max.Y - wizardConfig.TooltipRect.Min.Y
        };
    }

    wizardConfig.UseBatchMode = true;
    wizardConfig.GameLanguage = gameLang;

    const cfg = { ...wizardConfig };
    delete cfg._wbRow;
    delete cfg._wbCol;
    delete cfg._pendRow;
    delete cfg._pendCol;
    delete cfg._resRow;
    delete cfg._resCol;

    return cfg;
}

async function wizardSave() {
    const cfg = prepareConfigForSave();

    try {
        const resp = await fetch('/api/config', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(cfg)
        });

        if (resp.ok) {
            showToast(t('toast.configSaved'), 'success');
        } else {
            showToast(t('toast.saveFailed'), 'error');
        }
    } catch (e) {
        showToast(t('toast.saveError') + ': ' + e.message, 'error');
    }
}

async function wizardSaveAndStart() {
    await wizardSave();
    switchTab('dashboard');
    setTimeout(() => startCrafting(), 500);
}

// ===== Toast Notifications =====
function showToast(message, type = 'info') {
    const toast = document.createElement('div');
    toast.className = `toast toast-${type}`;
    toast.textContent = message;
    document.body.appendChild(toast);

    setTimeout(() => {
        if (toast.parentNode) toast.parentNode.removeChild(toast);
    }, 3000);
}

// ===== Init =====
document.getElementById('lang-select').value = currentLang;
document.getElementById('game-lang-select').value = gameLang;
applyTranslations();
connectWebSocket();
