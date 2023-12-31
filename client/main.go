package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"sync/atomic"
	"time"

	"github.com/lixoi/system_stats_daemon/internal/server/grpc/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

var lastID uint64

func GetNextID() uint64 {
	curID := atomic.AddUint64(&lastID, 1)
	return curID
}

func GenerateActionID() string {
	curID := GetNextID()
	return fmt.Sprintf("%v:%v", time.Now().UTC().Format("20060102150405"), curID)
}

func GetDump() {
	conn, err := grpc.Dial(":8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	client := api.NewSystemStatisticsClient(conn)
	md := metadata.New(nil)
	md.Append("request_id", GenerateActionID())
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	dump, err := client.GetSystemDump(ctx, &api.GetSystemDumpRequest{N: 5, M: 20})
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	jsonDump, err := json.Marshal(dump)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(string(jsonDump))
}

func GetStreamDump() {
	conn, err := grpc.Dial(":8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	defer conn.Close()

	client := api.NewSystemStatisticsClient(conn)
	stream, err := client.StreamSystemDump(context.Background(), &api.GetSystemDumpRequest{N: 5, M: 20})
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	for {
		dump, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			fmt.Printf("%v.StreamSystemDump(_) = _, %v\n", client, err.Error())
			return
		}

		fmt.Println("")
		fmt.Println(time.Now())
		fmt.Println("\tTopTalkersProtocol:")
		fmt.Println("\tProtocol\tBytes\tRate")
		fmt.Println("\t----------------------------")
		for i := range dump.SystemDump.TT.Ttp {
			fmt.Printf("\t%s\t\t%d\t%d\n",
				dump.SystemDump.TT.Ttp[i].Protocol,
				dump.SystemDump.TT.Ttp[i].Bytes,
				dump.SystemDump.TT.Ttp[i].Rate)
		}
		fmt.Println("\n\tTopTalkersTraffic:")
		fmt.Println("\tProtocol\tSource\t\t\t\tDistination\t\t\tBps")
		fmt.Println("\t-----------------------------------------------------------------------------------")
		for i := range dump.SystemDump.TT.Ttt {
			fmt.Printf("\t%s\t\t%s\t\t%s\t%d\n",
				dump.SystemDump.TT.Ttt[i].Protocol,
				dump.SystemDump.TT.Ttt[i].Source,
				dump.SystemDump.TT.Ttt[i].Distination,
				dump.SystemDump.TT.Ttt[i].Bps)
		}
	}
}

func main() {
	GetStreamDump()
	// GetDump()
}
