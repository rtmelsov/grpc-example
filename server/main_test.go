package main

import (
	"context"
	pb "demo/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"log"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestMultiServer(t *testing.T) {
	conn, err := grpc.NewClient(":3200", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewStreamMultiServiceClient(conn)
	rand.Seed(time.Now().UnixNano())
	tests := []struct {
		num   int32
		limit int32
	}{
		{18, 12},
		{100, 7},
		{25, 10},
		// testing rand clue
		{rand.Int31(), rand.Int31n(50)},
		{rand.Int31(), rand.Int31n(50)},
		{rand.Int31(), rand.Int31n(50)},
	}

	var wg sync.WaitGroup

	// function for return result of test streaming
	want := func(initial, limit int32) (result int64) {
		for i := int64(1); i <= int64(limit); i++ {
			result += int64(initial) * i
		}
		return
	}

	for i := range tests {
		wg.Add(1)

		go func(i int) {
			stream, err := client.MultiResponse(context.Background(),
				&pb.Request{Num: tests[i].num, Limit: tests[i].limit})
			if err != nil {
				log.Fatal(err)
			}

			var get int64
			for {
				resp, err := stream.Recv()
				if err == io.EOF {
					wg.Done()
					break
				}
				if err != nil {
					log.Fatal(err)
				}
				get += resp.Result
			}
			if want := want(tests[i].num, tests[i].limit); get != want {
				t.Errorf("%d * [1..%d] expected %d, but get %d", tests[i].num, tests[i].limit, want, get)
			}
		}(i)
	}
	wg.Wait()
}
