
set -x
export T=$1

if [[ $# -eq 0 ]];
 then
 die "usage: ./state.sh g"
fi


while true; do
echo "deploys.test.state:$(( ( RANDOM % 2 )  + 0 ))|$T" | nc -w 1 -u 172.30.168.10 8125
sleep 5
done

