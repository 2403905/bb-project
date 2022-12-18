## Assessment Task

##### The service implemented regarding an assessment task has three key elements with architecture below:

1. An API receives the callback requests from the tester_service.go and sends objects data to the kafka topic. The
   object data represents a list of the id given from a single callback.
2. The 'Object handler' works concurrently. It performs:

* the batch read of data objects from kafka topic
* removes the duplicates of the id
* call concurrently the objects endpoint to check the online status of each id.
* save the result to the database

3. The 'Cleanup' worker wakes up every second and deletes the records that were saved/updated more than 30 seconds ago.

##### Possible improvements:
Make three independent services 'API',  'Object handler', 'Cleanup handler'

* The  'API' could be scaled for availability and reliability purposes.
* The 'Object handler' could be scaled for scale throughput, availability, and reliability purposes.
* The 'Cleanup' worker could have more time out adjustments.

#### How to run the service in a docker
**Note:** The .env file is used to run locally without docker.
* Clone a repository.
* Make sure that you have the Docker and Docker-compose already installed.
* Go to the repo directory.
* Lunch a project via the command.
```bash
docker-compose up -d
```
* Lookup services logs
```bash
docker-compose logs -f --tail=100 service tester-service
```
or all logs
```bash
docker-compose logs -f --tail=100
```


#### Test
The purpose of test coverage is to show an example of usage the third party packages like 'gomock', 'assert' and 'convey'

Usage
```bash
go test ./...
```
