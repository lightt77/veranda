import { useState, useEffect } from 'react';
import { Layout, Sidebar, TimerView, StopwatchView, TodoView } from './components';
import { useKeyboardShortcuts } from './hooks/useKeyboardShortcuts';
import { invoke } from '@tauri-apps/api/core';

type Tab = 'timers' | 'stopwatches' | 'todos';

function App() {
  const [activeTab, setActiveTab] = useState<Tab>('timers');
  const [isReady, setIsReady] = useState(false);

  useEffect(() => {
    invoke<string>('health_check')
      .then((result) => {
        console.log('Backend status:', result);
        setIsReady(true);
      })
      .catch((err) => {
        console.error('Backend error:', err);
        setIsReady(true);
      });
  }, []);

  // Keyboard shortcuts
  useKeyboardShortcuts({
    onNewTimer: () => {
      // Trigger new timer - this would need to be passed down to TimerView
      console.log('Shortcut: New Timer');
    },
    onNewStopwatch: () => {
      console.log('Shortcut: New Stopwatch');
    },
    onNewTodo: () => {
      console.log('Shortcut: New Todo');
    },
  });

  const renderContent = () => {
    switch (activeTab) {
      case 'timers':
        return <TimerView />;
      case 'stopwatches':
        return <StopwatchView />;
      case 'todos':
        return <TodoView />;
      default:
        return <TimerView />;
    }
  };

  if (!isReady) {
    return (
      <Layout>
        <div className="flex items-center justify-center h-screen">
          <div className="text-center">
            <div className="text-4xl mb-4">⏱️</div>
            <div className="text-xl text-gray-400">Loading Veranda...</div>
          </div>
        </div>
      </Layout>
    );
  }

  return (
    <Layout>
      <div className="flex h-screen">
        <Sidebar activeTab={activeTab} onTabChange={(tab) => setActiveTab(tab as Tab)} />
        <main className="flex-1 overflow-auto">
          {renderContent()}
        </main>
      </div>
    </Layout>
  );
}

export default App;
