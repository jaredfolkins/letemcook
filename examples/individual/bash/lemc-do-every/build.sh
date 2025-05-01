#!/bin/bash
IMAGE_NAME="docker.io/jfolkins/lemc-do-every:latest"
docker build -t $IMAGE_NAME .
docker push $IMAGE_NAME