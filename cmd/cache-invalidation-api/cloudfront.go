package main

import (
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
)

var sess *session.Session

// Get the CF distributions ID based on the domain
func getCFDistributions() []string {

	//AWS connection session
	sess, err := session.NewSession()
	if err != nil {
		panic(err)
	}

	svc := cloudfront.New(sess)
	var cfData []string
	ID := ""
	for true {
		var result *cloudfront.ListDistributionsOutput
		// Get the CF dist list in chunks of 50
		input := &cloudfront.ListDistributionsInput{
			Marker:   aws.String(ID),
			MaxItems: aws.Int64(50),
		}
		result, err := svc.ListDistributions(input)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				case cloudfront.ErrCodeInvalidArgument:
					log.Error(cloudfront.ErrCodeInvalidArgument, aerr.Error())
				default:
					log.Error(aerr.Error())
				}
			} else {
				// Print the error, cast err to awserr.Error to get the Code and
				// Message from an error.
				log.Error(err.Error())
			}
			return cfData
		}

		// If no more distribution chunks left exit the loop
		if result.DistributionList.Items == nil {
			break
		}

		// For every distribution
		for _, d := range result.DistributionList.Items {
			// For every CNAME in a distribution
			for _, c := range d.Aliases.Items {
				cfData = append(cfData, *d.Id+","+*c)
			}
			// Paginating results, i.e. using ID as the marker to indicate where to begin in the list for the next call
			ID = *d.Id
		}
	}
	return cfData
}

func loadCloudfrontData(cfData []string, cfRdb *redis.Client) {
	for _, cf := range cfData {
		id := strings.Split(cf, ",")[0]
		alias := strings.Split(cf, ",")[1]
		err := cfRdb.Set(ctx, alias, id, 6*time.Hour).Err()
		log.Debug("Loading " + id + " into redis for site " + string(alias) + ".")
		if err != nil {
			panic(err)
		}
	}
}

// Lookup in redis, the CloudFront Distribution ID for the host/site
func getDistributionID(sess *session.Session, host string) string {

	rdb := redis.NewClient(&redis.Options{
		Addr:     redis_host + ":" + redis_port,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	cfID, err := rdb.Get(ctx, host).Result()
	if err == redis.Nil {
		log.Info("Unable to find Distribution ID in Redis for " + host + ".")
		log.Info("Attempting to load in new Cloudfront data.")
		var cfData []string
		cfData = getCFDistributions()
		loadCloudfrontData(cfData, rdb)
		cfID, err := rdb.Get(ctx, host).Result()
		if err == redis.Nil {
			log.Error("Unable to find Distribution ID for " + host + " after refreshing data. Giving up.")
			return ""
		}
		return cfID
	} else {
		log.Info("Found " + cfID + " for " + host + ".")
		return cfID
	}
}
