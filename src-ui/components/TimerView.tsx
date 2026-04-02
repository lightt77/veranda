import React, { useEffect, useState } from 'react';
import { useTimerStore } from '../stores';
import { TimerCard } from './TimerCard';
import { CreateTimerModal } from './CreateTimerModal';
import { EditTimerModal } from './EditTimerModal';
import type { Timer } from '../types/index';

export const TimerView: React.FC = () => {
  const { timers, isLoading, loadTimers, createTimer, startTimer, pauseTimer, resetTimer, deleteTimer, updateTimer } = useTimerStore();
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [editingTimer, setEditingTimer] = useState<Timer | null>(null);

  useEffect(() => {
    loadTimers();
    const interval = setInterval(loadTimers, 1000);
    return () => clearInterval(interval);
  }, [loadTimers]);

  const handleCreateTimer = async (label: string, durationMinutes: number, note: string) => {
    await createTimer({
      label,
      duration_seconds: durationMinutes * 60,
      note: note || undefined,
    });
  };

  const handleEditTimer = async (id: string, label: string, note: string) => {
    await updateTimer(id, { label, note });
  };

  const runningCount = timers.filter(t => t.status === 'RUNNING').length;

  return (
    <div className="p-8">
      <div className="flex justify-between items-center mb-6">
        <div>
          <h2 className="text-2xl font-bold text-white">Timers</h2>
          {runningCount > 0 && (
            <p className="text-blue-400 text-sm mt-1">
              {runningCount} timer{runningCount !== 1 ? 's' : ''} running
            </p>
          )}
        </div>
        <button
          onClick={() => setIsCreateModalOpen(true)}
          className="px-6 py-3 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-medium transition-colors flex items-center gap-2 shadow-lg shadow-blue-900/20"
        >
          <span className="text-xl">+</span>
          New Timer
        </button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {timers.map((timer) => (
          <TimerCard
            key={timer.id}
            timer={timer}
            onStart={startTimer}
            onPause={pauseTimer}
            onReset={resetTimer}
            onDelete={deleteTimer}
            onEdit={setEditingTimer}
          />
        ))}
      </div>

      {timers.length === 0 && !isLoading && (
        <div className="text-center py-16">
          <div className="text-6xl mb-4">⏱️</div>
          <h3 className="text-xl font-semibold text-gray-300 mb-2">No timers yet</h3>
          <p className="text-gray-500 mb-6">Create a timer to stay focused and track your time</p>
          <button
            onClick={() => setIsCreateModalOpen(true)}
            className="px-6 py-3 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-medium transition-colors"
          >
            Create your first timer
          </button>
        </div>
      )}

      <CreateTimerModal
        isOpen={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
        onCreate={handleCreateTimer}
      />

      <EditTimerModal
        timer={editingTimer}
        isOpen={!!editingTimer}
        onClose={() => setEditingTimer(null)}
        onSave={handleEditTimer}
      />
    </div>
  );
};
