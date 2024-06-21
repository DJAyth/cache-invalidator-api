package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudfront"

	log "github.com/sirupsen/logrus"
)

// The main function to invalidate the cloudfront cache of the distribution
func invalidate(w http.ResponseWriter, r *http.Request) {

	//AWS connection session
	sess, err := session.NewSession()
	if err != nil {
		panic(err)
	}

	var post Post

	//Set the content-type header if need be, and import/decode the json body of it.
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewDecoder(r.Body).Decode(&post)
	var paths []*string

	//Loop through the paths provided and create a *string array for it.
	for _, path := range post.PATHS {
		paths = append(paths, &path)
	}
	ID := getDistributionID(sess, post.URL)

	//The invalidation needs a Caller reference setup that needs to be unique, use the unixtimestamp to give us something unique.
	tag := strconv.FormatInt(time.Now().UTC().UnixNano(), 10)

	//Build the input for the invalidation
	svc := cloudfront.New(sess)
	input := &cloudfront.CreateInvalidationInput{
		DistributionId: aws.String(ID),
		InvalidationBatch: &cloudfront.InvalidationBatch{
			CallerReference: aws.String(
				fmt.Sprintf("invalidation-id-%s-%s", post.URL, tag)),
			Paths: &cloudfront.Paths{
				Quantity: aws.Int64(int64(len(paths))),
				Items:    paths,
			},
		},
	}

	// Run the actual invalidation, catch the error if there is one.
	result, err := svc.CreateInvalidation(input)
	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		log.Error(err.Error())
	}

	// Return a json response with the ID produced from the invalidation. Whatever is calling this can use that with the /status route to get the status of the invalidation.
	json.NewEncoder(w).Encode(map[string]*string{"Invalidation_ID": result.Invalidation.Id})
	return
}

// Function to consume an invalidation id that's produced from the /invalidate route.
func invalidation_status(w http.ResponseWriter, r *http.Request) {

	//AWS connection session
	sess, err := session.NewSession()
	if err != nil {
		panic(err)
	}
	svc := cloudfront.New(sess)

	//Decode and consume the json body of the call
	var status Status
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewDecoder(r.Body).Decode(&status)

	ID := getDistributionID(sess, status.URL)

	//Create the input and get the status of the invalidation of the ID provided.
	getInvalInput := &cloudfront.GetInvalidationInput{
		DistributionId: aws.String(ID),
		Id:             status.ID,
	}
	invalidationStatus, err := svc.GetInvalidation(getInvalInput)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"Error": "Incorrect Invalidation ID, or Distribution ID"})
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"Status": *invalidationStatus.Invalidation.Status})
	return
}
