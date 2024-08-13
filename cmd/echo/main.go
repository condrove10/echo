package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/alessandrofascini/log4go"
	"github.com/condrove10/echo/internal/docker"
	"github.com/joho/godotenv"
)

func main() {
	if len(os.Args) == 2 {
		err := godotenv.Overload(os.Args[1])
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}

	if err := log4go.Configure(map[string]any{
    "appenders": map[string]any{
        "out": map[string]any{
            "type": "stdout",
            "layout": map[string]any{
                "type":    "pattern",
                "pattern": "%[%d\t%p\t%] [%X{origin}] [%X{id}] [%X{runtime}] ~ %m",
            },
            "channelSize": float64(20),
        },
        "app": map[string]any{
            "type":             "dateFile",
            "filename":         "log/log.log",
            "maxLogSize":       "100M",
            "keepFileExt":      true,
            "channelSize":      float64(160),
            "internalChannelSize": float64(120),
            "mode":             int(493),
            "flags":            int(511),
            "compress":         true,
            "compressMode":     "default",
            "fileNameSep":      ".",
            "backups":          int(30),
            "layout": map[string]any{
                "type":    "pattern",
                "pattern": "%[%d\t%p\t%] [%X{origin}] [%X{id}] [%X{runtime}] ~ %m",
            },
        },
    },
    "categories": map[string]any{
        "default": map[string]any{
            "appenders": []any{
                "out",
            },
            "level": "all",
        },
        "release": map[string]any{
            "appenders": []any{
                "out",
                "app",
            },
            "level": "info",
        },
    },
    "configuration": map[string]any{
        "appenderChannelSize":       float64(160),
        "internalAppenderChannelSize": float64(110),
        "createFolderPerm":          int(493),
    },
}); err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	if err := docker.Startup(ctx); err != nil {
		log.Fatal(err)
	}

	docker.Monitor(context.Background())
}
