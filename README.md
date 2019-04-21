# kfk
Kafka test task

## Requirements:
- Docker & docker-compose

## Deployment:
```
docker-compose up
```

## Testing
Test request to generate 1000 messages in 4 threads:
```
http://localhost:8003/generate/threads=4&requests=1000
```

## Consumers
"Visit" consumer:
```
http://localhost:8001
```
"Activity" consumer:
```
http://localhost:8002
```

## TODO
- [ ] Add configurational env variables for "send timeout" and "send batch size"
