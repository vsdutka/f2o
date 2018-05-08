// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/rana/ora.v4"
)

var configFileName string

// uploadManyCmd represents the uploadMany command
var uploadManyCmd = &cobra.Command{
	Use:   "upload_many",
	Short: "Upload files to Oracle DB",
	Long:  `Upload files to Oracle DB via configuration file`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Work your own magic here
		if (dsn == "") || (configFileName == "") {
			fmt.Println("Error:", "The UploadMany command demands presence of parameters")
			os.Exit(-1)
		}

		if err := uploadMany(); err != nil {
			fmt.Println("Error:", err)
			os.Exit(-1)
		}

		fmt.Println("upload completed")
	},
}

func init() {
	RootCmd.AddCommand(uploadManyCmd)
	uploadManyCmd.PersistentFlags().StringVar(&dsn, "dsn", "", "Username/Password@ConnStr")
	uploadManyCmd.PersistentFlags().StringVar(&configFileName, "config", "", "Name of config file")
}

func uploadMany() error {
	type info struct {
		Name   string
		Schema string
		Dl_id  *int64
	}
	fmt.Println("Open configuration file ", configFileName)
	buf, err := ioutil.ReadFile(configFileName)
	if err != nil {
		return err
	}

	infos := make([]info, 0)
	if err = json.Unmarshal(buf, &infos); err != nil {
		return err
	}
	if len(infos) == 0 {
		return nil
	}

	env, srv, ses, err := ora.NewEnvSrvSes(dsn)
	defer func() {
		if ses != nil {
			ses.Close()
		}
		if srv != nil {
			srv.Close()
		}
		if env != nil {
			env.Close()
		}
	}()
	if err != nil {
		return err
	}

	for _, v := range infos {

		filepath, filename := func() (string, string) {
			_, name := path.Split(strings.Replace(v.Name, "\\", "/", -1))
			//		ext := path.Ext(name)
			//		name = name[:len(name)-len(ext)]

			return filename, name
		}()
		dl_id_Var := ora.Int64{true, 0}
		if v.Dl_id != nil {
			dl_id_Var.Value = *v.Dl_id
			dl_id_Var.IsNull = false
		}
		schemaVar := ora.String{true, ""}
		if v.Schema != "" {
			schemaVar.IsNull = false
			schemaVar.Value = v.Schema
		}

		b, err := ioutil.ReadFile(filepath)
		if err != nil {
			return err
		}

		body := ora.Lob{Reader: bytes.NewReader(b)}

		if _, err = ses.PrepAndExe(stm, body, schemaVar, filename, filedesc, dl_id_Var); err != nil {
			return err
		}
		fmt.Printf("File \"%s\" for scheme \"%s\" and DL_ID=%v successfully uploaded\n", filepath, v.Schema, *v.Dl_id)
	}
	return nil
}
