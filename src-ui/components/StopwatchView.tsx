import React, { useEffect, useState } from 'react';
import { useStopwatchStore } from '../stores';
import { StopwatchCard } from './StopwatchCard';
import { CreateStopwatchModal } from './CreateStopwatchModal';


export const StopwatchView: React.FC = () => {
  const { 
    stopwatches, 
    isLoading, 
    loadStopwatches, 
    createStopwatch, 
    startStopwatch, 
    pauseStopwatch, 
    resetStopwatch,
    deleteStopwatch,
    recordLap
  } = useStopwatchStore();
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);

  useEffect(() => {
    loadStopwatches();
    const interval = setInterval(loadStopwatches, 100);
    return () => clearInterval(interval);
  }, [loadStopwatches]);

  const handleCreateStopwatch = async (label: string) => {
    await createStopwatch({ label });
  };

  const runningCount = stopwatches.filter(s => s.status === 'RUNNING').length;

  return (
    <div className="p-8">
      <div className="flex justify-between items-center mb-6">
        <div>
          <h2 className="text-2xl font-bold text-white">Stopwatches</h2>
          {runningCount > 0 && (
            <p className="text-blue-400 text-sm mt-1">
              {runningCount} stopwatch{runningCount !== 1 ? 'es' : ''} running
            </p>
          )}
        </div>
        <button
          onClick={() => setIsCreateModalOpen(true)}
          className="px-6 py-3 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-medium transition-colors flex items-center gap-2 shadow-lg shadow-blue-900/20"
        >
          <span className="text-xl">+</span>
          New Stopwatch
        </button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {stopwatches.map((stopwatch) => (
          <StopwatchCard
            key={stopwatch.id}
            stopwatch={stopwatch}
            onStart={startStopwatch}
            onPause={pauseStopwatch}
            onReset={resetStopwatch}
            onDelete={deleteStopwatch}
            onRecordLap={recordLap}
            onEdit={() => {}} // TODO: Add edit modal
          />
        ))}
      </div>

      {stopwatches.length === 0 && !isLoading && (
        <div className="text-center py-16">
          <div className="text-6xl mb-4">⏲️</div>
          <h3 className="text-xl font-semibold text-gray-300 mb-2">No stopwatches yet</h3>
          <p className="text-gray-500 mb-6">Create a stopwatch to track elapsed time with laps</p>
          <button
            onClick={() => setIsCreateModalOpen(true)}
            className="px-6 py-3 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-medium transition-colors"
          >
            Create your first stopwatch
          </button>
        </div>
      )}

      <CreateStopwatchModal
        isOpen={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
        onCreate={handleCreateStopwatch}
      />
    </div>
  );
};
