package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		time.Sleep(time.Second * 4)
		c.JSON(200, gin.H{"message": "ok"})
	})

	addr := "8081"
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", addr),
		Handler: r,
	}

	go func() {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Println("error received on ListenAnd Serve", err.Error())
		}
	}()

	sigs := make(chan os.Signal, 1)
	done := make(chan bool)

	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		done <- true
	}()

	<-done

	fmt.Println("Received a shutdown signal, quiting...")

	ctxTimeout, _ := context.WithTimeout(context.Background(), 10*time.Second)
	_ = server.Shutdown(ctxTimeout)

	fmt.Println("shutdown completed")
}
