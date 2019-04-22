# kfk
Kafka test task

## Requirements:
- Docker & docker-compose

## Deployment:
```
docker-compose up
```

## Testing
Test endpoint to generate 1000 requests in 4 threads:
```
http://localhost:8003/?threads=4&requests=1000
```
One request generates one "activity" and one "visit". So upper example will generate 2 * 1000 * 4 requests. Each thread (client) will have unique user id.

## Consumers
"Visit" consumer JSON files:
```
http://localhost:8001
```
"Activity" consumer JSON files:
```
http://localhost:8002
```

## TODO
- [ ] Add configurational env variables for "send timeout" and "send batch size"
