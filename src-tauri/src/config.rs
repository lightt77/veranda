use std::path::PathBuf;

/// Central configuration for Veranda
pub const DATA_DIR: &str = "$HOME/.veranda-data";
pub const DB_NAME: &str = "veranda.db";
pub const IMAGES_DIR: &str = "images";
pub const SOUNDS_DIR: &str = "sounds";

/// Get the absolute path to the data directory
pub fn get_data_dir() -> PathBuf {
    if DATA_DIR.starts_with("$HOME/") {
        if let Some(home) = dirs::home_dir() {
            return home.join(&DATA_DIR[6..]);
        }
    }
    PathBuf::from(DATA_DIR)
}

/// Get the path to the SQLite database
pub fn get_db_path() -> PathBuf {
    get_data_dir().join(DB_NAME)
}

/// Get the path to the images directory
pub fn get_images_dir() -> PathBuf {
    get_data_dir().join(IMAGES_DIR)
}

/// Get the path to the sounds directory
pub fn get_sounds_dir() -> PathBuf {
    get_data_dir().join(SOUNDS_DIR)
}

/// Ensure all data directories exist
pub fn ensure_data_dirs() -> std::io::Result<()> {
    std::fs::create_dir_all(get_data_dir())?;
    std::fs::create_dir_all(get_images_dir())?;
    std::fs::create_dir_all(get_sounds_dir())?;
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_data_paths() {
        assert_eq!(DATA_DIR, "../veranda-data");
        assert_eq!(DB_NAME, "veranda.db");

        let data_dir = get_data_dir();
        assert!(data_dir.ends_with("veranda-data"));
    }
}
