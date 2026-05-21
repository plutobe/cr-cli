package review

import (
	"context"
	"fmt"

	"cr-cli/internal/ai"
	"cr-cli/internal/git"
)

type Options struct {
	BatchSize int
	Rules     []string
}

type Engine struct {
	provider ai.Provider
	options  Options
}

type Result struct {
	FileResults []FileResult
	TotalFiles  int
	Batches     int
}

type FileResult struct {
	Filename string
	Language string
	Issues   []ai.Issue
}

func NewEngine(provider ai.Provider, opts Options) *Engine {
	if opts.BatchSize <= 0 {
		opts.BatchSize = 10
	}
	return &Engine{provider: provider, options: opts}
}

func (e *Engine) Review(ctx context.Context, files []git.FileChange) (*Result, error) {
	result := &Result{TotalFiles: len(files)}
	batches := e.splitBatches(files)
	result.Batches = len(batches)

	for i, batch := range batches {
		fmt.Printf("审查批次 %d/%d (%d 个文件)...\n", i+1, len(batches), len(batch))
		for _, file := range batch {
			resp, err := e.provider.Review(ctx, &ai.ReviewRequest{
				Language: file.Language,
				Filename: file.Filename,
				Diff:     file.Diff,
				Rules:    e.options.Rules,
			})
			if err != nil {
				return nil, fmt.Errorf("review %s: %w", file.Filename, err)
			}
			result.FileResults = append(result.FileResults, FileResult{
				Filename: file.Filename,
				Language: file.Language,
				Issues:   resp.Issues,
			})
		}
	}
	return result, nil
}

func (e *Engine) splitBatches(files []git.FileChange) [][]git.FileChange {
	var batches [][]git.FileChange
	for i := 0; i < len(files); i += e.options.BatchSize {
		end := i + e.options.BatchSize
		if end > len(files) {
			end = len(files)
		}
		batches = append(batches, files[i:end])
	}
	return batches
}
