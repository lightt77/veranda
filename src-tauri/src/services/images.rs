//! Image download and management service

use crate::config::{DATA_DIR, IMAGES_DIR};
use crate::db::Database;
use sha2::{Digest, Sha256};
use std::fs;
use std::path::Path;
use std::sync::Arc;

/// Image service for downloading and managing background images
pub struct ImageService {
    db: Arc<Database>,
}

impl ImageService {
    pub fn new(db: Arc<Database>) -> Self {
        // Ensure images directory exists
        let images_path = Path::new(DATA_DIR).join(IMAGES_DIR);
        fs::create_dir_all(&images_path).ok();
        
        Self { db }
    }

    /// Download image from URL and save to local storage
    pub async fn download_image(&self, url: &str) -> anyhow::Result<String> {
        // Validate URL
        if !url.starts_with("http://") && !url.starts_with("https://") {
            return Err(anyhow::anyhow!("Invalid URL: must be HTTP or HTTPS"));
        }

        // Check if already exists
        if self.db.image_url_exists(url)? {
            return Err(anyhow::anyhow!("Image already exists"));
        }

        // Download image
        let response = reqwest::get(url).await?;
        
        if !response.status().is_success() {
            return Err(anyhow::anyhow!("Failed to download image: {}", response.status()));
        }

        // Get content type
        let content_type: String = response
            .headers()
            .get("content-type")
            .and_then(|ct| ct.to_str().ok())
            .unwrap_or("image/jpeg")
            .to_string();

        // Determine extension
        let extension = if content_type.contains("gif") {
            "gif"
        } else if content_type.contains("png") {
            "png"
        } else if content_type.contains("webp") {
            "webp"
        } else {
            "jpg"
        };

        // Generate unique ID from URL
        let id = format!("{:x}", Sha256::digest(url.as_bytes()))[..16].to_string();
        let filename = format!("{}.{}", id, extension);
        let local_path = Path::new(DATA_DIR).join(IMAGES_DIR).join(&filename);

        // Save image data
        let image_data = response.bytes().await?;
        fs::write(&local_path, &image_data)?;

        // Get image dimensions if possible
        let metadata = serde_json::json!({
            "content_type": content_type,
            "size_bytes": image_data.len(),
            "source_url": url,
        });

        // Save to database
        let local_path_str = local_path.to_string_lossy().to_string();
        self.db.save_image(&id, url, &local_path_str, Some(&metadata.to_string()))?;

        println!("Downloaded image: {} -> {}", url, local_path_str);
        
        Ok(id)
    }

    /// Get all active images
    pub fn get_active_images(&self) -> anyhow::Result<Vec<crate::db::images::Image>> {
        Ok(self.db.get_images(true)?)
    }

    /// Get image by ID
    pub fn get_image(&self, id: &str) -> anyhow::Result<Option<crate::db::images::Image>> {
        Ok(self.db.get_image(id)?)
    }

    /// Set image as active/inactive
    pub fn set_image_active(&self, id: &str, active: bool) -> anyhow::Result<bool> {
        Ok(self.db.set_image_active(id, active)?)
    }

    /// Delete image
    pub async fn delete_image(&self, id: &str) -> anyhow::Result<bool> {
        // Get image info first
        if let Some(image) = self.db.get_image(id)? {
            // Delete file
            if std::path::Path::new(&image.local_path).exists() {
                fs::remove_file(&image.local_path).ok();
            }
            
            // Delete from database
            Ok(self.db.delete_image(id)?)
        } else {
            Ok(false)
        }
    }

    /// Get a random active image for background
    pub fn get_random_background(&self) -> anyhow::Result<Option<String>> {
        let images = self.db.get_images(true)?;
        
        if images.is_empty() {
            return Ok(None);
        }

        // Pick random image
        use rand::seq::SliceRandom;
        let mut rng = rand::thread_rng();
        let image = images.choose(&mut rng);
        
        Ok(image.map(|img| img.local_path.clone()))
    }
}
