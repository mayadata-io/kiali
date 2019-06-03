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
yarn --frozen-lockfile --non-interactive || (echo 'package.json is not in sync with yarn.lock, check that you include yarn.lock' && false)
# yarn prettier --list-different
# yarn build:dev
# cd ../kiali