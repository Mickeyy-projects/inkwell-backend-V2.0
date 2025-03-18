package service

import (
	"fmt"
	"inkwell-backend-V2.0/utilities"
	"os"
	"path/filepath"
	"time"

	"github.com/jung-kurt/gofpdf"
	"inkwell-backend-V2.0/internal/model"
	"inkwell-backend-V2.0/internal/repository"
)

type ComicService interface {
	GenerateComic(storyID uint) error
}

type comicService struct {
	storyRepo repository.StoryRepository
}

func NewComicService(storyRepo repository.StoryRepository) ComicService {
	return &comicService{storyRepo: storyRepo}
}
func InitComicEventListeners(storyRepo repository.StoryRepository) {
	utilities.GlobalEventBus.Subscribe("story_completed", func(data interface{}) {
		storyID, ok := data.(uint)
		if !ok {
			fmt.Println("Invalid story ID received for comic generation")
			return
		}

		comicService := NewComicService(storyRepo)
		err := comicService.GenerateComic(storyID)
		if err != nil {
			fmt.Printf("Error generating comic for story %d: %v\n", storyID, err)
		}
	})
}

func (s *comicService) GenerateComic(storyID uint) error {
	story, err := s.storyRepo.GetCurrentStoryByUser(storyID)
	if err != nil {
		return fmt.Errorf("failed to fetch story: %w", err)
	}

	// Get all sentences with images
	sentences, err := s.storyRepo.GetSentencesByStory(storyID)
	if err != nil {
		return fmt.Errorf("failed to fetch sentences: %w", err)
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetFont("Arial", "B", 16)
	pdf.AddPage()

	pdf.Cell(40, 10, story.Title)
	pdf.Ln(20)

	for _, sentence := range sentences {
		if sentence.ImageURL != "" {
			imgPath := filepath.Join("working/storyImages", filepath.Base(sentence.ImageURL))
			if _, err := os.Stat(imgPath); err == nil {
				pdf.Image(imgPath, 10, pdf.GetY(), 180, 0, false, "", 0, "")
			}
		}
		pdf.Ln(5)
		pdf.MultiCell(0, 10, sentence.CorrectedText, "", "L", false)
		pdf.Ln(10)
	}

	// Save the PDF
	outputPath := filepath.Join("working/comics", fmt.Sprintf("comic_%d.pdf", storyID))
	err = pdf.OutputFileAndClose(outputPath)
	if err != nil {
		return fmt.Errorf("failed to save PDF: %w", err)
	}

	// Save comic record in DB
	comic := model.Comic{
		UserID:      story.UserID,
		Title:       story.Title,
		Thumbnail:   generateThumbnail(sentences),
		ViewURL:     outputPath,
		DownloadURL: outputPath,
		DoneOn:      time.Now(),
	}
	err = s.storyRepo.SaveComic(&comic)
	if err != nil {
		return fmt.Errorf("failed to save comic record: %w", err)
	}

	return nil
}

func generateThumbnail(sentences []model.Sentence) string {
	for _, sentence := range sentences {
		if sentence.ImageURL != "" {
			return sentence.ImageURL // First image as thumbnail
		}
	}
	return ""
}
