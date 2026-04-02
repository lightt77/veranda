import React from 'react';
import type { Timer } from '../types/index';

interface TimerCardProps {
  timer: Timer;
  onStart: (id: string) => void;
  onPause: (id: string) => void;
  onReset: (id: string) => void;
  onDelete: (id: string) => void;
  onEdit: (timer: Timer) => void;
}

const formatTime = (seconds: number): string => {
  const hrs = Math.floor(seconds / 3600);
  const mins = Math.floor((seconds % 3600) / 60);
  const secs = seconds % 60;
  if (hrs > 0) {
    return `${hrs}:${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
  }
  return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
};

const getProgressPercentage = (timer: Timer): number => {
  if (timer.duration_seconds === 0) return 100;
  const elapsed = timer.duration_seconds - timer.remaining_seconds;
  return Math.min(100, Math.max(0, (elapsed / timer.duration_seconds) * 100));
};

export const TimerCard: React.FC<TimerCardProps> = ({
  timer,
  onStart,
  onPause,
  onReset,
  onDelete,
  onEdit,
}) => {
  const progress = getProgressPercentage(timer);
  const isRunning = timer.status === 'RUNNING';
  const isDone = timer.status === 'DONE';

  return (
    <div
      className={`relative p-6 rounded-xl border-2 transition-all duration-300 ${
        isRunning
          ? 'bg-blue-900/30 border-blue-500 shadow-lg shadow-blue-900/20'
          : isDone
          ? 'bg-green-900/30 border-green-500'
          : 'bg-gray-800/50 border-gray-700 hover:border-gray-600'
      }`}
    >
      {/* Progress bar background */}
      <div
        className={`absolute bottom-0 left-0 h-1 rounded-b-xl transition-all duration-1000 ${
          isDone ? 'bg-green-500' : 'bg-blue-500'
        }`}
        style={{ width: `${progress}%` }}
      />

      <div className="flex justify-between items-start mb-4">
        <div className="flex-1">
          <h3 
            className="font-semibold text-lg text-white cursor-pointer hover:text-blue-400 transition-colors"
            onClick={() => onEdit(timer)}
          >
            {timer.label}
          </h3>
          {timer.note && (
            <p className="text-sm text-gray-400 mt-1">{timer.note}</p>
          )}
        </div>
        <span
          className={`px-2 py-1 rounded-full text-xs font-medium ${
            isRunning
              ? 'bg-blue-600 text-white animate-pulse'
              : isDone
              ? 'bg-green-600 text-white'
              : 'bg-gray-700 text-gray-300'
          }`}
        >
          {isRunning ? 'Running' : isDone ? 'Done' : timer.status}
        </span>
      </div>

      <div className="text-5xl font-mono font-bold text-white mb-6 tracking-wider">
        {formatTime(timer.remaining_seconds)}
      </div>

      <div className="flex gap-2">
        {isRunning ? (
          <button
            onClick={() => onPause(timer.id)}
            className="flex-1 px-4 py-2 bg-yellow-600 hover:bg-yellow-700 text-white rounded-lg font-medium transition-colors flex items-center justify-center gap-2"
          >
            <span>⏸</span> Pause
          </button>
        ) : isDone ? (
          <button
            onClick={() => onReset(timer.id)}
            className="flex-1 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-medium transition-colors flex items-center justify-center gap-2"
          >
            <span>↺</span> Restart
          </button>
        ) : (
          <button
            onClick={() => onStart(timer.id)}
            className="flex-1 px-4 py-2 bg-green-600 hover:bg-green-700 text-white rounded-lg font-medium transition-colors flex items-center justify-center gap-2"
          >
            <span>▶</span> Start
          </button>
        )}

        {!isRunning && !isDone && (
          <button
            onClick={() => onReset(timer.id)}
            className="px-4 py-2 bg-gray-700 hover:bg-gray-600 text-white rounded-lg font-medium transition-colors"
          >
            ↺
          </button>
        )}

        <button
          onClick={() => onDelete(timer.id)}
          className="px-4 py-2 bg-red-600/80 hover:bg-red-700 text-white rounded-lg font-medium transition-colors"
        >
          🗑
        </button>
      </div>
    </div>
  );
};
