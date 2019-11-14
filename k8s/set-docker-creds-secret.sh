#!/bin/bash

kubectl create secret docker-registry regcred \
   --docker-server=https://docker.grepped.org/v1/ \
   --docker-username=admin \
   --docker-password=admin123 \
   --docker-email=sig.lange@gmail.com
