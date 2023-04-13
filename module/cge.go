package module

import (
	"errors"

	"github.com/code-game-project/cge-parser/adapter"
	"github.com/code-game-project/cli-utils/components"
	"github.com/code-game-project/cli-utils/feedback"
	"github.com/code-game-project/cli-utils/server"
	"github.com/code-game-project/cli-utils/versions"
)

func LoadCGEData(gameURL string) (adapter.ParserResponse, error) {
	cgeFile, err := server.FetchCGEFile(gameURL)
	if err != nil {
		return adapter.ParserResponse{}, err
	}
	defer cgeFile.Close()
	metadata, cgeReader, diagnostics, err := adapter.ParseMetadata(cgeFile)
	if err != nil {
		for _, d := range diagnostics {
			if d.Type == adapter.DiagError {
				feedback.Error("cli-module", "CGE: %w", d.Message)
			}
		}
		return adapter.ParserResponse{}, err
	}
	parser, err := components.CGEParser(versions.MustParse(metadata.CGEVersion))
	if err != nil {
		return adapter.ParserResponse{}, err
	}
	response, errs := adapter.ParseCGE(cgeReader, parser, adapter.Config{
		IncludeComments: true,
	})
	if len(errs) > 0 {
		return adapter.ParserResponse{}, errors.Join(errs...)
	}
	return response, nil
}
