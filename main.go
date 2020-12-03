package main

import (
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"sort"
	"talk/global"
	"talk/request"
)

func main() {
	// 分发对应函数
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:    "class",
				Usage:   "爬一爬课程",
				Action:  func(c *cli.Context) error {
					request.RedisClient()
					request.HandleClass()
					defer global.REDIS_CLIENT.Close()
					return nil
				},
			},
		},
	}

	sort.Sort(cli.CommandsByName(app.Commands)) // 通过命令函数来排序，在help中进行展示

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
