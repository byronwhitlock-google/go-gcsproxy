load 'helpers/bats-support/load'
load 'helpers/bats-assert/load'

setup() {
    export TESTFILE="file.bin"
    export CONTENT="This is a test File"
    echo $CONTENT > $TESTFILE
}

teardown() {
    rm $TESTFILE
}

@test "Test gsutil cp" {
  run gsutil cp $TESTFILE gs://$BUCKET/$TESTFILE
  assert_success
}

@test "Test gsutil cat - returns exact object" {
  local expected_output=$(cat $TESTFILE)

  run gsutil cat gs://$BUCKET/$TESTFILE
  
  assert_success
  assert_output "$expected_output"
}

@test "Test gsutil cat - returns correct byte length" {
  # Get the expected byte length (replace with your actual logic)
  local expected_length=$(wc -c $TESTFILE)

  # Run the gcloud command and capture its output
  run gsutil cat gs://$BUCKET/$TESTFILE 
  assert_success

  # Calculate the actual length of the output
  local actual_length=$(echo "$output" | wc -c)

  assert_equal $actual_length $expected_length

  # # Check if the lengths match
  # if [ "$actual_length" -ne "$expected_length" ]; then
  #   printf "Expected length: %d, got %d\n" "$expected_length" "$actual_length"
  #   assert_failure #Incorrect byte length
  # fi
}

@test "Test gsutil rm" {
  run gsutil rm gs://$BUCKET/$TESTFILE
  assert_success
}


# @test "Test gsutil unit tests" {
#   run gsutil test
#   assert_success
# }

