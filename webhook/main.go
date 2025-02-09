package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"golang.ngrok.com/ngrok"
	"golang.ngrok.com/ngrok/config"
)

var (
	logPath = "history.log"
	port        = "8080"
	endpoint    = "/webhook"
	withTunnel  = false
)

func parseFlags() {
	flag.StringVar(&logPath, "l", logPath, "log path")
	flag.StringVar(&port, "p", port, "listen por")
	flag.StringVar(&endpoint, "e", endpoint, "webhook endpoint")
	flag.BoolVar(&withTunnel, "t", withTunnel, "running with ngrok tunnel")
	flag.Parse()
}

func main() {
	parseFlags()

	app := fiber.New()

	app.Use(recover.New())

	var logFileLock sync.Mutex

	app.All(endpoint, func(c *fiber.Ctx) error {
		logFileLock.Lock()
		defer logFileLock.Unlock()

		logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			panic(err)
		}
		defer logFile.Close()

		timestamp := time.Now().UTC()
		method := c.Method()
		headersData, _ := json.Marshal(c.GetReqHeaders())
		bodyData := c.Body()

		fmt.Fprintln(logFile, "+++++\n"+
			"Timestamp: "+fmt.Sprint(timestamp)+"\n"+
			"Method: "+method+"\n"+
			"Headers: "+string(headersData)+"\n"+
			"Body: "+string(bodyData)+"\n"+
			"-----\n",
		)

		logFile.Sync()

		return c.SendStatus(200)
	})

	if withTunnel {
		tunnel, err := ngrok.Listen(context.Background(), config.HTTPEndpoint())
		if err != nil {
			log.Fatal("failed to start ngrok tunnel")
		}
		app.Listener(tunnel)
    } else {
		app.Listen(":" + port)
	}
}
