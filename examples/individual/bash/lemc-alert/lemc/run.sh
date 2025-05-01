#!/bin/bash

echo "lemc.html.trunc; "
echo "lemc.css.trunc; "
echo "lemc.js.trunc; "
script="alert(1);"
single=$(echo $script | tr -d '\n')
echo "lemc.js.exec;  $single"
echo "lemc.js.trunc;"
