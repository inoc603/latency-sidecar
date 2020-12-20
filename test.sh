echo "target ip: $1, agent port: $2"

ping $1 -c 3

echo "\nset latency to $3"
curl -XGET http://$1:$2/latency/$3
echo "\n"

ping $1 -c 3

echo "\nreset latency"
curl -XGET http://$1:$2/latency/0ms
echo "\n"

echo "ping target"
ping $1 -c 3
