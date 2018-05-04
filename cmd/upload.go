// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
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
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/rana/ora.v4"
)

var (
	dsn      string
	filename string
	filedesc string
	schema   string
	dl_id    int64
)

// uploadCmd represents the upload command
var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload file to Oracle DB",
	Long: `File to Oracle DB uploader.
This application is a tool to upload files to Oracle DB.`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Work your own magic here
		if (dsn == "") || (filename == "") {
			fmt.Println("Error:", "The Upload command demands presence of parameters")
			os.Exit(-1)
		}

		if err := upload(); err != nil {
			fmt.Println("Error:", err)
			os.Exit(-1)
		}

		fmt.Println("upload completed")
	},
}

func init() {
	RootCmd.AddCommand(uploadCmd)
	uploadCmd.PersistentFlags().StringVar(&dsn, "dsn", "", "Username/Password@ConnStr")
	uploadCmd.PersistentFlags().StringVar(&filename, "file", "", "Name of file")
	uploadCmd.PersistentFlags().StringVar(&filedesc, "desc", "", "Description of file")
	uploadCmd.PersistentFlags().StringVar(&schema, "schema", "", "Oracle schema")
	uploadCmd.PersistentFlags().Int64Var(&dl_id, "dl_id", -1, "Dealer ID")

}

func upload() error {

	filepath, filename, filedesc := func() (string, string, string) {
		_, name := path.Split(strings.Replace(filename, "\\", "/", -1))
		ext := path.Ext(name)
		name = name[:len(name)-len(ext)]

		if filedesc == "" {
			parts := strings.Split(name, "~")
			if len(parts) < 2 {
				filedesc = parts[0]
			} else {
				filedesc = parts[1]
			}
			name = parts[0] + ext
		}
		return filename, name, filedesc

	}()
	//fmt.Println("filepath =", filepath, " filename =", filename, " filedesc =", filedesc)

	b, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
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
	dl_id_Var := ora.Int64{true, 0}
	if dl_id != -1 {
		dl_id_Var.Value = dl_id
		dl_id_Var.IsNull = false
	}
	schemaVar := ora.String{true, ""}
	if schema != "" {
		schemaVar.IsNull = false
		schemaVar.Value = schema
	}
	body := ora.Lob{Reader: bytes.NewReader(b)}

	_, err = ses.PrepAndExe(stm, body, schemaVar, filename, filedesc, dl_id_Var)
	return err

}

const stm = `begin
  -- Call the procedure
  storm_base.pkg_reports.e_load_template_b(p_body => :1,
                                           p_schema_name => :2,
                                           p_file_name => :3,
                                           p_report_caption => :4,
                                           p_dl_id => :5
										);
end;
`
