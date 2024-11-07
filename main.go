package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/Mostafapy/wpool/work"
)

func main() {
	wp, err := work.NewPool(5, 5)

	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wp.Start(ctx)

	for range 20 {
		task := work.NewTask(func() error {
			const urlString = "https://google.com"
			res, err := http.Get(urlString)
			if err != nil {
				return err
			}

			fmt.Printf("%s returned status code %d\n", urlString, res.StatusCode)
			return nil
		}, func(err error) {
			fmt.Println(err)
		})

		wp.AddNonBlockingTask(task)
	}

	counter := 0

	for completed := range wp.TasksCompleted() {
		if completed {
			counter++
		}

		if counter == 20 {
			wp.Stop()
			return
		}
	}
}
