load 'helpers/bats-support/load'
load 'helpers/bats-assert/load'

setup() {
    export TESTFILE="hugefile.bin"
    dd if=/dev/zero of=$TESTFILE bs=1M count=100
}

teardown() {
    rm $TESTFILE
}

@test "Test 100MB Upload using gcloud" {
    run gcloud storage cp $TESTFILE gs://$BUCKET/$TESTFILE
    assert_success
    
    
    # make sure it is there using gcloud storgea
    run gcloud storage ls gs://$BUCKET/$TESTFILE
    ! assert_output ""

    #delete it
    run gcloud storage rm gs://$BUCKET/$TESTFILE
    assert_success
}
