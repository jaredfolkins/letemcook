#!/bin/bash
IMAGE_NAME="docker.io/jfolkins/lemc-env-set-vars:latest"
docker build -t $IMAGE_NAME .
docker push $IMAGE_NAME

# The  build.sh  script is a simple shell script that uses the  docker build  command to build the Docker image.
# The  -t  flag is used to tag the image with the name  echo .
# The  .  at the end of the command specifies the build context, which is the current directory.
# The  Dockerfile  is in the same directory as the  build.sh  script, so the build context is set to the current directory.
# The  Dockerfile  is used to build the Docker image.
# The  Dockerfile  is a text file that contains a series of instructions that are used to build the Docker image.
# The  Dockerfile  for the  echo  image is shown below: