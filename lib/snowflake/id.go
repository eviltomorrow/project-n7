package snowflake

import (
	"fmt"

	"github.com/bwmarrin/snowflake"
)

var (
	machineID int64 = 1
	node      *snowflake.Node
)

func init() {
	n, err := snowflake.NewNode(machineID)
	if err != nil {
		panic(fmt.Errorf("snowflake NewNode failure, nest error: %v", err))
	}
	node = n
}

func Generate() snowflake.ID {
	return node.Generate()
}
