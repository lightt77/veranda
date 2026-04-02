import { create } from 'zustand';
import { invoke } from '@tauri-apps/api/core';

export interface Image {
  id: string;
  url: string;
  local_path: string;
  downloaded_at: string;
  is_active: boolean;
  metadata?: string;
}

interface ImageStore {
  images: Image[];
  currentBackground: string | null;
  isLoading: boolean;
  error: string | null;
  
  // Actions
  loadImages: () => Promise<void>;
  downloadImage: (url: string) => Promise<string | null>;
  setImageActive: (id: string, active: boolean) => Promise<boolean>;
  deleteImage: (id: string) => Promise<boolean>;
  loadRandomBackground: () => Promise<void>;
  setBackground: (path: string) => void;
}

export const useImageStore = create<ImageStore>((set, get) => ({
  images: [],
  currentBackground: null,
  isLoading: false,
  error: null,

  loadImages: async () => {
    set({ isLoading: true, error: null });
    try {
      const images = await invoke<Image[]>('get_images', { onlyActive: false });
      set({ images, isLoading: false });
    } catch (err) {
      set({ error: String(err), isLoading: false });
    }
  },

  downloadImage: async (url: string) => {
    try {
      const id = await invoke<string>('download_image', { url });
      await get().loadImages();
      return id;
    } catch (err) {
      set({ error: String(err) });
      return null;
    }
  },

  setImageActive: async (id: string, active: boolean) => {
    try {
      const success = await invoke<boolean>('set_image_active', { id, active });
      if (success) {
        await get().loadImages();
      }
      return success;
    } catch (err) {
      set({ error: String(err) });
      return false;
    }
  },

  deleteImage: async (id: string) => {
    try {
      const success = await invoke<boolean>('delete_image', { id });
      if (success) {
        set((state) => ({
          images: state.images.filter((img) => img.id !== id),
        }));
      }
      return success;
    } catch (err) {
      set({ error: String(err) });
      return false;
    }
  },

  loadRandomBackground: async () => {
    try {
      const path = await invoke<string | null>('get_random_background');
      if (path) {
        // Convert to asset URL for Tauri
        const assetUrl = `asset://${path}`;
        set({ currentBackground: assetUrl });
      }
    } catch (err) {
      console.error('Failed to load background:', err);
    }
  },

  setBackground: (path: string) => {
    set({ currentBackground: path });
  },
}));
