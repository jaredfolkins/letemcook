#!/bin/bash

if [ ! -d "/lemc/public" ]; then
  mkdir -p "/lemc/public"
fi

if [ ! -d "/lemc/private" ]; then
  mkdir -p "/lemc/private"
fi

rm -rf /lemc/public/file-*
#rm -rf /lemc/private/*

timestamp=$(date +%s)
output="$timestamp - some string in some file"
filename="file-$timestamp.txt"
echo $output > /lemc/public/$filename
echo "lemc.html.trunc; "
echo "lemc.js.trunc; "

# Construct the link using the provided base URL and the filename
link="${LEMC_HTTP_DOWNLOAD_BASE_URL}/${filename}"

echo "lemc.css.trunc; #$LEMC_HTML_ID { font-family: Arial, sans-serif; }"
echo "lemc.html.append; <a id='ricky' class='btn' style='color: blue' data-isfired=0 data-ricky='https://www.youtube.com/watch?v=dQw4w9WgXcQ' data-link='$link'>download</a><br>"
echo "lemc.html.append; <div id='video-container'></div>"

single=$(cat ./script.js | tr -d '\n')
echo "lemc.js.exec; $single"
