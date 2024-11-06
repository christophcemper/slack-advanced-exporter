package cmd

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// FetchOperation defines what operations to perform during zip processing
type FetchOperation struct {
	Name         string
	ApiToken     string
	ProcessFile  func(file *zip.File, w *zip.Writer, token string) error
	TargetFiles  map[string]bool // files to process, empty means process all
	ApiRateLimit time.Duration   // time to wait between API calls
}

// ProcessZipArchive handles the common zip processing logic for all fetch operations
func ProcessZipArchive(ops ...FetchOperation) error {
	// Open the input archive
	r, err := zip.OpenReader(inputArchive)
	if err != nil {
		return fmt.Errorf("input archive error: %w", err)
	}
	defer r.Close()

	// Open the output archive
	f, err := os.Create(outputArchive)
	if err != nil {
		return fmt.Errorf("output archive error: %w", err)
	}
	defer f.Close()

	w := zip.NewWriter(f)
	defer w.Close()

	var processingErrors []string
	processedFiles := make(map[string]bool)

	// Process all files in the archive
	for _, file := range r.File {
		verbosePrintln(fmt.Sprintf("Processing file: %s\n", file.Name))

		// Find matching operations for this file
		for _, op := range ops {
			// Skip if this file isn't targeted by this operation
			if len(op.TargetFiles) > 0 && !op.TargetFiles[file.Name] {
				continue
			}

			if processedFiles[file.Name] {
				continue // Skip if already processed
			}

			if err := processFileWithRetry(file, w, op); err != nil {
				processingErrors = append(processingErrors,
					fmt.Sprintf("[%s] Failed processing %s: %v", op.Name, file.Name, err))
				continue
			}

			processedFiles[file.Name] = true
			time.Sleep(op.ApiRateLimit) // Respect rate limits
		}

		// If no operation processed this file, copy it as-is
		if !processedFiles[file.Name] {
			if err := copyFileToOutput(file, w); err != nil {
				processingErrors = append(processingErrors,
					fmt.Sprintf("Failed copying %s: %v", file.Name, err))
			}
		}
	}

	if len(processingErrors) > 0 {
		return fmt.Errorf("encountered %d errors:\n%s",
			len(processingErrors), strings.Join(processingErrors, "\n"))
	}

	return nil
}

// Example usage combining multiple operations
func fetchCombined(cmd *cobra.Command, args []string) error {
	ops := []FetchOperation{
		{
			Name:     "emails",
			ApiToken: emailsApiToken,
			TargetFiles: map[string]bool{
				"users.json": true,
			},
			ProcessFile: func(file *zip.File, w *zip.Writer, token string) error {
				// Existing email fetch logic
				// return processUsersJson(w, file, token)
				return nil
			},
			ApiRateLimit: time.Second, // 1 request per second
		},
		{
			Name:     "profile-pics",
			ApiToken: "NA", // not used for this step
			TargetFiles: map[string]bool{
				"users.json": true,
			},
			ProcessFile: func(file *zip.File, w *zip.Writer, token string) error {
				// Existing profile pics logic
				return processProfilePics(w, file, token)
			},
			ApiRateLimit: time.Second * 2,
		},
		{
			Name:     "attachments",
			ApiToken: attachmentsApiToken,
			ProcessFile: func(file *zip.File, w *zip.Writer, token string) error {
				// Existing attachments logic
				return processAttachment(w, file, token)
			},
			ApiRateLimit: time.Millisecond * 500,
		},
	}

	return ProcessZipArchive(ops...)
}

// processFileWithRetry attempts to process a file with retries on failure
func processFileWithRetry(file *zip.File, w *zip.Writer, op FetchOperation) error {
	maxRetries := 3
	baseDelay := time.Second
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			backoffDuration := baseDelay * time.Duration(1<<uint(attempt))
			fmt.Printf("attempt=%d, shift=%d, duration=%v\n",
				attempt, 1<<uint(attempt), backoffDuration)
			time.Sleep(backoffDuration)
		}
		if err := op.ProcessFile(file, w, op.ApiToken); err != nil {
			lastErr = err
			verbosePrintln(fmt.Sprintf("Attempt %d failed for %s: %v\n", attempt+1, file.Name, err))
			continue
		}
		return nil // Success
	}

	return fmt.Errorf("failed after %d attempts: %v", maxRetries, lastErr)
}

// copyFileToOutput copies a file from input zip to output zip without modification
func copyFileToOutput(file *zip.File, w *zip.Writer) error {
	reader, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer reader.Close()

	writer, err := w.Create(file.Name)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}

	_, err = io.Copy(writer, reader)
	if err != nil {
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	return nil
}

// Helper functions for specific processing tasks
func processProfilePics(w *zip.Writer, file *zip.File, token string) error {
	// Implementation depends on your specific needs for profile pics
	// This is just a placeholder
	return fmt.Errorf("processProfilePics not implemented")
}

func processAttachment(w *zip.Writer, file *zip.File, token string) error {
	// Implementation depends on your specific needs for attachments
	// This is just a placeholder
	return fmt.Errorf("processAttachment not implemented")
}
