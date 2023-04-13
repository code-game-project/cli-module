package module

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/code-game-project/cli-utils/feedback"
	"github.com/code-game-project/cli-utils/modules"
	"github.com/code-game-project/cli-utils/versions"
)

var (
	errUnsupportedAction      = errors.New("unsupported action")
	errUnsupportedProjectType = errors.New("unsupported project type")
)

type Config struct {
	Create func(data *modules.ActionCreateData) error

	name            string
	displayName     string
	libraryVersions map[modules.ProjectType][]versions.Version
}

func info(config Config) error {
	var ok bool
	libraryVersions := make(map[string][]versions.Version)
	libraryVersions["server"], ok = config.libraryVersions[modules.ProjectType_SERVER]
	if !ok {
		delete(libraryVersions, "server")
	}
	libraryVersions["client"], ok = config.libraryVersions[modules.ProjectType_CLIENT]
	if !ok {
		delete(libraryVersions, "client")
	}

	applicationTypes := make([]string, 0, len(libraryVersions))
	if libraryVersions["client"] != nil {
		applicationTypes = append(applicationTypes, "client")
	}
	if libraryVersions["server"] != nil {
		applicationTypes = append(applicationTypes, "server")
	}

	actions := make([]modules.Action, 0, 5)
	actions = append(actions, modules.ActionInfo)
	if config.Create != nil {
		actions = append(actions, modules.ActionCreate)
	}

	return json.NewEncoder(os.Stdout).Encode(modules.ModuleInfo{
		LibraryVersions:  libraryVersions,
		ApplicationTypes: applicationTypes,
		Actions:          actions,
	})
}

func Run(langName, langDisplayName string, libraryVersions map[modules.ProjectType][]versions.Version, config Config, minLogSeverity feedback.Severity) {
	feedback.Enable(feedback.NewCLIFeedback(minLogSeverity))
	config.name = langName
	config.displayName = langDisplayName
	config.libraryVersions = libraryVersions
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "USAGE: %s <action>\n", os.Args[0])
		os.Exit(1)
	}

	var err error
	switch modules.Action(os.Args[1]) {
	case modules.ActionInfo:
		err = info(config)
	case modules.ActionCreate:
		if config.Create == nil {
			err = errUnsupportedAction
		} else {
			data := modules.GetCreateData()
			if config.libraryVersions[data.ProjectType] == nil {
				err = errUnsupportedProjectType
			} else {
				err = config.Create(data)
			}
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
