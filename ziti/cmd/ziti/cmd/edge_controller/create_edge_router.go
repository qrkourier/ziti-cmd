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
	"io/ioutil"

	"github.com/Jeffail/gabs"
	"github.com/netfoundry/ziti-cmd/ziti/cmd/ziti/cmd/common"
	cmdutil "github.com/netfoundry/ziti-cmd/ziti/cmd/ziti/cmd/factory"
	cmdhelper "github.com/netfoundry/ziti-cmd/ziti/cmd/ziti/cmd/helpers"
	"github.com/spf13/cobra"
)

type createEdgeRouterOptions struct {
	commonOptions
	roleAttributes []string
	jwtOutputFile  string
}

func newCreateEdgeRouterCmd(f cmdutil.Factory, out io.Writer, errOut io.Writer) *cobra.Command {
	options := &createEdgeRouterOptions{
		commonOptions: commonOptions{
			CommonOptions: common.CommonOptions{Factory: f, Out: out, Err: errOut},
		},
	}

	cmd := &cobra.Command{
		Use:     "edge-router <name>",
		Aliases: []string{"gateway"},
		Short:   "creates an edge router managed by the Ziti Edge Controller",
		Long:    "creates an edge router managed by the Ziti Edge Controller",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			options.Cmd = cmd
			options.Args = args
			err := runCreateEdgeRouter(options)
			cmdhelper.CheckErr(err)
		},
		SuggestFor: []string{},
	}

	// allow interspersing positional args and flags
	cmd.Flags().SetInterspersed(true)
	cmd.Flags().StringSliceVarP(&options.roleAttributes, "role-attributes", "a", nil, "Role attributes of the new edge router")
	cmd.Flags().BoolVarP(&options.OutputJSONResponse, "output-json", "j", false, "Output the full JSON response from the Ziti Edge Controller")
	cmd.Flags().StringVarP(&options.jwtOutputFile, "jwt-output-file", "o", "", "File to which to output the JWT used for enrolling the edge router")
	return cmd
}

// runCreateEdgeRouter implements the command to create a gateway on the edge controller
func runCreateEdgeRouter(o *createEdgeRouterOptions) error {
	routerData := gabs.New()
	setJSONValue(routerData, o.Args[0], "name")
	setJSONValue(routerData, o.roleAttributes, "roleAttributes")

	result, err := createEntityOfType("edge-routers", routerData.String(), &o.commonOptions)

	if err != nil {
		return err
	}

	id := result.S("data", "id").Data().(string)

	if _, err = fmt.Fprintf(o.Out, "%v\n", id); err != nil {
		panic(err)
	}

	if o.jwtOutputFile != "" {
		if err := getEdgeRouterJwt(o, id); err != nil {
			return err
		}
	}
	return nil
}

func getEdgeRouterJwt(o *createEdgeRouterOptions, id string) error {
	list, err := listEntitiesOfType("edge-routers", &o.commonOptions)
	if err != nil {
		return err
	}

	var newRouter *gabs.Container
	for _, gw := range list {
		gwId := gw.Path("id").Data().(string)
		if gwId == id {
			newRouter = gw
			break
		}
	}

	if newRouter == nil {
		return fmt.Errorf("no error during edge router creation, but edge router with id %v not found... unable to extract JWT", id)
	}

	jwt := newRouter.Path("enrollmentJwt").Data().(string)
	if jwt == "" {
		return fmt.Errorf("enrollment JWT not present for new edge router")
	}

	if err := ioutil.WriteFile(o.jwtOutputFile, []byte(jwt), 0600); err != nil {
		fmt.Printf("Failed to write JWT to file(%v)\n", o.jwtOutputFile)
		return err
	}

	jwtExpiration := newRouter.Path("enrollmentExpiresAt").Data().(string)
	if jwtExpiration != "" {
		fmt.Printf("Enrollment expires at %v\n", jwtExpiration)
	}

	return err
}
