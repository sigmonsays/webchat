#!/bin/bash

kubectl delete secret docker-registry regcred
kubectl create secret docker-registry regcred \
   --docker-server=docker.grepped.org \
   --docker-username=admin \
   --docker-password=admin123
