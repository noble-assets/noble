cd proto
buf generate
cd ..

cp -r github.com/noble-assets/noble/v7/* ./
rm -rf github.com
