package main

import (
	"proj2/server"
	"os"
	"strconv"
	"fmt"
	"encoding/json"

)

const usage = `Usage: twitter <number of consumers>
<number of consumers> = the number of goroutines (i.e., consumers) to be part of the parallel version.`

func main() {
	if len(os.Args) == 0 {
		fmt.Println(usage)
		return
	}

	config := server.Config{Encoder : json.NewEncoder(os.Stdout), Decoder: json.NewDecoder(os.Stdin), ConsumersCount: 0}

	if len(os.Args[1:]) == 0 {
		config.Mode = "s"
	} else {
		count, err := strconv.Atoi(os.Args[1])
		if err != nil {
			return
		}
		config.ConsumersCount = count
		config.Mode = "p"
	}
	server.Run(config)
}
