import React, { useEffect } from 'react';
import { useStopwatchStore } from '../stores';

export const StopwatchView: React.FC = () => {
  const { 
    stopwatches, 
    isLoading, 
    loadStopwatches, 
    createStopwatch, 
    startStopwatch, 
    pauseStopwatch, 
    resetStopwatch,
    deleteStopwatch 
  } = useStopwatchStore();

  useEffect(() => {
    loadStopwatches();
    // Refresh frequently for smooth display
    const interval = setInterval(loadStopwatches, 100);
    return () => clearInterval(interval);
  }, [loadStopwatches]);

  const formatTime = (seconds: number): string => {
    const hrs = Math.floor(seconds / 3600);
    const mins = Math.floor((seconds % 3600) / 60);
    const secs = seconds % 60;
    return `${hrs.toString().padStart(2, '0')}:${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
  };

  const handleCreateStopwatch = async () => {
    await createStopwatch({
      label: 'New Stopwatch',
    });
  };

  if (isLoading && stopwatches.length === 0) {
    return <div className="p-8 text-gray-400">Loading stopwatches...</div>;
  }

  return (
    <div className="p-8">
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-2xl font-bold">Stopwatches</h2>
        <button
          onClick={handleCreateStopwatch}
          className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-medium transition-colors"
        >
          + New Stopwatch
        </button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {stopwatches.map((stopwatch) => (
          <div
            key={stopwatch.id}
            className={`p-6 rounded-xl border ${
              stopwatch.status === 'RUNNING'
                ? 'bg-blue-900/20 border-blue-700'
                : 'bg-gray-800 border-gray-700'
            }`}
          >
            <h3 className="font-semibold text-lg mb-2">{stopwatch.label}</h3>
            <div className="text-4xl font-mono font-bold mb-4">
              {formatTime(stopwatch.elapsed_seconds)}
            </div>
            <div className="flex gap-2 flex-wrap">
              {stopwatch.status === 'RUNNING' ? (
                <button
                  onClick={() => pauseStopwatch(stopwatch.id)}
                  className="px-3 py-1 bg-yellow-600 hover:bg-yellow-700 text-white rounded text-sm"
                >
                  Pause
                </button>
              ) : (
                <button
                  onClick={() => startStopwatch(stopwatch.id)}
                  className="px-3 py-1 bg-green-600 hover:bg-green-700 text-white rounded text-sm"
                >
                  Start
                </button>
              )}
              <button
                onClick={() => resetStopwatch(stopwatch.id)}
                className="px-3 py-1 bg-gray-600 hover:bg-gray-700 text-white rounded text-sm"
              >
                Reset
              </button>
              <button
                onClick={() => deleteStopwatch(stopwatch.id)}
                className="px-3 py-1 bg-red-600 hover:bg-red-700 text-white rounded text-sm"
              >
                Delete
              </button>
            </div>
            {stopwatch.laps && stopwatch.laps.length > 0 && (
              <div className="mt-4 pt-4 border-t border-gray-700">
                <p className="text-sm text-gray-400 mb-2">Laps: {stopwatch.laps.length}</p>
              </div>
            )}
          </div>
        ))}
      </div>

      {stopwatches.length === 0 && (
        <div className="text-center py-12 text-gray-500">
          <p>No stopwatches yet. Create one to get started!</p>
        </div>
      )}
    </div>
  );
};
