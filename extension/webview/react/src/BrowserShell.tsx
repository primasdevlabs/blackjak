import React, { useEffect, useState } from 'react';
import App from './App';
import { SettingsPage } from './settings/SettingsPage';
import { ActivityPanel } from './components/activity-panel';
import { settingsStore } from './state/settingsStore';
import {
  ChatBubbleLeftRightIcon,
  Cog6ToothIcon,
  ClockIcon,
} from '@heroicons/react/24/outline';

type Route = 'chat' | 'settings' | 'activity';

function routeFromHash(): Route {
  const h = (window.location.hash || '#/').replace(/^#\/?/, '');
  if (h.startsWith('settings')) return 'settings';
  if (h.startsWith('activity')) return 'activity';
  return 'chat';
}

/** Full-page browser / native shell with chat, settings, and activity routes. */
export const BrowserShell: React.FC = () => {
  const [route, setRoute] = useState<Route>(routeFromHash());
  const [settingsTick, setSettingsTick] = useState(0);

  useEffect(() => {
    const onHash = () => setRoute(routeFromHash());
    window.addEventListener('hashchange', onHash);
    const openSettings = () => {
      window.location.hash = '#/settings';
    };
    const openActivity = () => {
      window.location.hash = '#/activity';
    };
    window.addEventListener('blackjak:open-settings', openSettings);
    window.addEventListener('blackjak:open-activity', openActivity);
    const unsub = settingsStore.subscribe(() => setSettingsTick((n) => n + 1));
    void settingsStore.fetchSettings();
    return () => {
      window.removeEventListener('hashchange', onHash);
      window.removeEventListener('blackjak:open-settings', openSettings);
      window.removeEventListener('blackjak:open-activity', openActivity);
      unsub();
    };
  }, []);

  const go = (r: Route) => {
    window.location.hash = r === 'chat' ? '#/' : `#/${r}`;
  };

  return (
    <div className="app-browser-shell" data-settings-tick={settingsTick}>
      <nav className="app-browser-nav">
        <button className={route === 'chat' ? 'active' : ''} title="Chat" onClick={() => go('chat')}>
          <ChatBubbleLeftRightIcon className="icon" />
        </button>
        <button className={route === 'settings' ? 'active' : ''} title="Settings" onClick={() => go('settings')}>
          <Cog6ToothIcon className="icon" />
        </button>
        <button className={route === 'activity' ? 'active' : ''} title="Activity" onClick={() => go('activity')}>
          <ClockIcon className="icon" />
        </button>
      </nav>
      <main className="app-browser-main">
        {route === 'chat' && <App />}
        {route === 'settings' && (
          <SettingsPage
            agentStatus="connected"
            workspacePath=""
            onClose={() => go('chat')}
          />
        )}
        {route === 'activity' && <ActivityPanel onClose={() => go('chat')} />}
      </main>
    </div>
  );
};
