package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/JustinYeoh93/etl/brain/db"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

const (
	QUEUEURL = "https://sqs.ap-northeast-1.amazonaws.com/634023999481/shineos-poc"
	BRAINURL = "http://localhost:8080/"
	DLQ      = "https://sqs.ap-northeast-1.amazonaws.com/634023999481/shineos-poc-dlq"
)

type CachePost struct {
	Id       string `json:"id"`
	Scraping bool   `json:"scraping"`
}

type Source struct {
	ID         string
	URL        string
	Credential string
}

func main() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic(err)
	}

	client := sqs.NewFromConfig(cfg)

	var s db.Source

	res, err := client.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(QUEUEURL),
		MaxNumberOfMessages: 1,
	})
	if err != nil {
		panic(err)
	}

	if len(res.Messages) == 0 {
		log.Println("no messages found")
		return
	}

	err = json.Unmarshal([]byte(*res.Messages[0].Body), &s)
	if err != nil {
		panic(err)
	}

	// Scrape Destination
	req, err := http.NewRequest(http.MethodGet, strings.Join([]string{s.URL, "orders.json?status=any"}, ""), nil)
	req.Header.Add("X-Shopify-Access-Token", s.Credential)
	req.Header.Add("Content-type", "application/json")

	sourceRes, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	// Ping that scrape is complete
	p := CachePost{
		Id:       s.ID,
		Scraping: false,
	}
	b, _ := json.Marshal(p)

	_, err = http.Post(BRAINURL, "application/json", bytes.NewBuffer(b))
	if err != nil {
		log.Println(err)
	}

	// Delete from queue
	client.DeleteMessage(context.TODO(), &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(QUEUEURL),
		ReceiptHandle: res.Messages[0].ReceiptHandle,
	})

	// Process data
	data, err := io.ReadAll(sourceRes.Body)
	if err != nil {
		panic(err)
	}

	// Send to DLQ if failed
	if sourceRes.StatusCode != http.StatusOK {
		client.SendMessage(context.TODO(), &sqs.SendMessageInput{
			MessageBody: aws.String(string(data)),
		})
	}

	// Process it
	f, err := os.Create(strings.Join([]string{s.ID, ".json"}, ""))
	if err != nil {
		panic(err)
	}

	f.Write(data)

	log.Println("complete scrape")
}
