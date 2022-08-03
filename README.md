Assume we are using a database as a permanent store for jobs. A job is identified when the ttl expires (Ex: ttl is 7/21/22 00:00, job expires on 7/21/22 00:01).

```
CREATE TABLE jobs (
	id SERIAL PRIVATE KEY,
	resource_id TEXT NOT NULL,
	ttl TIMESTAMP NOT NULL
);
```

We want a worker service that batch processes X jobs on a schedule - A scheduler that accumulates requests and performs batch processing on a number of incoming requests.

Processing consists of calling the endpoint `http://api.dummy.co/:resource_id` and then updating the ttl according to the response from the api - Response header contains the ttl for each resource id. To reiterate, the ttl is dynamically returned by the api - as ttls get updated in the database, the response .

As a flow chart:
[Get X Jobs from the database] -> [Call the api with the resource id] -> [Update the jobs ttl]

In a production environment we will run at all times at least 2 instances of the worker service.

Given:
```
type Job struct {
	Id int
	ResourceId string
	Ttl time.Time
}
type ApiResponse struct {
	NewTTL time.Time
}
```

TODO:
Write a sql query to get jobs from the database
Write a sql query to update job ttls
Write the method ProcessJobs() that does the following to accomplish the goal:
Invokes the method `func GetJobs() ([]Job, error)`. (Do not write GetJobs(), but assume this method is based on your aforementioned sql query)
Invokes the method `func CallApi(rid string) (ApiResponse, error)`. (Do not write this)
Invokes a method to update the jobs ttl. Come up with a function header as the header might change based on the implementation.
Handles failures that might occur during any call to the database or the api


Extension:

Given the api now acts as a broker between multiple weakly-consistent servers that may or may not contain data about a given resourceID. Instead of returning data about resource_id it now returns all servers that may or may not contain data about the given resource.

Calling `http://api.dummy.co/dealers/:resource_id` may return

{`http://api.buddy.co/:resource_id`, `http://api.howdy.co/:resource_id`}

Each one may or may not contain the NewTTL value.

The api response now looks like this:

```
type ApiResponse struct {
	DealerUrls []string
}
```

Given this change, how would you adapt `ProcessJobs()`


###
