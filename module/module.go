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
	Create    func(data *modules.ActionCreateData) error
	RunClient func(data *modules.ActionRunClientData) error
	RunServer func(data *modules.ActionRunServerData) error

	name            string
	version         versions.Version
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

	projectTypes := make([]string, 0, len(libraryVersions))
	if libraryVersions["client"] != nil {
		projectTypes = append(projectTypes, "client")
	}
	if libraryVersions["server"] != nil {
		projectTypes = append(projectTypes, "server")
	}

	actions := make([]modules.Action, 0, 5)
	actions = append(actions, modules.ActionInfo)
	if config.Create != nil {
		actions = append(actions, modules.ActionCreate)
	}
	if config.RunClient != nil {
		actions = append(actions, modules.ActionRunClient)
	}
	if config.RunServer != nil {
		actions = append(actions, modules.ActionRunServer)
	}

	return json.NewEncoder(os.Stdout).Encode(modules.ModuleInfo{
		Version:         config.version,
		LibraryVersions: libraryVersions,
		ProjectTypes:    projectTypes,
		Actions:         actions,
	})
}

func Run(langName, langDisplayName string, moduleVersion versions.Version, libraryVersions map[modules.ProjectType][]versions.Version, config Config, minLogSeverity feedback.Severity) {
	feedback.Enable(feedback.NewCLIFeedback(minLogSeverity))
	config.name = langName
	config.displayName = langDisplayName
	config.version = moduleVersion
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
	case modules.ActionRunClient:
		data := modules.GetRunClientData()
		if config.RunClient == nil {
			err = errUnsupportedAction
		} else {
			err = config.RunClient(data)
		}
	case modules.ActionRunServer:
		data := modules.GetRunServerData()
		if config.RunServer == nil {
			err = errUnsupportedAction
		} else {
			err = config.RunServer(data)
		}
	default:
		err = errors.New("unsupported action")
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
