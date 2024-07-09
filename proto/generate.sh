cd proto
buf generate
cd ..

cp -r github.com/noble-assets/noble/v6/* ./
rm -rf github.com
