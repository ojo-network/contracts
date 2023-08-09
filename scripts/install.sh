tmp_dir=$( mktemp -d)
cd $tmp_dir

git clone https://github.com/CosmWasm/wasmd.git
cd wasmd
git fetch --tags
git checkout v0.40.2

make install
echo "wasmd version" $(wasmd version)