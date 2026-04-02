import React, { useState } from 'react';
import type { Timer } from '../types/index';

interface EditTimerModalProps {
  timer: Timer | null;
  isOpen: boolean;
  onClose: () => void;
  onSave: (id: string, label: string, note: string) => void;
}

export const EditTimerModal: React.FC<EditTimerModalProps> = ({
  timer,
  isOpen,
  onClose,
  onSave,
}) => {
  const [label, setLabel] = useState(timer?.label || '');
  const [note, setNote] = useState(timer?.note || '');

  React.useEffect(() => {
    if (timer) {
      setLabel(timer.label);
      setNote(timer.note || '');
    }
  }, [timer]);

  if (!isOpen || !timer) return null;

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSave(timer.id, label, note);
    onClose();
  };

  return (
    <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50">
      <div className="bg-gray-800 rounded-xl p-6 w-full max-w-md border border-gray-700">
        <h2 className="text-xl font-bold text-white mb-4">Edit Timer</h2>
        
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-300 mb-1">
              Label
            </label>
            <input
              type="text"
              value={label}
              onChange={(e) => setLabel(e.target.value)}
              className="w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:border-blue-500"
              placeholder="Timer name..."
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-300 mb-1">
              Note
            </label>
            <textarea
              value={note}
              onChange={(e) => setNote(e.target.value)}
              className="w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:border-blue-500"
              placeholder="Optional note..."
              rows={3}
            />
          </div>

          <div className="flex gap-3 pt-4">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 px-4 py-2 bg-gray-700 hover:bg-gray-600 text-white rounded-lg font-medium transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              className="flex-1 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-medium transition-colors"
            >
              Save
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
