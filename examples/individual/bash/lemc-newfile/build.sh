#!/bin/bash
IMAGE_NAME="docker.io/jfolkins/lemc-newfile:latest"
docker build -t $IMAGE_NAME .
docker push $IMAGE_NAME

docker build -t lemc-newfile .