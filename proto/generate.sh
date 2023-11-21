cd proto
buf generate
cd ..

cp -r github.com/strangelove-ventures/noble/v5/* ./
rm -rf github.com
