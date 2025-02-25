load '../helpers/bats-support/load'
load '../helpers/bats-assert/load'

setup() {
    export TESTFILE="resume_file.bin"
    export CONTENT="This is a resumable test File"
    echo $CONTENT > $TESTFILE
    echo '{"contentType": "text/plain"}'> $TESTFILE.meta
}

teardown() {
    rm $TESTFILE
    rm $TESTFILE.meta
}
    
#https://cloud.google.com/storage/docs/performing-resumable-uploads
@test "Test Single chunk upload - Step 1: POST" {
  skip
  local content_length=$(wc -c < $TESTFILE)
  export TOKEN=$(gcloud auth print-access-token)
  export TARGET="https://storage.googleapis.com/upload/storage/v1/b$BUCKET/o?uploadType=resumable&name=$TESTFILE"
  run curl -X POST \
    -H "Authorization: Bearer $(gcloud auth print-access-token)" \
    -H "Content-Length: $content_length" \
    $TARGET \
    --cacert "$CA_BUNDLE" \
    --proxy "$HTTPS_PROXY" 2>/dev/null 
  
  #        -H "Content-Type: application/json" \
  #--data-binary "$TESTFILE.meta" \

  echo $output
  assert_success    
}


#https://cloud.google.com/storage/docs/performing-resumable-uploads
@test "Test Single chunk upload - Step 2: PUT" {
    skip
    export TOKEN=$(gcloud auth print-access-token)
    
    run curl -X PUT --data-binary $TESTFILE https://storage.googleapis.com/$BUCKET/$TESTFILE \
            -H "Authorization: Bearer $(gcloud auth print-access-token)" \
            --cacert $CA_BUNDLE \
            --proxy $HTTPS_PROXY 2>/dev/null 
    assert_success    
}

@test "Test gsutil rm" {
  skip
  run gsutil rm gs://$BUCKET/$TESTFILE
  assert_success
}


# @test "Test gsutil unit tests" {
#   run gsutil test
#   assert_success
# }

