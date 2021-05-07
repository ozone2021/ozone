package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"net/rpc"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "test",
	Long:  `List running processes`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("testing")
		client, err := rpc.DialHTTP("tcp", ":8000")
		if err != nil {
			log.Fatal("dialing:", err)
		}
		a := 4
		err = client.Call("ProcessManager.Test", &a, nil)
		if err != nil {
			log.Fatal("arith error:", err)
		}
	},
}