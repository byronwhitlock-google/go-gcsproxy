load '../helpers/bats-support/load'
load '../helpers/bats-assert/load'

setup() {
  export TESTFILE="byte-range-requests.txt"
  # Create a temporary file with some content
  echo "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ" > $TESTFILE
}

teardown() {
  # Remove the temporary file
  rm $TESTFILE
}

@test "Setup - gcloud storage cp" {
  run gcloud storage cp $TESTFILE gs://$BUCKET/$TESTFILE 
  assert_success
}
# Helper function to download a byte range using curl
download_range() {
  curl -s https://storage.googleapis.com/$BUCKET/$TESTFILE  \
        -H "Range: bytes=$1-$2" \
        -H "Authorization: Bearer $(gcloud auth print-access-token)" \
        --cacert $CA_BUNDLE \
        --proxy $HTTPS_PROXY
}

@test "GCS byte range: download first 10 bytes" {
  
  run download_range 0 10
  assert_success
  # Assuming your object contains predictable content, like "0123456789ABCDEF..."
  assert_output "0123456789"
}

@test "GCS byte range: download middle range" {
  skip

  run download_range 5 14
  assert_success
  # Assuming your object contains predictable content
  assert_output "56789ABCDE"
}

@test "GCS byte range: download last 10 bytes" {
  skip

  object_size=$(get_object_size)
  start_byte=$((object_size - 10))
  end_byte=$((object_size - 1))
  run download_range $start_byte $end_byte
  assert_success
  # Assuming your object contains predictable content
  # and that the object is at least 10 bytes long.
  # If you do not know the content, you cannot assert the output.
  # You could however, assert the length of the returned content.
  run echo "$output" | wc -c
  assert_output "10"
}

@test "GCS byte range: download single byte" {
  skip

  run download_range 5 5
  assert_success
  # Assuming your object contains predictable content
  assert_output "5"
}

@test "GCS byte range: download from start to end (full object)" {
  local object_size=$(wc -c < $TESTFILE)
  object_size=$(xargs <<< $object_size)
  
  run download_range 0 $((object_size - 1))
  assert_success
  assert_output $(cat $TESTFILE)
}

@test "GCS byte range: download from specific byte to end" {
  skip

  object_size=$(wc -c $TESTFILE)
  run download_range 5 "" #Download from byte 5 to the end.
  assert_success
  run gsutil cat gs://$BUCKET/$TESTFILE | tail -c $((object_size - 5)) > expected_content
  run echo "$output" > actual_content
  run diff expected_content actual_content
  assert_success
  rm expected_content actual_content
}

@test "GCS byte range: invalid range (start > end)" {
  run download_range 10 5
  assert_failure
}

@test "GCS byte range: invalid range (negative start)" {
  run download_range -1 5
  assert_failure
}

@test "GCS byte range: invalid range (negative end)" {
  run download_range 0 -1
  assert_failure
}

@test "GCS byte range: range beyond object size" {
  local object_size=$(wc -c < $TESTFILE)
  object_size=$(xargs <<< $object_size)

  run download_range 0 $((object_size + 1000))
  assert_success
  run gsutil cat gs://$BUCKET/$TESTFILE > expected_content
  run echo "$output" > actual_content
  run diff expected_content actual_content
  assert_success
  rm expected_content actual_content
}

@test "Teardown - gcloud storage rm" {
  run gcloud storage rm gs://$BUCKET/$TESTFILE
  assert_success
}
