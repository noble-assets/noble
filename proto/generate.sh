cd proto
buf generate
cd ..

<<<<<<< HEAD
cp -r github.com/strangelove-ventures/noble/v4/* ./
=======
cp -r github.com/noble-assets/noble/v5/* ./
>>>>>>> a4ad980 (chore: rename module path (#283))
rm -rf github.com
