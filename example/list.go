//
// lsof list opened file+pid
// support linux only
//

package main

import (
	"fmt"
	"github.com/arriqaaq/lsof"
)

func main() {
	prefix := ""
	list, err := lsof.Lsof(prefix)
	if err != nil {
		fmt.Printf("Lsof failed: %s\n", err)
		return
	}
	fmt.Printf(" ----- FILE LIST(MAP)\n")
	for key, val := range list.File2PIDsMap() {
		fmt.Printf("%s(%s): %v\n", key,  val.PIDs)
	}
	fmt.Printf(" -----\n")
}
