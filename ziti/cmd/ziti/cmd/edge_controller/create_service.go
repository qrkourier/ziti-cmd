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

package edge_controller

import (
	"fmt"
	"io"
	"strconv"

	"github.com/Jeffail/gabs"
	"github.com/netfoundry/ziti-cmd/ziti/cmd/ziti/cmd/common"
	cmdutil "github.com/netfoundry/ziti-cmd/ziti/cmd/ziti/cmd/factory"
	cmdhelper "github.com/netfoundry/ziti-cmd/ziti/cmd/ziti/cmd/helpers"
	"github.com/spf13/cobra"
)

type createServiceOptions struct {
	commonOptions
	hostedService   bool
	tags            map[string]string
	edgeRouterRoles []string
	roleAttributes  []string
}

// newCreateServiceCmd creates the 'edge controller create service local' command for the given entity type
func newCreateServiceCmd(f cmdutil.Factory, out io.Writer, errOut io.Writer) *cobra.Command {
	options := &createServiceOptions{
		commonOptions: commonOptions{
			CommonOptions: common.CommonOptions{
				Factory: f,
				Out:     out,
				Err:     errOut,
			},
		},
		tags: make(map[string]string),
	}

	cmd := &cobra.Command{
		Use:   "service <name> <dns host> <dns port> [egress node]? [egress endpoint uri]?",
		Short: "creates a service managed by the Ziti Edge Controller",
		Long:  "creates a service managed by the Ziti Edge Controller",
		Args:  cobra.MinimumNArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			options.Cmd = cmd
			options.Args = args
			err := runCreateService(options)
			cmdhelper.CheckErr(err)
		},
		SuggestFor: []string{},
	}

	// allow interspersing positional args and flags
	cmd.Flags().SetInterspersed(true)
	cmd.Flags().StringToStringVarP(&options.tags, "tags", "t", nil, "Add tags to service definition")
	cmd.Flags().BoolVarP(&options.OutputJSONResponse, "output-json", "j", false, "Output the full JSON response from the Ziti Edge Controller")
	cmd.Flags().BoolVar(&options.hostedService, "hosted", false, "Indicates that this is a hosted service")
	cmd.Flags().StringSliceVarP(&options.edgeRouterRoles, "edge-router-roles", "r", nil, "Edge router roles of the new service")
	cmd.Flags().StringSliceVarP(&options.roleAttributes, "role-attributes", "a", nil, "Role attributes of the new identity")

	return cmd
}

// runCreateNativeService implements the command to create a service
func runCreateService(o *createServiceOptions) (err error) {
	var port int
	if port, err = strconv.Atoi(o.Args[2]); err != nil {
		return err
	}

	entityData := gabs.New()
	setJSONValue(entityData, o.Args[0], "name")
	setJSONValue(entityData, o.edgeRouterRoles, "edgeRouterRoles")
	setJSONValue(entityData, o.Args[1], "dns", "hostname")
	setJSONValue(entityData, port, "dns", "port")
	setJSONValue(entityData, o.roleAttributes, "roleAttributes")

	if o.hostedService {
		setJSONValue(entityData, "unclaimed", "egressRouter")
		setJSONValue(entityData, "hosted:unclaimed", "endpointAddress")
	} else {
		setJSONValue(entityData, o.Args[3], "egressRouter")
		setJSONValue(entityData, o.Args[4], "endpointAddress")
	}

	setJSONValue(entityData, o.tags, "tags")

	result, err := createEntityOfType("services", entityData.String(), &o.commonOptions)

	if err != nil {
		panic(err)
	}

	serviceId := result.S("data", "id").Data()

	if _, err = fmt.Fprintf(o.Out, "%v\n", serviceId); err != nil {
		panic(err)
	}

	return err
}
