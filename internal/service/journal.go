package service

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/lightt77/veranda/internal/models"
	"github.com/lightt77/veranda/internal/repository"
)

// JournalService handles journal business logic
type JournalService struct {
	journalRepo *repository.JournalRepository
}

// NewJournalService creates a new journal service
func NewJournalService(journalRepo *repository.JournalRepository) *JournalService {
	return &JournalService{
		journalRepo: journalRepo,
	}
}

// Create creates a new journal entry
func (s *JournalService) Create(log string) (*models.JournalEntry, error) {
	now := time.Now()
	entry := models.NewJournalEntry(GenerateID(), log, now)

	if err := s.journalRepo.Create(entry); err != nil {
		return nil, fmt.Errorf("failed to create entry: %w", err)
	}

	return entry, nil
}

// CreateWithEditor opens the system editor for the user to write a journal entry
func (s *JournalService) CreateWithEditor() (*models.JournalEntry, error) {
	// Get the editor from environment or use a default
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		// Try common editors
		for _, ed := range []string{"vim", "nano", "emacs", "code"} {
			if _, err := exec.LookPath(ed); err == nil {
				editor = ed
				break
			}
		}
	}
	if editor == "" {
		return nil, fmt.Errorf("no editor found. Set $EDITOR environment variable")
	}

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "veranda-journal-*.txt")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Open the editor
	cmd := exec.Command(editor, tmpFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("editor exited with error: %w", err)
	}

	// Read the content
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to read journal file: %w", err)
	}

	// Trim whitespace and check if empty
	log := string(content)
	if len(log) == 0 || (len(log) == 1 && log[0] == '\n') {
		return nil, fmt.Errorf("journal entry is empty")
	}

	return s.Create(log)
}

// Get retrieves a journal entry by ID
func (s *JournalService) Get(id int64) (*models.JournalEntry, error) {
	return s.journalRepo.GetByID(id)
}

// GetAll retrieves all non-deleted journal entries
func (s *JournalService) GetAll() ([]*models.JournalEntry, error) {
	return s.journalRepo.GetAll()
}

// GetRecent retrieves recent journal entries
func (s *JournalService) GetRecent(limit int) ([]*models.JournalEntry, error) {
	return s.journalRepo.GetRecent(limit)
}

// Delete soft-deletes a journal entry
func (s *JournalService) Delete(id int64) error {
	return s.journalRepo.SoftDelete(id)
}
