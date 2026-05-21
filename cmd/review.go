package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"cr-cli/internal/ai"
	"cr-cli/internal/config"
	"cr-cli/internal/git"
	"cr-cli/internal/gitlab"
	"cr-cli/internal/i18n"
	"cr-cli/internal/output"
	"cr-cli/internal/review"

	"github.com/spf13/cobra"
)

var (
	commitFlag    string
	mrFlag        int
	branchFlag    string
	baseFlag      string
	batchSizeFlag int
	severityFlag  string
)

var reviewCmd = &cobra.Command{
	Use:   "review",
	Short: i18n.T("审查代码变更", "Review code changes"),
	Long: i18n.T(
		"自动检测代码变更或指定提交/MR 进行 AI 代码审查。",
		"Auto-detect code changes or review specific commits/MRs with AI.",
	),
	RunE: runReview,
}

func init() {
	reviewCmd.Flags().StringVar(&commitFlag, "commit", "", i18n.T("审查指定提交 (SHA 或 ref)", "Review specific commit (SHA or ref)"))
	reviewCmd.Flags().IntVar(&mrFlag, "mr", 0, i18n.T("审查指定 GitLab MR", "Review specific GitLab MR"))
	reviewCmd.Flags().StringVar(&branchFlag, "branch", "", i18n.T("审查指定分支", "Review specific branch"))
	reviewCmd.Flags().StringVar(&baseFlag, "base", "main", i18n.T("比较的基础分支", "Base branch for comparison"))
	reviewCmd.Flags().IntVar(&batchSizeFlag, "batch-size", 0, i18n.T("每批审查文件数 (覆盖配置)", "Files per batch (override config)"))
	reviewCmd.Flags().StringVar(&severityFlag, "severity", "", i18n.T("审查模式: blocking|warning (覆盖配置)", "Review mode: blocking|warning (override config)"))
	rootCmd.AddCommand(reviewCmd)
}

func runReview(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfigOrSetup()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	if severityFlag != "" {
		cfg.Review.Severity = severityFlag
	}
	if batchSizeFlag > 0 {
		cfg.Review.BatchSize = batchSizeFlag
	}

	repoDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working dir: %w", err)
	}

	// Get diff
	var diff string
	if mrFlag > 0 {
		if cfg.GitLab.URL == "" || cfg.GitLab.Token == "" {
			return fmt.Errorf("%s", i18n.T(
				"GitLab MR 审查需要在配置中设置 gitlab.url 和 gitlab.token",
				"GitLab MR review requires gitlab.url and gitlab.token in config",
			))
		}
		client := gitlab.NewClient(cfg.GitLab.URL, cfg.GitLab.Token)
		project, err := git.GetRemoteProject(repoDir)
		if err != nil {
			return fmt.Errorf("get project name: %w", err)
		}
		diff, err = client.GetMergeRequestDiff(project, mrFlag)
		if err != nil {
			return fmt.Errorf("get MR diff: %w", err)
		}
	} else {
		diffSource := git.DiffSource{
			Commit: commitFlag,
			Branch: branchFlag,
			Base:   baseFlag,
		}
		diff, err = git.GetDiff(repoDir, diffSource)
		if err != nil {
			return fmt.Errorf("get diff: %w", err)
		}
	}

	if diff == "" {
		fmt.Println(i18n.T("没有发现代码变更。", "No code changes found."))
		return nil
	}

	// Parse changeset
	files := git.ParseChangeset(diff, git.WithIgnorePatterns(cfg.Review.IgnorePatterns))
	if len(files) == 0 {
		fmt.Println(i18n.T("没有需要审查的文件。", "No files to review."))
		return nil
	}

	fmt.Printf(i18n.T("发现 %d 个文件需要审查。\n", "Found %d file(s) to review.\n"), len(files))

	// Create provider
	provider, err := createProvider(cfg)
	if err != nil {
		return fmt.Errorf("create provider: %w", err)
	}

	// Run review
	engine := review.NewEngine(provider, review.Options{
		BatchSize: cfg.Review.BatchSize,
		Rules:     cfg.Review.Rules,
	})

	ctx := context.Background()
	startTime := time.Now()
	result, err := engine.Review(ctx, files)
	if err != nil {
		return fmt.Errorf("review failed: %w", err)
	}

	// Build output
	outputResult := buildOutputResult(result)

	// Render terminal output
	renderer := output.NewTerminalRenderer(os.Stdout, cfg.Output.Color)
	renderer.Render(outputResult)

	// Send webhook if configured
	if cfg.Webhook.Enabled {
		sender := output.NewWebhookSender(cfg.Webhook.URL, cfg.Webhook.Type, cfg.Webhook.Secret)
		commitInfo := commitFlag
		if commitInfo == "" {
			commitInfo = "auto-detect"
		}
		if err := sender.Send(outputResult, repoDir, commitInfo, startTime); err != nil {
			fmt.Fprintf(os.Stderr, i18n.T("Webhook 发送失败: %v\n", "Webhook send failed: %v\n"), err)
		}
	}

	// Exit code
	if cfg.Review.Severity == "blocking" {
		for _, fr := range result.FileResults {
			for _, issue := range fr.Issues {
				if issue.Severity == "error" {
					os.Exit(1)
				}
			}
		}
	}

	return nil
}

func createProvider(cfg *config.Config) (ai.Provider, error) {
	switch cfg.Provider.Type {
	case "openai":
		return ai.NewOpenAIProvider(
			cfg.Provider.BaseURL,
			cfg.Provider.APIKey,
			cfg.Provider.Model,
			cfg.Provider.MaxTokens,
			cfg.Provider.Temperature,
			debug,
		), nil
	case "anthropic":
		return ai.NewAnthropicProvider(
			cfg.Provider.BaseURL,
			cfg.Provider.APIKey,
			cfg.Provider.Model,
			cfg.Provider.MaxTokens,
			cfg.Provider.Temperature,
			debug,
		), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", cfg.Provider.Type)
	}
}

func buildOutputResult(result *review.Result) *output.ReviewResult {
	out := &output.ReviewResult{}
	for _, fr := range result.FileResults {
		out.Files = append(out.Files, output.FileResult{
			Filename: fr.Filename,
			Language: fr.Language,
			Issues:   fr.Issues,
		})
		for _, issue := range fr.Issues {
			out.TotalIssues++
			switch issue.Severity {
			case "error":
				out.ErrorCount++
			case "warning":
				out.WarningCount++
			case "info":
				out.InfoCount++
			}
		}
	}
	return out
}
