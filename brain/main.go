package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/JustinYeoh93/etl/brain/cache"
	"github.com/JustinYeoh93/etl/brain/db"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/gofiber/fiber/v2"
)

const (
	SLEEPTIME = 30 * time.Second
	QUEUEURL  = "https://sqs.ap-northeast-1.amazonaws.com/634023999481/shineos-poc"
)

type CachePost struct {
	Id       string `json:"id"`
	Scraping bool   `json:"scraping"`
}

var scrapeCache = cache.Cache{}

func main() {
	crds, ok := os.LookupEnv("SHOPIFY_CREDENTIALS")
	if !ok {
		panic("SHOPIFY_CREDENTIALS not found")
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic(err)
	}
	client := sqs.NewFromConfig(cfg)

	go startScrape(client, crds)

	app := fiber.New()
	app.Post("/", func(c *fiber.Ctx) error {
		var cp CachePost
		if err := c.BodyParser(&cp); err != nil {
			return err
		}

		_, ok := scrapeCache[cp.Id]
		if !ok {
			log.Printf("unable to find id %s", cp.Id)
			return c.SendString(fmt.Sprintf("unable to find id %s", cp.Id))
		}

		scrapeCache[cp.Id] = cp.Scraping

		log.Printf("update cache %s", cp.Id)
		return c.JSON(scrapeCache)
	})

	app.Listen(":8080")
}

func startScrape(client *sqs.Client, crds string) {
	sources := db.NewSource(crds)

	for _, source := range sources {
		scrapeCache[source.ID] = false
	}

	// For every 30 seconds
	for {
		for _, source := range sources {
			if !isScraping(scrapeCache, source.ID) {
				byte, err := json.Marshal(source)
				if err != nil {
					log.Println(err)
				}
				err = appendToQueue(client, QUEUEURL, string(byte))
				if err != nil {
					log.Println(err)
				}
				scrapeCache[source.ID] = true
			}
		}

		// Sleep until next cycle
		time.Sleep(SLEEPTIME)
	}
}

func isScraping(cache cache.Cache, key string) bool {
	return cache[key]
}

func appendToQueue(client *sqs.Client, url string, detail string) error {
	_, err := client.SendMessage(context.TODO(), &sqs.SendMessageInput{
		MessageBody: aws.String(detail),
		QueueUrl:    aws.String(url),
	})

	if err != nil {
		return err
	}

	log.Println("appended a new message to queue")

	return nil
}
