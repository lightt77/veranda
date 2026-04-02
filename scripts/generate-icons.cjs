const { createCanvas } = require('canvas');
const fs = require('fs');
const path = require('path');

function createIcon(size) {
  const canvas = createCanvas(size, size);
  const ctx = canvas.getContext('2d');
  
  // Create gradient
  const gradient = ctx.createLinearGradient(0, 0, 0, size);
  gradient.addColorStop(0, '#2DD4BF');
  gradient.addColorStop(1, '#3B82F6');
  
  // Draw rounded rectangle
  const radius = size / 8;
  ctx.beginPath();
  ctx.roundRect(0, 0, size, size, radius);
  ctx.fillStyle = gradient;
  ctx.fill();
  
  // Draw "V"
  ctx.fillStyle = 'white';
  ctx.font = `bold ${size * 0.55}px Arial, Helvetica, sans-serif`;
  ctx.textAlign = 'center';
  ctx.textBaseline = 'middle';
  ctx.fillText('V', size / 2, size / 2 + size * 0.05);
  
  return canvas.toBuffer('image/png');
}

const iconDir = path.join(__dirname, '../src-tauri/icons');
const sizes = [1024, 512, 256, 128, 64, 32];

sizes.forEach(size => {
  const buffer = createIcon(size);
  fs.writeFileSync(path.join(iconDir, `icon-${size}x${size}.png`), buffer);
  console.log(`Created icon-${size}x${size}.png`);
});

// Create main icon.png
const buffer = createIcon(1024);
fs.writeFileSync(path.join(iconDir, 'icon.png'), buffer);
console.log('Created icon.png');

console.log('All icons created successfully!');