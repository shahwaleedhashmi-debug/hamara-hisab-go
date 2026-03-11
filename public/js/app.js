'use strict';

// ── API BASE ──
const API = '/.netlify/functions';

// ── STATE ──
let appData = { shareholders: [], accounts: [] };
let allTxns  = [];
let txnPage  = 1;
const PER_PAGE = 30;
let currentFilter = 'all';
let currentSearch = '';

// wizard state
const wData = { amount: 0, des: '', ac: 0, incomeExpense: '', cashCredit: '', commonIndividual: '', step: 1 };

// ── FORMATTING ──
function fmt(n) {
  n = Math.abs(n);
  if (n >= 1e7) return 'Rs.' + (n / 1e7).toFixed(2) + ' Cr';
  if (n >= 1e5) return 'Rs.' + (n / 1e5).toFixed(2) + ' L';
  if (n >= 1e3) return 'Rs.' + (n / 1e3).toFixed(1) + 'K';
  return 'Rs.' + n.toFixed(0);
}
function fmtFull(n) {
  const abs = Math.abs(n);
  const s = abs.toFixed(0).replace(/\B(?=(\d{2})+(?!\d))/g, ',');
  return (n < 0 ? '-' : '') + 'Rs.' + s;
}
function initial(name) { return name ? name.charAt(0).toUpperCase() : '?'; }
function accountName(ac) {
  const a = appData.accounts.find(x => x.ac === ac);
  return a ? a.acname : '#' + ac;
}
function formatDate(ts) {
  if (!ts) return '';
  return ts.split(' ')[0];
}

// ── TOAST ──
let toastTimer;
function showToast(msg) {
  const el = document.getElementById('toast');
  el.textContent = msg;
  el.classList.add('show');
  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => el.classList.remove('show'), 2800);
}

// ── LOADING ──
function showLoading() { document.getElementById('loading-overlay').style.display = 'flex'; }
function hideLoading() { document.getElementById('loading-overlay').style.display = 'none'; }

// ── NAVIGATION ──
function nav(pageId) {
  document.querySelectorAll('.page').forEach(p => p.classList.remove('active'));
  document.querySelectorAll('.nav-item').forEach(n => n.classList.remove('active'));
  document.getElementById('page-' + pageId).classList.add('active');
  const navEl = document.querySelector(`.nav-item[data-page="${pageId}"]`);
  if (navEl) navEl.classList.add('active');

  if (pageId === 'home')     initHome();
  if (pageId === 'txns')     initTxns();
  if (pageId === 'reports')  initReports();
  if (pageId === 'deposit')  initDeposit();
  if (pageId === 'admin')    initAdmin();
}

// ── FETCH HELPERS ──
async function apiFetch(path, opts = {}) {
  const r = await fetch(API + path, {
    headers: { 'Content-Type': 'application/json' },
    ...opts,
  });
  if (!r.ok) throw new Error(`HTTP ${r.status}`);
  return r.json();
}

// ── BOOT ──
async function boot() {
  try {
    appData = await apiFetch('/data');
    allTxns = await apiFetch('/transactions');
    allTxns.sort((a, b) => (b.trs || 0) - (a.trs || 0));
    initHome();
    buildDepositForm();
  } catch (e) {
    console.error('Boot failed', e);
    showToast('Failed to load data. Check connection.');
  }
}

// ── HOME ──
function initHome() {
  let income = 0, expense = 0, cashIn = 0, cashOut = 0, bankDep = 0;
  allTxns.forEach(t => {
    if (t.ac === 100) { bankDep += t.amount; return; }
    if (t.income_expense === 'income') {
      income += t.amount;
      if (t.cash_credit === 'Cash') cashIn += t.amount;
    } else {
      expense += t.amount;
      if (t.cash_credit === 'Cash') cashOut += t.amount;
    }
  });
  const cash = cashIn - cashOut - bankDep;

  const el = document.getElementById('home-balance');
  el.textContent = fmtFull(Math.abs(cash));
  el.className = 'balance-amount ' + (cash >= 0 ? 'pos' : 'neg');

  document.getElementById('home-income').textContent  = fmt(income);
  document.getElementById('home-expense').textContent = fmt(expense);
  document.getElementById('home-txncount').textContent = allTxns.length + ' txns';
}

// ── ADD TRANSACTION WIZARD ──
function gotoAdd() {
  Object.assign(wData, { amount: 0, des: '', ac: 0, incomeExpense: '', cashCredit: '', commonIndividual: '' });
  nav('add');
  wShowStep(1);
}

function wShowStep(n) {
  wData.step = n;
  document.querySelectorAll('.w-step').forEach(s => s.style.display = 'none');
  const el = document.getElementById('w-step-' + n);
  if (el) el.style.display = 'block';
  // update progress
  for (let i = 1; i <= 8; i++) {
    const bar = document.getElementById('prog-' + i);
    if (!bar) continue;
    bar.className = 'progress-step ' + (i < n ? 'done' : i === n ? 'active' : '');
  }
  if (n === 4) buildAccountList();
  if (n === 5) buildShareList();
  if (n === 6) buildIndividualList();
  if (n === 7) buildConfirm();
  updateNextBtn(n);
}

function updateNextBtn(step) {
  const btn = document.getElementById('w-next-' + (step || wData.step));
  if (!btn) return;
  btn.disabled = !isStepValid(step || wData.step);
}

function isStepValid(n) {
  if (n === 1) return wData.amount > 0;
  if (n === 2) return !!wData.incomeExpense;
  if (n === 3) return !!wData.cashCredit;
  if (n === 4) return wData.ac > 0;
  if (n === 5) return !!wData.commonIndividual;
  if (n === 6) return !!wData.des.trim();
  return true;
}

function wNext(to) { if (isStepValid(wData.step)) wShowStep(to); }
function wBack(to) { wShowStep(to); }

function selectOpt(group, val, card) {
  document.querySelectorAll(`[data-group="${group}"]`).forEach(c => {
    c.classList.remove('selected', 'selected-success', 'selected-danger');
  });
  if (group === 'ie') {
    wData.incomeExpense = val;
    card.classList.add(val === 'income' ? 'selected-success' : 'selected-danger');
  } else if (group === 'cc') {
    wData.cashCredit = val;
    card.classList.add('selected');
  }
  updateNextBtn(wData.step);
}

function buildAccountList() {
  const ie = wData.incomeExpense;
  const list = appData.accounts.filter(a => {
    if (a.income_flag === 2) return true;
    if (ie === 'income')  return a.income_flag === 1;
    if (ie === 'expense') return a.income_flag === 0;
    return true;
  });
  const container = document.getElementById('account-list');
  container.innerHTML = list.map(a => `
    <div class="account-item ${wData.ac === a.ac ? 'selected' : ''}" onclick="selectAccount(${a.ac}, this)">
      <span class="account-name">${a.acname}</span>
      <span class="account-badge ${a.income_flag === 1 ? 'income' : a.income_flag === 0 ? 'expense' : 'special'}">
        ${a.income_flag === 1 ? 'Income' : a.income_flag === 0 ? 'Expense' : 'Special'}
      </span>
    </div>
  `).join('');
}

function selectAccount(ac, el) {
  wData.ac = ac;
  document.querySelectorAll('.account-item').forEach(c => c.classList.remove('selected'));
  el.classList.add('selected');
  updateNextBtn(wData.step);
}

function buildShareList() {
  // Step 5: choose split scheme (A, B) or individual (1-5)
  const opts = [
    { val: 'A', label: 'Scheme A', desc: 'Ammi/Jahanzeb/Waleed 25%, Alka/Memoona 12.5%' },
    { val: 'B', label: 'Scheme B', desc: 'Ammi 12.4%, Alka/Memoona 14.6%, Jahanzeb/Waleed 29.2%' },
    { val: 'C', label: 'Scheme C', desc: 'Jahanzeb 50%, Waleed 50%' },
  ].concat(appData.shareholders.map(sh => ({ val: String(sh.id), label: sh.name, desc: 'Individual' })));

  document.getElementById('share-list').innerHTML = opts.map(o => `
    <div class="option-card ${wData.commonIndividual === o.val ? 'selected' : ''}"
         data-group="ci" onclick="selectCI('${o.val}', this)">
      <div class="option-radio"></div>
      <div class="option-text">
        <div class="option-title">${o.label}</div>
        <div class="option-desc">${o.desc}</div>
      </div>
    </div>
  `).join('');
}

function selectCI(val, el) {
  wData.commonIndividual = val;
  document.querySelectorAll('[data-group="ci"]').forEach(c => c.classList.remove('selected'));
  el.classList.add('selected');
  updateNextBtn(wData.step);
}

function buildIndividualList() {
  // Step 6: description
  document.getElementById('w-desc-input').value = wData.des;
}

function buildConfirm() {
  const acName = accountName(wData.ac);
  const splits = calcSplit(wData.amount, wData.commonIndividual);

  document.getElementById('conf-amount').textContent = fmtFull(wData.amount);
  document.getElementById('conf-amount').className = 'confirm-amount ' + wData.incomeExpense;
  document.getElementById('conf-ac').textContent = acName;
  document.getElementById('conf-type').textContent = wData.incomeExpense.charAt(0).toUpperCase() + wData.incomeExpense.slice(1);
  document.getElementById('conf-cc').textContent = wData.cashCredit;
  document.getElementById('conf-des').textContent = wData.des || '—';
  document.getElementById('conf-ci').textContent = ciLabel(wData.commonIndividual);

  const tbody = document.getElementById('conf-splits');
  tbody.innerHTML = appData.shareholders.map(sh => `
    <tr>
      <td>${sh.name}</td>
      <td>${fmtFull(splits[sh.name.toLowerCase()] || 0)}</td>
    </tr>
  `).join('');
}

function ciLabel(ci) {
  if (ci === 'A') return 'Scheme A';
  if (ci === 'B') return 'Scheme B';
  if (ci === 'C') return 'Scheme C';
  const sh = appData.shareholders.find(s => String(s.id) === ci);
  return sh ? sh.name : ci;
}

function calcSplit(amount, ci) {
  const result = {};
  appData.shareholders.forEach(sh => result[sh.name.toLowerCase()] = 0);
  const r = v => Math.round(amount * v * 100) / 100;
  if (ci === 'A') {
    // Ammi 25%, Alka 12.5%, Jahanzeb 25%, Memoona 12.5%, Waleed 25%
    Object.assign(result, { ammi: r(0.25), alka: r(0.125), jahanzeb: r(0.25), memoona: r(0.125), waleed: r(0.25) });
  } else if (ci === 'B') {
    // Ammi 12.4%, Alka 14.6%, Jahanzeb 29.2%, Memoona 14.6%, Waleed 29.2%
    Object.assign(result, { ammi: r(0.124), alka: r(0.146), jahanzeb: r(0.292), memoona: r(0.146), waleed: r(0.292) });
  } else if (ci === 'C') {
    // Jahanzeb 50%, Waleed 50%
    Object.assign(result, { jahanzeb: r(0.5), waleed: r(0.5) });
  } else {
    const sh = appData.shareholders.find(s => String(s.id) === ci);
    if (sh) result[sh.name.toLowerCase()] = amount;
  }
  return result;
}

async function submitWizard() {
  const btn = document.getElementById('w-submit-btn');
  btn.disabled = true;
  btn.innerHTML = '<span class="spinner"></span>';

  try {
    const payload = {
      des: wData.des,
      amount: wData.amount,
      ac: wData.ac,
      income_expense: wData.incomeExpense,
      cash_credit: wData.cashCredit,
      common_individual: wData.commonIndividual,
    };
    const saved = await apiFetch('/transactions', { method: 'POST', body: JSON.stringify(payload) });
    allTxns.unshift(saved);
    showToast('✓ Transaction saved!');
    nav('home');
  } catch (e) {
    showToast('Error saving transaction');
    btn.disabled = false;
    btn.textContent = 'Submit';
  }
}

// ── TRANSACTIONS PAGE ──
function initTxns() {
  txnPage = 1;
  currentFilter = 'all';
  currentSearch = '';
  document.getElementById('txn-search').value = '';
  renderTxns();
}

function filterTxns() {
  return allTxns.filter(t => {
    const matchFilter =
      currentFilter === 'all' ||
      (currentFilter === 'income'  && t.income_expense === 'income') ||
      (currentFilter === 'expense' && t.income_expense === 'expense') ||
      (currentFilter === 'bank'    && t.ac === 100);
    const q = currentSearch.toLowerCase();
    const matchSearch = !q || (t.des || '').toLowerCase().includes(q) || accountName(t.ac).toLowerCase().includes(q);
    return matchFilter && matchSearch;
  });
}

function renderTxns() {
  const filtered = filterTxns();
  const total = filtered.length;
  const pages = Math.ceil(total / PER_PAGE) || 1;
  txnPage = Math.min(txnPage, pages);
  const slice = filtered.slice((txnPage - 1) * PER_PAGE, txnPage * PER_PAGE);

  const container = document.getElementById('txn-list');
  if (slice.length === 0) {
    container.innerHTML = `
      <div class="empty-state">
        <svg fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"/></svg>
        <h3>No transactions</h3>
        <p>Try a different filter or search term</p>
      </div>`;
  } else {
    container.innerHTML = slice.map(t => {
      const isInc = t.income_expense === 'income';
      const isBnk = t.ac === 100;
      const icon  = isInc ? '↑' : isBnk ? '🏦' : '↓';
      const cls   = isInc ? 'income' : isBnk ? 'bank' : 'expense';
      return `
        <div class="txn-item">
          <div class="txn-icon ${cls}">${icon}</div>
          <div class="txn-desc">
            <div class="txn-name">${t.des || accountName(t.ac)}</div>
            <div class="txn-meta">${accountName(t.ac)} · ${formatDate(t.tstamp)}</div>
          </div>
          <div class="txn-amount">
            <div class="amount ${cls}">${isInc ? '+' : '−'}${fmt(t.amount)}</div>
            <div class="cash-tag">${t.cash_credit || 'Cash'}</div>
          </div>
        </div>`;
    }).join('');
  }

  document.getElementById('txn-page-info').textContent = `Page ${txnPage} of ${pages}`;
  document.getElementById('txn-prev').disabled = txnPage <= 1;
  document.getElementById('txn-next').disabled = txnPage >= pages;
}

function setFilter(f, el) {
  currentFilter = f;
  document.querySelectorAll('.filter-pill').forEach(p => p.classList.remove('active'));
  el.classList.add('active');
  txnPage = 1;
  renderTxns();
}

function txnSearch(val) {
  currentSearch = val;
  txnPage = 1;
  renderTxns();
}

function txnChangePage(dir) {
  txnPage += dir;
  renderTxns();
  document.getElementById('txn-list').scrollIntoView({ behavior: 'smooth' });
}

// ── REPORTS ──
async function initReports() {
  const body = document.getElementById('reports-body');
  body.innerHTML = `<div style="text-align:center;padding:40px"><span class="spinner dark"></span></div>`;
  try {
    const r = await apiFetch('/reports');
    renderReports(r);
  } catch (e) {
    body.innerHTML = '<p style="padding:20px;color:var(--danger)">Failed to load reports</p>';
  }
}

function renderReports(r) {
  const body = document.getElementById('reports-body');
  const cash = r.cash_on_hand;

  const last10HTML = (r.last_10 || []).map(t => {
    const isInc = t.income_expense === 'income';
    return `
      <div class="txn-item" style="box-shadow:none;border-bottom:1px solid var(--gray-100);border-radius:0;padding:12px 0">
        <div class="txn-icon ${isInc ? 'income' : 'expense'}" style="width:32px;height:32px;border-radius:8px;font-size:0.9rem">${isInc ? '↑' : '↓'}</div>
        <div class="txn-desc">
          <div class="txn-name" style="font-size:0.82rem">${t.des || accountName(t.ac)}</div>
          <div class="txn-meta">${formatDate(t.tstamp)}</div>
        </div>
        <div class="txn-amount">
          <div class="amount ${isInc ? 'income' : 'expense'}" style="font-size:0.85rem">${isInc ? '+' : '−'}${fmt(t.amount)}</div>
        </div>
      </div>`;
  }).join('');

  const shHTML = (r.shareholders || []).map(sh => `
    <div class="sh-balance-row">
      <div class="sh-bal-avatar">${initial(sh.name)}</div>
      <div class="sh-bal-name">${sh.name.charAt(0).toUpperCase() + sh.name.slice(1)}</div>
      <div class="sh-bal-amount ${sh.balance >= 0 ? 'pos' : 'neg'}">${fmtFull(sh.balance)}</div>
    </div>`).join('');

  const yearHTML = (r.yearly_summaries || []).map(y => `
    <tr>
      <td><strong>${y.year}</strong></td>
      <td class="inc">+${fmt(y.income)}</td>
      <td class="exp">−${fmt(y.expense)}</td>
      <td class="net ${y.net >= 0 ? 'pos' : 'neg'}">${fmtFull(y.net)}</td>
    </tr>`).join('');

  body.innerHTML = `
    <div class="report-card">
      <div class="report-card-hdr">📊 Summary</div>
      <div class="report-card-body">
        <div class="summary-grid">
          <div class="summary-item">
            <div class="s-label">Income</div>
            <div class="s-value income">${fmt(r.total_income)}</div>
          </div>
          <div class="summary-item">
            <div class="s-label">Expense</div>
            <div class="s-value expense">${fmt(r.total_expense)}</div>
          </div>
          <div class="summary-item">
            <div class="s-label">Cash</div>
            <div class="s-value cash">${fmt(cash)}</div>
          </div>
        </div>
      </div>
    </div>

    <div class="report-card">
      <div class="report-card-hdr">👥 Shareholder Balances</div>
      <div class="report-card-body">
        <div class="sh-balance-list">${shHTML}</div>
      </div>
    </div>

    <div class="report-card">
      <div class="report-card-hdr">📅 Yearly Summary</div>
      <div class="report-card-body" style="padding:0">
        <table class="year-table">
          <thead><tr><th>Year</th><th>Income</th><th>Expense</th><th>Net</th></tr></thead>
          <tbody>${yearHTML}</tbody>
        </table>
      </div>
    </div>

    <div class="report-card">
      <div class="report-card-hdr">🕐 Last 10 Transactions</div>
      <div class="report-card-body" style="padding:0 16px">${last10HTML || '<div class="empty-state" style="padding:24px">No transactions</div>'}</div>
    </div>`;
}

// ── DEPOSIT ──
function buildDepositForm() {
  const container = document.getElementById('deposit-rows');
  container.innerHTML = appData.shareholders.map(sh => `
    <div class="dep-row">
      <div class="dep-avatar">${initial(sh.name)}</div>
      <div class="dep-name">${sh.name}</div>
      <input class="dep-input" id="dep-${sh.name.toLowerCase()}" type="number"
             inputmode="decimal" value="0" min="0" step="0.01"
             oninput="updateDepTotal()">
    </div>
  `).join('');
  updateDepTotal();
}

function initDeposit() { updateDepTotal(); }

function updateDepTotal() {
  let total = 0;
  appData.shareholders.forEach(sh => {
    const v = parseFloat(document.getElementById('dep-' + sh.name.toLowerCase())?.value || 0) || 0;
    total += v;
  });
  document.getElementById('dep-total').textContent = fmtFull(total);
}

async function saveDeposit() {
  const payload = { total: 0 };
  appData.shareholders.forEach(sh => {
    const v = parseFloat(document.getElementById('dep-' + sh.name.toLowerCase())?.value || 0) || 0;
    payload[sh.name.toLowerCase()] = v;
    payload.total += v;
  });

  if (payload.total === 0) { showToast('Enter at least one amount'); return; }

  const btn = document.getElementById('dep-save-btn');
  btn.disabled = true;
  btn.innerHTML = '<span class="spinner"></span>';

  try {
    await apiFetch('/deposit', { method: 'POST', body: JSON.stringify(payload) });
    showToast('✓ Deposit saved!');
    appData.shareholders.forEach(sh => {
      const el = document.getElementById('dep-' + sh.name.toLowerCase());
      if (el) el.value = '0';
    });
    updateDepTotal();
  } catch (e) {
    showToast('Error saving deposit');
  } finally {
    btn.disabled = false;
    btn.textContent = 'Save Deposit';
  }
}

// ── ADMIN ──
function initAdmin() {
  let income = 0, expense = 0, cashIn = 0, cashOut = 0, bankDep = 0;
  allTxns.forEach(t => {
    if (t.ac === 100) { bankDep += t.amount; return; }
    if (t.income_expense === 'income') {
      income += t.amount;
      if (t.cash_credit === 'Cash') cashIn += t.amount;
    } else {
      expense += t.amount;
      if (t.cash_credit === 'Cash') cashOut += t.amount;
    }
  });
  document.getElementById('admin-income').textContent  = fmt(income);
  document.getElementById('admin-expense').textContent = fmt(expense);
  document.getElementById('admin-cash').textContent    = fmt(cashIn - cashOut - bankDep);
  document.getElementById('admin-count').textContent   = allTxns.length;
}

async function refreshTxns() {
  showLoading();
  try {
    allTxns = await apiFetch('/transactions');
    allTxns.sort((a, b) => (b.trs || 0) - (a.trs || 0));
    initHome();
    showToast('✓ Data refreshed');
  } catch (e) {
    showToast('Refresh failed');
  } finally {
    hideLoading();
  }
}

function exportCSV() {
  const rows = [['Date','Description','Account','Type','Cash/Credit','Amount','Ammi','Alka','Jahanzeb','Memoona','Waleed']];
  allTxns.forEach(t => {
    rows.push([t.tstamp, t.des, accountName(t.ac), t.income_expense, t.cash_credit,
               t.amount, t.ammi, t.alka, t.jahanzeb, t.memoona, t.waleed]);
  });
  const csv = rows.map(r => r.map(v => `"${String(v||'').replace(/"/g,'""')}"`).join(',')).join('\n');
  const blob = new Blob([csv], { type: 'text/csv' });
  const a = document.createElement('a');
  a.href = URL.createObjectURL(blob);
  a.download = 'hamara-hisab-' + new Date().toISOString().split('T')[0] + '.csv';
  a.click();
  showToast('✓ CSV downloaded');
}

// ── INIT ──
document.addEventListener('DOMContentLoaded', () => {
  // nav item clicks
  document.querySelectorAll('.nav-item').forEach(item => {
    item.addEventListener('click', () => nav(item.dataset.page));
  });
  // start
  nav('home');
  boot();
});
