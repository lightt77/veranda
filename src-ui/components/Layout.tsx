import React from 'react';
import { Background } from './Background';

interface LayoutProps {
  children: React.ReactNode;
}

export const Layout: React.FC<LayoutProps> = ({ children }) => {
  return (
    <div className="min-h-screen text-gray-100 relative">
      <Background opacity={0.2} />
      <div className="relative z-10">
        {children}
      </div>
    </div>
  );
};
