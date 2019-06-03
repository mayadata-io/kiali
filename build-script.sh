echo pwd
KIALI_PATH=pwd
echo $KIALI_PATH
echo $GOPATH
echo $GO_BUILD_ENVVARS
# build kiali-ui
git clone https://github.com/mayadata-io/kiali-ui.git
git checout master
npm install -g yarn
yarn
yarn build