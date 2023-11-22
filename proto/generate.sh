cd proto
buf generate
cd ..

cp -r github.com/noble-assets/noble/v5/* ./
rm -rf github.com
