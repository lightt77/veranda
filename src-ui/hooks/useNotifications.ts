import { useEffect } from 'react';
import { invoke } from '@tauri-apps/api/core';

export const useNotifications = () => {
  useEffect(() => {
    // Request notification permission on mount
    if ('Notification' in window && Notification.permission === 'default') {
      Notification.requestPermission();
    }
  }, []);

  const showNotification = async (title: string, body: string) => {
    // Try native notification first
    try {
      // Check if notifications are enabled in settings
      const enabled = await invoke<boolean>('get_setting', { 
        key: 'notification_sound_enabled' 
      }).catch(() => true);
      
      if (!enabled) return;
    } catch {
      // If backend call fails, continue with browser notification
    }

    // Browser notification
    if ('Notification' in window && Notification.permission === 'granted') {
      new Notification(title, {
        body,
        icon: '/tauri.svg', // Default icon
      });
    }
  };

  return { showNotification };
};
