cd proto
buf generate
cd ..

cp -r github.com/strangelove-ventures/noble/v4/* ./
rm -rf github.com
