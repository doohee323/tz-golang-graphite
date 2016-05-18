
set -x
export T=$1

lat[0]="ord2"
lat[1]="sin1"
lat[2]="sin2"
lat[3]="sjc1"
lat[4]="sjc2"
lat[5]="syd1"
lat[6]="syd2"
lat[7]="yvr1"
lat[8]="yvr2"
lat[9]="nuq1"
lat[10]="nuq2"

if [[ $# -eq 0 ]];
 then
 die "usage: ./count2.sh g"
fi

while true; do
echo "deploys.test4.${lat[$[ $RANDOM % 3 ]]}.count2:$(( ( RANDOM % 5 )  + 0 ))|$T" | nc -w 1 -u 172.30.168.10 8125
sleep 1
done

exit 0;

lat[0]="fra1"
lat[1]="gru1"
lat[2]="gru2"
lat[3]="lhr1"
lat[4]="lhr2"
lat[5]="lhr3"
lat[6]="mia1"
lat[7]="mia2"
lat[8]="nrt1"
lat[9]="nrt2"
lat[10]="ord1"
lat[11]="ord2"
lat[12]="sin1"
lat[13]="sin2"
lat[14]="sjc1"
lat[15]="sjc2"
lat[16]="syd1"
lat[17]="syd2"
lat[18]="yvr1"
lat[19]="yvr2"
lat[20]="nuq1"
lat[21]="nuq2"


