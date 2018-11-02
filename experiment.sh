#!/bin/bash

CLIENTS=("1-ff00:0:110,[127.0.0.1]:0"  "1-ff00:0:120,[127.0.0.3]:0"  "1-ff00:0:111,[127.0.0.4]:0" "1-ff00:0:111,[127.0.0.5]:0") 
#./bin/sibra_bandwidth -sciondFromIA -remote "2-ff00:0:210,[127.0.0.2]:4444" -local "1-ff00:0:112,[127.0.0.1]:0" -sibra=F -duration 65 -bw 30 -bandwidth 1000000 &

END=7
DURATION=60

for i in $(seq 1 $END); do
	echo "$i"

	./bin/sibra_bandwidth_quic -sciondFromIA -remote "2-ff00:0:210,[127.0.0.1]:3333" -local "1-ff00:0:110,[127.0.0.$i]:0" -duration "$DURATION" -bw 2 -packetSize 1200 -sibra=T &> /dev/null & 
done
	
./bin/sibra_bandwidth_quic -sciondFromIA -remote "2-ff00:0:210,[127.0.0.1]:3333" -local "1-ff00:0:110,[127.0.0.42]:0" -duration "$DURATION" -bw 8 -packetSize 1200 -sibra=T
