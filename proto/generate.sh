cd proto
buf generate
cd ..

cp -r github.com/strangelove-ventures/noble/* ./
rm -rf github.com

swagger-combine ./docs/config.json -o ./docs/static/openapi.yml
rm -rf tmp-swagger-gen
