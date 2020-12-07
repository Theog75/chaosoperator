#!/bin/bash

# VER should be the image tag
make docker-build IMG=docker.rct.co.il/chaoskube:$VER
make docker-push IMG=docker.rct.co.il/chaoskube:$VER
make deploy IMG=docker.rct.co.il/chaoskube:$VER
