# Project Title

Cloudfront Cache Invalidation API.

## Description

API written in Golang that implements the ability to automatically invalidate cloudfront cache for the supplied distribution and paths

## Getting Started

### Installing

* Uses the following environment variables setup:
```
Required:

API_KEY - Whatever key value you wish to use for authentication/authorization.
AWSAWS_SECRET_ACCESS_KEY - The secret key of the AWS credentials that will be used.
AWS_ACCESS_KEY_ID - The ID of the AWS credentails that will be used.
REDIS_HOST - Host for the redis data stored
REDIS_PORT - Port that the redis host listens on, default is 6379

Optional:
API_PORT - Port number the api will listen on.
```

### Running

API has 3 endpoints currently:

```
/api/health - Health check endpoint.
/invalidate - Main endpoint which receives a post request with a json body like so

    {
        "URL": "URL-OF-SITE",
        "PATHS": [
            "/path1",
            "/path2",
            "/path3"
        ]
    }

And returns the Cloudfront invalidation ID for that call.

Example result:

{"Invalidation_ID":"I1BQF7VF0OK7IUKM69P9CTUJS9"}

/status - This endpoint takes in the invalidation ID that the API returns in the /invalidate endpoint, along with the URL of the site and gets the status of it.

    {
        "ID": "<Invalidation ID that the other endpoint returns>",
        "URL": "URL-of-site"
    }

Example Result:

{"Status":"Completed"}

```

