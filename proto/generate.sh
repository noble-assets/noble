cd proto
buf generate --template buf.gen.gogo.yaml
buf generate --template buf.gen.pulsar.yaml
cd ..

cp -r noble.xyz/* ./
cp -r api/noble/* api/

rm -rf noble.xyz
rm -rf api/noble
