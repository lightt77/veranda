import React, { useEffect } from 'react';
import { useTimerStore } from '../stores';

export const TimerView: React.FC = () => {
  const { timers, isLoading, loadTimers, createTimer, startTimer, pauseTimer, deleteTimer } = useTimerStore();

  useEffect(() => {
    loadTimers();
    // Refresh every second to update display
    const interval = setInterval(loadTimers, 1000);
    return () => clearInterval(interval);
  }, [loadTimers]);

  const formatTime = (seconds: number): string => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
  };

  const handleCreateTimer = async () => {
    await createTimer({
      label: 'New Timer',
      duration_seconds: 300, // 5 minutes
    });
  };

  if (isLoading && timers.length === 0) {
    return <div className="p-8 text-gray-400">Loading timers...</div>;
  }

  return (
    <div className="p-8">
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-2xl font-bold">Timers</h2>
        <button
          onClick={handleCreateTimer}
          className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-medium transition-colors"
        >
          + New Timer
        </button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {timers.map((timer) => (
          <div
            key={timer.id}
            className={`p-6 rounded-xl border ${
              timer.status === 'RUNNING'
                ? 'bg-blue-900/20 border-blue-700'
                : timer.status === 'DONE'
                ? 'bg-green-900/20 border-green-700'
                : 'bg-gray-800 border-gray-700'
            }`}
          >
            <h3 className="font-semibold text-lg mb-2">{timer.label}</h3>
            <div className="text-4xl font-mono font-bold mb-4">
              {formatTime(timer.remaining_seconds)}
            </div>
            <div className="flex gap-2">
              {timer.status === 'RUNNING' ? (
                <button
                  onClick={() => pauseTimer(timer.id)}
                  className="px-3 py-1 bg-yellow-600 hover:bg-yellow-700 text-white rounded text-sm"
                >
                  Pause
                </button>
              ) : timer.status === 'DONE' ? (
                <span className="px-3 py-1 text-green-400 text-sm">Completed</span>
              ) : (
                <button
                  onClick={() => startTimer(timer.id)}
                  className="px-3 py-1 bg-green-600 hover:bg-green-700 text-white rounded text-sm"
                >
                  Start
                </button>
              )}
              <button
                onClick={() => deleteTimer(timer.id)}
                className="px-3 py-1 bg-red-600 hover:bg-red-700 text-white rounded text-sm"
              >
                Delete
              </button>
            </div>
          </div>
        ))}
      </div>

      {timers.length === 0 && (
        <div className="text-center py-12 text-gray-500">
          <p>No timers yet. Create one to get started!</p>
        </div>
      )}
    </div>
  );
};
