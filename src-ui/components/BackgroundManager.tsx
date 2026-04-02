import React, { useState } from 'react';
import { useImageStore, type Image } from '../stores';

interface BackgroundManagerProps {
  isOpen: boolean;
  onClose: () => void;
}

export const BackgroundManager: React.FC<BackgroundManagerProps> = ({
  isOpen,
  onClose,
}) => {
  const { images, downloadImage, setImageActive, deleteImage } = useImageStore();
  const [url, setUrl] = useState('');
  const [isDownloading, setIsDownloading] = useState(false);

  if (!isOpen) return null;

  const handleAddImage = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!url.trim()) return;

    setIsDownloading(true);
    try {
      await downloadImage(url);
      setUrl('');
    } catch (err) {
      console.error('Failed to download:', err);
    } finally {
      setIsDownloading(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50">
      <div className="bg-gray-800 rounded-xl p-6 w-full max-w-2xl border border-gray-700 max-h-[80vh] overflow-auto">
        <div className="flex justify-between items-center mb-6">
          <h2 className="text-xl font-bold text-white">Background Images</h2>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-white transition-colors"
          >
            ✕
          </button>
        </div>

        {/* Add new image */}
        <form onSubmit={handleAddImage} className="mb-6">
          <label className="block text-sm font-medium text-gray-300 mb-2">
            Add Image URL (Pinterest, etc.)
          </label>
          <div className="flex gap-2">
            <input
              type="url"
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              placeholder="https://pinterest.com/..."
              className="flex-1 px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"
            />
            <button
              type="submit"
              disabled={!url.trim() || isDownloading}
              className="px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-600 disabled:cursor-not-allowed text-white rounded-lg font-medium transition-colors"
            >
              {isDownloading ? '...' : 'Add'}
            </button>
          </div>
          <p className="text-xs text-gray-500 mt-2">
            Supports: JPG, PNG, GIF, WebP from Pinterest and other sources
          </p>
        </form>

        {/* Image grid */}
        {images.length > 0 ? (
          <div className="grid grid-cols-3 gap-4">
            {images.map((image: Image) => (
              <div
                key={image.id}
                className={`relative aspect-video rounded-lg overflow-hidden border-2 ${
                  image.is_active ? 'border-blue-500' : 'border-gray-700'
                }`}
              >
                <img
                  src={`asset://${image.local_path}`}
                  alt="Background"
                  className="w-full h-full object-cover"
                />
                <div className="absolute inset-0 bg-black/50 opacity-0 hover:opacity-100 transition-opacity flex items-center justify-center gap-2">
                  <button
                    onClick={() => setImageActive(image.id, !image.is_active)}
                    className={`px-3 py-1 rounded text-sm font-medium ${
                      image.is_active
                        ? 'bg-yellow-600 hover:bg-yellow-700 text-white'
                        : 'bg-green-600 hover:bg-green-700 text-white'
                    }`}
                  >
                    {image.is_active ? 'Disable' : 'Enable'}
                  </button>
                  <button
                    onClick={() => deleteImage(image.id)}
                    className="px-3 py-1 bg-red-600 hover:bg-red-700 text-white rounded text-sm font-medium"
                  >
                    Delete
                  </button>
                </div>
                {image.is_active && (
                  <div className="absolute top-2 right-2 px-2 py-1 bg-blue-600 text-white text-xs rounded-full">
                    Active
                  </div>
                )}
              </div>
            ))}
          </div>
        ) : (
          <div className="text-center py-8 text-gray-500">
            <p>No background images yet.</p>
            <p className="text-sm">Add a Pinterest URL above to get started!</p>
          </div>
        )}
      </div>
    </div>
  );
};
