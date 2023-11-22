cd proto
buf generate
cd ..

cp -r github.com/noble-assets/noble/v4/* ./
rm -rf github.com
