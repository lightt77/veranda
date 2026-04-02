//! Sound service for playing notification sounds
//!
//! Uses macOS system sounds via afplay command

/// Sound service for playing audio notifications
pub struct SoundService;

impl SoundService {
    /// Create a new sound service
    pub fn new() -> anyhow::Result<Self> {
        Ok(Self)
    }

    /// Play the default timer completion sound
    pub fn play_timer_complete(&self) {
        // Play macOS system sound using afplay command
        if let Err(e) = Self::play_system_sound("Ping") {
            eprintln!("Failed to play system sound: {}", e);
        }
    }

    /// Play a macOS system sound by name
    /// Available sounds: Basso, Blow, Bottle, Frog, Funk, Glass, Hero,
    /// Morse, Ping, Pop, Purr, Sosumi, Submarine, Tink
    fn play_system_sound(sound_name: &str) -> anyhow::Result<()> {
        let sound_path = format!("/System/Library/Sounds/{}.aiff", sound_name);

        // Spawn a background process to play the sound
        std::process::Command::new("afplay")
            .arg(&sound_path)
            .stdout(std::process::Stdio::null())
            .stderr(std::process::Stdio::null())
            .spawn()?;

        Ok(())
    }

    /// Play the "Tink" sound (subtle notification)
    pub fn play_notification(&self) {
        if let Err(e) = Self::play_system_sound("Tink") {
            eprintln!("Failed to play notification sound: {}", e);
        }
    }
}

impl Default for SoundService {
    fn default() -> Self {
        Self::new().expect("Failed to initialize sound service")
    }
}
