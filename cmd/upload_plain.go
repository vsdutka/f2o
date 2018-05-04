// upload_plain
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
	script string
)

// uploadCmd represents the upload_plain command
var uploadPlainCmd = &cobra.Command{
	Use:   "upload_plain",
	Short: "Upload file to Oracle DB and exec script",
	Long: `File to Oracle DB uploader (plain).
This application is a tool to upload files to Oracle DB and execute script.`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Work your own magic here
		if (dsn == "") || (filename == "") {
			fmt.Println("Error:", "The Upload command demands presence of parameters")
			os.Exit(-1)
		}

		if err := uploadPlain(); err != nil {
			fmt.Println("Error:", err)
			os.Exit(-1)
		}

		fmt.Println("upload completed")
	},
}

func init() {
	RootCmd.AddCommand(uploadPlainCmd)
	uploadPlainCmd.PersistentFlags().StringVar(&dsn, "dsn", "", "Username/Password@ConnStr")
	uploadPlainCmd.PersistentFlags().StringVar(&filename, "file", "", "Name of file")
	uploadPlainCmd.PersistentFlags().StringVar(&script, "script", "", "Script fo uploading")
}

func uploadPlain() error {

	filepath, filename := func() (string, string) {
		_, name := path.Split(strings.Replace(filename, "\\", "/", -1))
		//		ext := path.Ext(name)
		//		name = name[:len(name)-len(ext)]

		return filename, name
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

	body := ora.Lob{Reader: bytes.NewReader(b)}

	_, err = ses.PrepAndExe(script, body, filename)
	return err

}
