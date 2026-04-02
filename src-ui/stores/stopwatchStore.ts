import { create } from 'zustand';
import { invoke } from '@tauri-apps/api/core';
import type { Stopwatch, StopwatchLap, CreateStopwatchRequest, UpdateStopwatchRequest } from '../types/index';

interface StopwatchState {
  stopwatches: Stopwatch[];
  activeStopwatch: Stopwatch | null;
  isLoading: boolean;
  error: string | null;
  
  // Actions
  loadStopwatches: () => Promise<void>;
  createStopwatch: (req: CreateStopwatchRequest) => Promise<Stopwatch | null>;
  updateStopwatch: (id: string, req: UpdateStopwatchRequest) => Promise<boolean>;
  deleteStopwatch: (id: string) => Promise<boolean>;
  startStopwatch: (id: string) => Promise<boolean>;
  pauseStopwatch: (id: string) => Promise<boolean>;
  resetStopwatch: (id: string) => Promise<boolean>;
  recordLap: (id: string) => Promise<StopwatchLap | null>;
  setActiveStopwatch: (stopwatch: Stopwatch | null) => void;
}

export const useStopwatchStore = create<StopwatchState>((set, get) => ({
  stopwatches: [],
  activeStopwatch: null,
  isLoading: false,
  error: null,

  loadStopwatches: async () => {
    set({ isLoading: true, error: null });
    try {
      const stopwatches = await invoke<Stopwatch[]>('get_stopwatches');
      set({ stopwatches, isLoading: false });
    } catch (err) {
      set({ error: String(err), isLoading: false });
    }
  },

  createStopwatch: async (req) => {
    try {
      const stopwatch = await invoke<Stopwatch>('create_stopwatch', { req });
      set((state) => ({ stopwatches: [stopwatch, ...state.stopwatches] }));
      return stopwatch;
    } catch (err) {
      set({ error: String(err) });
      return null;
    }
  },

  updateStopwatch: async (id, req) => {
    try {
      const success = await invoke<boolean>('update_stopwatch', { id, req });
      if (success) {
        await get().loadStopwatches();
      }
      return success;
    } catch (err) {
      set({ error: String(err) });
      return false;
    }
  },

  deleteStopwatch: async (id) => {
    try {
      const success = await invoke<boolean>('delete_stopwatch', { id });
      if (success) {
        set((state) => ({
          stopwatches: state.stopwatches.filter((s) => s.id !== id),
          activeStopwatch: state.activeStopwatch?.id === id ? null : state.activeStopwatch,
        }));
      }
      return success;
    } catch (err) {
      set({ error: String(err) });
      return false;
    }
  },

  startStopwatch: async (id) => {
    try {
      const success = await invoke<boolean>('start_stopwatch', { id });
      if (success) {
        await get().loadStopwatches();
      }
      return success;
    } catch (err) {
      set({ error: String(err) });
      return false;
    }
  },

  pauseStopwatch: async (id) => {
    try {
      const success = await invoke<boolean>('pause_stopwatch', { id });
      if (success) {
        await get().loadStopwatches();
      }
      return success;
    } catch (err) {
      set({ error: String(err) });
      return false;
    }
  },

  resetStopwatch: async (id) => {
    try {
      const success = await invoke<boolean>('reset_stopwatch', { id });
      if (success) {
        await get().loadStopwatches();
      }
      return success;
    } catch (err) {
      set({ error: String(err) });
      return false;
    }
  },

  recordLap: async (id) => {
    try {
      const lap = await invoke<StopwatchLap | null>('record_lap', { id });
      if (lap) {
        await get().loadStopwatches();
      }
      return lap;
    } catch (err) {
      set({ error: String(err) });
      return null;
    }
  },

  setActiveStopwatch: (stopwatch) => set({ activeStopwatch: stopwatch }),
}));
