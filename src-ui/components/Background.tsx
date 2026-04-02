import React, { useEffect } from 'react';
import { useImageStore } from '../stores';

interface BackgroundProps {
  opacity?: number;
}

export const Background: React.FC<BackgroundProps> = ({ opacity = 0.3 }) => {
  const { currentBackground, loadRandomBackground } = useImageStore();

  useEffect(() => {
    loadRandomBackground();
  }, [loadRandomBackground]);

  return (
    <div 
      className="fixed inset-0 -z-10 bg-gray-950"
      style={{
        backgroundImage: currentBackground ? `url(${currentBackground})` : undefined,
        backgroundSize: 'cover',
        backgroundPosition: 'center',
        backgroundRepeat: 'no-repeat',
      }}
    >
      {/* Dark overlay for readability */}
      <div 
        className="absolute inset-0 bg-gray-950"
        style={{ opacity: 1 - opacity }}
      />
    </div>
  );
};
