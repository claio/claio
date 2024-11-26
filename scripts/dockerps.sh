#!/bin/bash

docker ps --format 'table {{.ID}}\t{{.Names}}\t{{.Status}}'