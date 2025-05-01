#!/bin/bash

ip=$(curl -s icanhazip.com)

echo -e "lemc.html.trunc; The public IP address of this lemc instance is $ip"