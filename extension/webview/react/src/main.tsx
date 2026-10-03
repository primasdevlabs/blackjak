import React from 'react';
import ReactDOM from 'react-dom/client';
import App from './App';
import { BrowserShell } from './BrowserShell';
import { SettingsPage } from './settings/SettingsPage';
import { ActivityPanel } from './components/activity-panel';
import './App.css';

const w = window as any;
const isEmbedded = typeof w.vscode !== 'undefined' && typeof w.vscode.postMessage === 'function';
const injectedRoute = String(w.__BLACKJAK_ROUTE__ || '');
const hashRoute = (window.location.hash || '').replace(/^#\/?/, '');

function resolveRoute(): 'chat' | 'settings' | 'activity' {
  const r = injectedRoute || hashRoute;
  if (r.startsWith('settings')) return 'settings';
  if (r.startsWith('activity')) return 'activity';
  return 'chat';
}

const route = resolveRoute();

function Root() {
  // Dedicated extension panels inject __BLACKJAK_ROUTE__ and render a single surface.
  if (isEmbedded && route === 'settings') {
    return (
      <SettingsPage
        agentStatus="connected"
        workspacePath=""
        onClose={() => w.vscode?.postMessage?.({ command: 'closePanel' })}
      />
    );
  }
  if (isEmbedded && route === 'activity') {
    return <ActivityPanel onClose={() => w.vscode?.postMessage?.({ command: 'closePanel' })} />;
  }
  if (isEmbedded) {
    return <App />;
  }
  // Browser / Tauri: full shell with nav + hash routes.
  return <BrowserShell />;
}

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <Root />
  </React.StrictMode>
);
