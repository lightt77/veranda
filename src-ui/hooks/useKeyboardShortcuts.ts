import { useEffect, useCallback } from 'react';

interface KeyboardShortcutsProps {
  onNewTimer?: () => void;
  onNewStopwatch?: () => void;
  onNewTodo?: () => void;
  onSearch?: () => void;
}

export const useKeyboardShortcuts = ({
  onNewTimer,
  onNewStopwatch,
  onNewTodo,
  onSearch,
}: KeyboardShortcutsProps) => {
  const handleKeyDown = useCallback(
    (event: KeyboardEvent) => {
      // Only trigger if not in an input field
      if (event.target instanceof HTMLInputElement || 
          event.target instanceof HTMLTextAreaElement) {
        return;
      }

      // Cmd/Ctrl + key shortcuts
      if (event.metaKey || event.ctrlKey) {
        switch (event.key.toLowerCase()) {
          case 't':
            event.preventDefault();
            onNewTimer?.();
            break;
          case 's':
            event.preventDefault();
            onNewStopwatch?.();
            break;
          case 'd':
            event.preventDefault();
            onNewTodo?.();
            break;
          case 'f':
            event.preventDefault();
            onSearch?.();
            break;
        }
      }
    },
    [onNewTimer, onNewStopwatch, onNewTodo, onSearch]
  );

  useEffect(() => {
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [handleKeyDown]);
};
