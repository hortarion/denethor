#!/bin/bash

start_or_run () {
    docker inspect denethor_rabbitmq > /dev/null 2>&1

    if [ $? -eq 0 ]; then
        echo "Starting Denethor RabbitMQ container..."
        docker start denethor_rabbitmq
    else
        echo "Denethor RabbitMQ container not found, creating a new one..."
        docker run -d --name denethor_rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3.13-management
    fi
}

case "$1" in
    start)
        start_or_run
        ;;
    stop)
        echo "Stopping Denethor RabbitMQ container..."
        docker stop denethor_rabbitmq
        ;;
    logs)
        echo "Fetching logs for Denethor RabbitMQ container..."
        docker logs -f denethor_rabbitmq
        ;;
    *)
        echo "Usage: $0 {start|stop|logs}"
        exit 1
esac
