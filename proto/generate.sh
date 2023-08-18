cd proto
buf generate
buf generate --template buf.gen.pulsar.yaml
cd ..

cp -r github.com/strangelove-ventures/noble/* ./
rm -rf github.com
rm -rf noble
