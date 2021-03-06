/*
	Copyright 2019 Netfoundry, Inc.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package subcmd

import (
	"github.com/netfoundry/ziti-cmd/ziti/agent"
	"github.com/netfoundry/ziti-edge/controller/server"
	"github.com/netfoundry/ziti-fabric/controller"
	"github.com/spf13/cobra"
)


func init() {
	root.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run <config>",
	Short: "Run controller configuration",
	Args:  cobra.ExactArgs(1),
	Run:   run,
}

func run(cmd *cobra.Command, args []string) {
	if config, err := controller.LoadConfig(args[0]); err == nil {
		if cliAgentEnabled {
			if err := agent.Listen(agent.Options{}); err != nil {
				panic(err)
			}
		}

		var c *controller.Controller
		if c, err = controller.NewController(config); err != nil {
			panic(err)
		}

		ec, err := server.NewController(config)

		if err != nil {
			panic(err)
		}

		ec.SetHostController(c)

		go ec.InitAndRun()

		if err = c.Run(); err != nil {
			panic(err)
		}

	} else {
		panic(err)
	}
}
