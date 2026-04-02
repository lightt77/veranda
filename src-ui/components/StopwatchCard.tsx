import React from 'react';
import type { Stopwatch } from '../types/index';

interface StopwatchCardProps {
  stopwatch: Stopwatch;
  onStart: (id: string) => void;
  onPause: (id: string) => void;
  onReset: (id: string) => void;
  onDelete: (id: string) => void;
  onRecordLap: (id: string) => void;
  onEdit: (stopwatch: Stopwatch) => void;
}

const formatTime = (seconds: number): string => {
  const hrs = Math.floor(seconds / 3600);
  const mins = Math.floor((seconds % 3600) / 60);
  const secs = seconds % 60;
  return `${hrs.toString().padStart(2, '0')}:${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
};

const formatLapTime = (ms: number): string => {
  const totalSeconds = Math.floor(ms / 1000);
  const mins = Math.floor(totalSeconds / 60);
  const secs = totalSeconds % 60;
  const milliseconds = Math.floor((ms % 1000) / 10);
  return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}.${milliseconds.toString().padStart(2, '0')}`;
};

export const StopwatchCard: React.FC<StopwatchCardProps> = ({
  stopwatch,
  onStart,
  onPause,
  onReset,
  onDelete,
  onRecordLap,
  onEdit,
}) => {
  const isRunning = stopwatch.status === 'RUNNING';
  const laps = stopwatch.laps || [];

  return (
    <div
      className={`p-6 rounded-xl border-2 transition-all duration-300 ${
        isRunning
          ? 'bg-blue-900/30 border-blue-500 shadow-lg shadow-blue-900/20'
          : 'bg-gray-800/50 border-gray-700 hover:border-gray-600'
      }`}
    >
      <div className="flex justify-between items-start mb-4">
        <h3 
          className="font-semibold text-lg text-white cursor-pointer hover:text-blue-400 transition-colors"
          onClick={() => onEdit(stopwatch)}
        >
          {stopwatch.label}
        </h3>
        {isRunning && (
          <span className="px-2 py-1 rounded-full text-xs font-medium bg-blue-600 text-white animate-pulse">
            Running
          </span>
        )}
      </div>

      <div className="text-5xl font-mono font-bold text-white mb-6 tracking-wider">
        {formatTime(stopwatch.elapsed_seconds)}
      </div>

      <div className="flex gap-2 flex-wrap mb-4">
        {isRunning ? (
          <>
            <button
              onClick={() => onRecordLap(stopwatch.id)}
              className="flex-1 px-4 py-2 bg-purple-600 hover:bg-purple-700 text-white rounded-lg font-medium transition-colors"
            >
              Lap
            </button>
            <button
              onClick={() => onPause(stopwatch.id)}
              className="flex-1 px-4 py-2 bg-yellow-600 hover:bg-yellow-700 text-white rounded-lg font-medium transition-colors"
            >
              Pause
            </button>
          </>
        ) : (
          <button
            onClick={() => onStart(stopwatch.id)}
            className="flex-1 px-4 py-2 bg-green-600 hover:bg-green-700 text-white rounded-lg font-medium transition-colors"
          >
            Start
          </button>
        )}

        <button
          onClick={() => onReset(stopwatch.id)}
          className="px-4 py-2 bg-gray-700 hover:bg-gray-600 text-white rounded-lg font-medium transition-colors"
        >
          ↺
        </button>

        <button
          onClick={() => onDelete(stopwatch.id)}
          className="px-4 py-2 bg-red-600/80 hover:bg-red-700 text-white rounded-lg font-medium transition-colors"
        >
          🗑
        </button>
      </div>

      {/* Laps */}
      {laps.length > 0 && (
        <div className="mt-4 pt-4 border-t border-gray-700">
          <p className="text-sm text-gray-400 mb-2">Laps ({laps.length})</p>
          <div className="max-h-32 overflow-y-auto space-y-1">
            {[...laps].reverse().slice(0, 5).map((lap) => (
              <div
                key={lap.id}
                className="flex justify-between text-sm px-2 py-1 bg-gray-900/50 rounded"
              >
                <span className="text-gray-400">#{lap.lap_number}</span>
                <span className="text-white font-mono">{formatLapTime(lap.lap_duration_ms)}</span>
              </div>
            ))}
            {laps.length > 5 && (
              <p className="text-xs text-gray-500 text-center py-1">
                +{laps.length - 5} more laps
              </p>
            )}
          </div>
        </div>
      )}
    </div>
  );
};
