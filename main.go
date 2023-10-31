package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		time.Sleep(time.Second * 4)
		c.JSON(200, gin.H{"message": "ok"})
	})

	chanError := make(chan error)

	go gracefullyShutdown(r, "8181", chanError)

	if err := <-chanError; err != nil {
		fmt.Println(err.Error())
	}

}

func gracefullyShutdown(handler http.Handler, addr string, chanError chan error) {

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", addr),
		Handler: handler,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-ctx.Done()
		fmt.Println("Received a shutdown signal, quiting...")

		shutdownTimeout := 10 * time.Second
		ctxTimeout, cancel := context.WithTimeout(context.Background(), shutdownTimeout)

		defer func() {
			stop()
			cancel()
			close(chanError)
		}()

		err := server.Shutdown(ctxTimeout)
		if err != nil {
			chanError <- fmt.Errorf("error on shutdown application: %v", err)
			return
		}
		fmt.Println("shutdown completed")

	}()

	go func() {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			chanError <- fmt.Errorf("error on start application: %v", err)
			return
		}
	}()

}
