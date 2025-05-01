#!/bin/bash

if [ ! -d "/lemc/public" ]; then
  mkdir -p "/lemc/public"
fi

if [ ! -d "/lemc/private" ]; then
  mkdir -p "/lemc/private"
fi

rm -rf /lemc/public/*
rm -rf /lemc/private/*

timestamp=$(date +%s)
output="$timestamp - some string in some file"
filename="file-$timestamp.txt"

echo $output > /lemc/public/$filename

echo "lemc.html.trunc; "

# This is a way to print all environment variables
#env | while IFS= read -r line; do
#  echo "lemc.html.append; $line<br>"
#done

script="alert(1);"

single=$(echo $script | tr -d '\n')
echo "lemc.js.exec;  $single"
echo "lemc.js.trunc;"
