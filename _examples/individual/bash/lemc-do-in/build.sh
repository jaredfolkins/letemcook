#!/bin/bash
IMAGE_NAME="docker.io/jfolkins/lemc-do-in:latest"
docker build -t $IMAGE_NAME .
docker push $IMAGE_NAME