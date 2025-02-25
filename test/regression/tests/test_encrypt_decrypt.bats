load '../helpers/bats-support/load'
load '../helpers/bats-assert/load'

setup() {
    export TESTFILE="hugefile.bin"
    dd if=/dev/zero of=$TESTFILE bs=1M count=100
}

teardown() {
    rm $TESTFILE
}

#TODO add some tests to make sure the file is encrypted on GCS. maybe check contetn lenght