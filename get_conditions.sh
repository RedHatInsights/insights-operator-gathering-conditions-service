#!/bin/bash

# Clone the conditions repo and build it to gather the conditions
if [ ! -d 'insights-operator-gathering-conditions' ]; then git clone https://github.com/RedHatInsights/insights-operator-gathering-conditions; fi
mkdir -p conditions
mkdir -p remote-configurations
cd insights-operator-gathering-conditions
# Retrieve all versions of conditions equal or greater than 1.1.0 (older versions lack rapid recommendations data)
for tag in `git tag --contains eb53ea55da02f87dc6e77a75c8c8ecee9cf41d8b` ; \
do \
	git checkout ${tag} && \
	./build.sh && \
	cp -r build/v1 ../conditions/${tag} && \
	cp -r build/v2 ../remote-configurations/${tag} && \
	rm -r build ; \
done
