package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type scenario struct {
	name                 string
	serverTaskDuration   time.Duration
	clientTimeout        time.Duration
	stopWhenClientCancel bool
}

var scenarios = []scenario{
	{
		name:                 "server task duration > client timeout",
		serverTaskDuration:   5 * time.Second,
		clientTimeout:        10 * time.Second,
		stopWhenClientCancel: false,
	},
	{
		name:                 "server task duration < client timeout, not stop when client cancel",
		serverTaskDuration:   5 * time.Second,
		clientTimeout:        2 * time.Second,
		stopWhenClientCancel: false,
	},
	{
		name:                 "server task duration > client timeout, stop when client cancel",
		serverTaskDuration:   5 * time.Second,
		clientTimeout:        2 * time.Second,
		stopWhenClientCancel: true,
	},
}

var selectedScenario scenario

func main() {

	sceneIdx, _ := strconv.Atoi(os.Args[1])
	selectedScenario = scenarios[sceneIdx]
	fmt.Printf("selected scenario: %s\n", selectedScenario.name)

	blocker := make(chan struct{})

	// server
	go func() {
		gin.SetMode(gin.ReleaseMode)
		r := gin.Default()
		r.GET("/", handler)
		r.Run("0.0.0.0:8080")
	}()

	time.Sleep(2 * time.Second)

	// client
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), selectedScenario.clientTimeout)
		defer func() {
			cancel()
			fmt.Println("client already canceled")
		}()

		fmt.Println("client start")
		req, err := http.NewRequest(http.MethodGet, "http://localhost:8080", nil)
		if err != nil {
			fmt.Println(err)
			return
		}

		req = req.WithContext(ctx)
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("client.Do error: %v\n", err)
			return
		}
		defer resp.Body.Close()
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("io.ReadAll error: %v\n", err)
			return
		}
		fmt.Println("response body:", string(b))
		fmt.Println("response status:", resp.Status)
	}()

	<-blocker
}

func handler(c *gin.Context) {
	if err := task(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "success",
	})
}

func task(ctx context.Context) error {
	end := time.Now().Add(selectedScenario.serverTaskDuration)
	for {
		fmt.Println("server task running...")
		select {
		case <-ctx.Done():
			if selectedScenario.stopWhenClientCancel {
				fmt.Println("server task canceled")
				return ctx.Err()
			} else if time.Now().After(end) {
				return nil
			}
		default:
			if time.Now().After(end) {
				return nil
			}
		}
		time.Sleep(1 * time.Second)
	}
}
