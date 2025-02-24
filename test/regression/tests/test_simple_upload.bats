load 'helpers/bats-support/load'
load 'helpers/bats-assert/load'

setup() {
  export TESTFILE="testfile_25.txt"
  # Create a temporary file with some content
  echo "This is a another reallyureallyureallyureallyureallyu long and prop file." > $TESTFILE
}

teardown() {
  # Remove the temporary file
  rm $TESTFILE
}

@test "Setup - gcloud storage cp" {
  run gcloud storage cp $TESTFILE gs://$BUCKET/$TESTFILE 
  assert_success
}

@test "Test GCS Metadata - verify content length" {
    local expected_size=$(wc -c < $TESTFILE)

    #trim whitespace
    expected_size=$(xargs <<< $expected_size)

    export TOKEN=$(gcloud auth print-access-token)
    
    #NOTE DO NOT TRY A PIPE USING CURL. `curl...| grep` does not work. YOU WILL HAVE BEEN WARNED.  
    run curl -I https://storage.googleapis.com/$BUCKET/$TESTFILE \
            -H "Authorization: Bearer $(gcloud auth print-access-token)" \
            --cacert $CA_BUNDLE \
            --proxy $HTTPS_PROXY 2>/dev/null 
    echo "expected_size: $expected_size"
    assert_output --partial "X-Goog-Meta-X-Unencrypted-Content-Length: $expected_size"     
}

@test "Test GCS Metadata - verify md5 hash" {      
    
    #calcuate base64 md5
    local expected_md5=$(openssl base64 -in <(openssl dgst -md5 -binary $TESTFILE))
    
    #trim whitespace
    expected_md5=$(xargs <<< $expected_md5)
    
    #NOTE DO NOT TRY A PIPE USING CURL. `curl...| grep` does not work. YOU WILL HAVE BEEN WARNED.  
    run curl -s -I https://storage.googleapis.com/$BUCKET/$TESTFILE  \
            -H "Authorization: Bearer $(gcloud auth print-access-token)" \
            --cacert $CA_BUNDLE \
            --proxy $HTTPS_PROXY

    
    assert_output --partial "X-Goog-Meta-X-Md5hash: $expected_md5"     
}

@test "Teardown - gcloud storage rm" {
  run gcloud storage rm gs://$BUCKET/$TESTFILE
  assert_success
}
