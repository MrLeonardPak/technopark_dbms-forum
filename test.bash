# Description: Test script for park
docker image rm park
docker build --no-cache -t park . \
&&  docker run -it --rm --memory 2G --log-opt max-size=5M --log-opt max-file=3 --name park_perf -p 5000:5000 park
