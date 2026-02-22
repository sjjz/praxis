import React, { useMemo, useState } from 'https://esm.sh/react@18.3.1';
import { createRoot } from 'https://esm.sh/react-dom@18.3.1/client';

const API_BASE = (window.PRAXIS_API_BASE || localStorage.getItem('praxis_api_base') || 'http://localhost:8080').replace(/\/$/, '');

async function api(path, options = {}) {
  const res = await fetch(`${API_BASE}/v1${path}`, {
    headers: { 'content-type': 'application/json', ...(options.headers || {}) },
    ...options,
  });
  if (!res.ok) {
    let detail = '';
    try {
      const payload = await res.json();
      detail = payload?.error || JSON.stringify(payload);
    } catch {
      detail = await res.text();
    }
    throw new Error(detail || `Request failed (${res.status})`);
  }
  if (res.status === 204) return null;
  return res.json();
}

function toISO(localDateTime) {
  return new Date(localDateTime).toISOString();
}

function optionalNumber(value) {
  if (value === '' || value == null) return null;
  const n = Number(value);
  return Number.isFinite(n) ? n : null;
}

function App() {
  const [windowSize, setWindowSize] = useState('7d');
  const [status, setStatus] = useState('');
  const [error, setError] = useState(false);
  const [trends, setTrends] = useState([]);
  const [summaries, setSummaries] = useState('No data loaded yet.');
  const [fabOpen, setFabOpen] = useState(false);
  const [showMood, setShowMood] = useState(false);
  const [showFood, setShowFood] = useState(false);

  const nowLocal = useMemo(() => new Date().toISOString().slice(0, 16), []);

  async function loadTrends(w) {
    setWindowSize(w);
    setError(false);
    setStatus(`Loading ${w} trends...`);
    try {
      const data = await api(`/trends?window=${w}`);
      const items = data?.items || [];
      setTrends(items);
      setStatus(`Loaded ${items.length} day(s) of trends.`);
    } catch (err) {
      setError(true);
      setStatus(err.message);
    }
  }

  async function submitMood(e) {
    e.preventDefault();
    const form = new FormData(e.currentTarget);
    const payload = {
      mood_type: form.get('mood_type'),
      quality: Number(form.get('quality')),
      timestamp: toISO(form.get('timestamp')),
      note: form.get('note') || null,
    };

    try {
      await api('/mood-checkins', { method: 'POST', body: JSON.stringify(payload) });
      setStatus('Mood check-in saved.');
      setError(false);
      e.currentTarget.reset();
      setShowMood(false);
      await loadTrends(windowSize);
    } catch (err) {
      setError(true);
      setStatus(err.message);
    }
  }

  async function submitFood(e) {
    e.preventDefault();
    const form = new FormData(e.currentTarget);
    const payload = {
      timestamp: toISO(form.get('timestamp')),
      meal_tag: form.get('meal_tag') || null,
      calories: optionalNumber(form.get('calories')),
      protein_g: optionalNumber(form.get('protein_g')),
      fiber_g: optionalNumber(form.get('fiber_g')),
      added_sugar_g: optionalNumber(form.get('added_sugar_g')),
      carbs_g: optionalNumber(form.get('carbs_g')),
    };

    try {
      await api('/nutrition-entries', { method: 'POST', body: JSON.stringify(payload) });
      setStatus('Nutrition entry saved.');
      setError(false);
      e.currentTarget.reset();
      setShowFood(false);
      await loadTrends(windowSize);
    } catch (err) {
      setError(true);
      setStatus(err.message);
    }
  }

  async function loadHistory(e) {
    e.preventDefault();
    const form = new FormData(e.currentTarget);
    const from = form.get('from');
    const to = form.get('to');
    try {
      const data = await api(`/daily-summaries?from=${from}&to=${to}`);
      setSummaries(JSON.stringify(data, null, 2));
      setError(false);
      setStatus(`Loaded ${data.items.length} day(s) of summaries.`);
    } catch (err) {
      setError(true);
      setStatus(err.message);
    }
  }

  React.useEffect(() => {
    loadTrends('7d');
  }, []);

  return (
    React.createElement('main', { className: 'container' },
      React.createElement('header', { className: 'header' },
        React.createElement('h1', null, 'Praxis'),
        React.createElement('p', null, 'Your daily trends first.')),

      React.createElement('section', { className: 'card' },
        React.createElement('div', { className: 'row-between' },
          React.createElement('h2', null, 'Daily Trends'),
          React.createElement('div', { className: 'btn-row' },
            React.createElement('button', { className: 'secondary', onClick: () => loadTrends('7d') }, '7d'),
            React.createElement('button', { className: 'secondary', onClick: () => loadTrends('30d') }, '30d'))),
        React.createElement('div', { className: 'trend-list' },
          trends.length === 0 ? React.createElement('p', null, 'No trend data yet.') : trends.map((item) =>
            React.createElement('article', { key: item.date, className: 'trend-day' },
              React.createElement('strong', null, item.date),
              React.createElement('div', null, `Mood score avg: ${item.mood_score_avg ?? 'n/a'}`),
              React.createElement('div', null, `Calories total: ${item.calories_total ?? 'n/a'}`),
              React.createElement('div', null, `Protein total: ${item.protein_total_g ?? 'n/a'}g`),
              React.createElement('div', null, `Fiber total: ${item.fiber_total_g ?? 'n/a'}g`),
              React.createElement('div', null, `Added sugar total: ${item.added_sugar_total_g ?? 'n/a'}g`),
            )))),

      React.createElement('section', { className: 'card' },
        React.createElement('h2', null, 'Aggregate History'),
        React.createElement('form', { className: 'form-grid inline', onSubmit: loadHistory },
          React.createElement('label', null, 'From', React.createElement('input', { type: 'date', name: 'from', required: true })),
          React.createElement('label', null, 'To', React.createElement('input', { type: 'date', name: 'to', required: true })),
          React.createElement('button', { type: 'submit' }, 'Load Daily Summary')),
        React.createElement('pre', { className: 'output' }, summaries)),

      React.createElement('p', { className: `status ${error ? 'error' : ''}` }, status),
      React.createElement('p', { className: 'api-base' }, `API: ${API_BASE}`),

      React.createElement('button', {
        className: 'fab',
        onClick: () => setFabOpen((v) => !v),
        'aria-label': 'Add entry',
      }, '＋'),

      React.createElement('div', { className: `fab-menu ${fabOpen ? '' : 'hidden'}` },
        React.createElement('button', { className: 'fab-item', onClick: () => { setFabOpen(false); setShowFood(true); } }, '🍽️ Food'),
        React.createElement('button', { className: 'fab-item', onClick: () => { setFabOpen(false); setShowMood(true); } }, '🙂 Mood')),

      showMood && React.createElement('div', { className: 'modal-backdrop' },
        React.createElement('div', { className: 'modal' },
          React.createElement('form', { className: 'form-grid', onSubmit: submitMood },
            React.createElement('h2', null, 'Record Mood'),
            React.createElement('label', null, 'Mood Type',
              React.createElement('select', { name: 'mood_type', required: true },
                React.createElement('option', { value: 'energy' }, 'Energy'),
                React.createElement('option', { value: 'fog_heaviness' }, 'Fog Heaviness'),
                React.createElement('option', { value: 'stress' }, 'Stress'),
                React.createElement('option', { value: 'motivation' }, 'Motivation'))),
            React.createElement('label', null, 'Quality (1-5)', React.createElement('input', { type: 'number', name: 'quality', min: '1', max: '5', required: true })),
            React.createElement('label', null, 'Timestamp', React.createElement('input', { type: 'datetime-local', name: 'timestamp', defaultValue: nowLocal, required: true })),
            React.createElement('label', null, 'Note', React.createElement('textarea', { name: 'note', rows: '2', placeholder: 'Optional' })),
            React.createElement('div', { className: 'btn-row' },
              React.createElement('button', { type: 'submit' }, 'Save Mood'),
              React.createElement('button', { type: 'button', className: 'secondary', onClick: () => setShowMood(false) }, 'Cancel'))))),

      showFood && React.createElement('div', { className: 'modal-backdrop' },
        React.createElement('div', { className: 'modal' },
          React.createElement('form', { className: 'form-grid', onSubmit: submitFood },
            React.createElement('h2', null, 'Record Food'),
            React.createElement('label', null, 'Timestamp', React.createElement('input', { type: 'datetime-local', name: 'timestamp', defaultValue: nowLocal, required: true })),
            React.createElement('label', null, 'Meal',
              React.createElement('select', { name: 'meal_tag' },
                React.createElement('option', { value: '' }, '(none)'),
                React.createElement('option', { value: 'breakfast' }, 'Breakfast'),
                React.createElement('option', { value: 'lunch' }, 'Lunch'),
                React.createElement('option', { value: 'dinner' }, 'Dinner'),
                React.createElement('option', { value: 'snack' }, 'Snack'),
                React.createElement('option', { value: 'other' }, 'Other'))),
            React.createElement('label', null, 'Calories', React.createElement('input', { type: 'number', step: '0.1', name: 'calories' })),
            React.createElement('label', null, 'Protein (g)', React.createElement('input', { type: 'number', step: '0.1', name: 'protein_g' })),
            React.createElement('label', null, 'Fiber (g)', React.createElement('input', { type: 'number', step: '0.1', name: 'fiber_g' })),
            React.createElement('label', null, 'Added Sugar (g)', React.createElement('input', { type: 'number', step: '0.1', name: 'added_sugar_g' })),
            React.createElement('label', null, 'Carbs (g)', React.createElement('input', { type: 'number', step: '0.1', name: 'carbs_g' })),
            React.createElement('div', { className: 'btn-row' },
              React.createElement('button', { type: 'submit' }, 'Save Food'),
              React.createElement('button', { type: 'button', className: 'secondary', onClick: () => setShowFood(false) }, 'Cancel')))))),
    )
  );
}

createRoot(document.getElementById('root')).render(React.createElement(App));
