echo pwd
KIALI_PATH=pwd
echo $KIALI_PATH
echo $GOPATH
echo $GO_BUILD_ENVVARS
# build kiali-ui
cd ../
git clone https://github.com/mayadata-io/kiali-ui.git
cd kiali-ui/
git checkout master
npm install -g yarn
yarn
yarn build
cd ../kiali