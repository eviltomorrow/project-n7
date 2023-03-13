package snowflake

import (
	"fmt"

	"github.com/bwmarrin/snowflake"
	"github.com/eviltomorrow/project-n7/lib/helper"
)

var (
	machineID int64 = 1
	node      *snowflake.Node
)

func init() {
	snowflake.Epoch = helper.Runtime.LaunchTime.UnixNano() / 1e6
	var (
		err error
	)

	node, err = snowflake.NewNode(machineID)
	if err != nil {
		panic(fmt.Errorf("snowflake NewNode failure, nest error: %v", err))
	}
}

func Generate() snowflake.ID {
	return node.Generate()
}
